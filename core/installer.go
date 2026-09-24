package core

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/klauspost/pgzip"
)

func Install(pack string, callback Callback) {
	callback.Log("Trying to access home dir...")
	home_dir, err := os.UserHomeDir()

	if err != nil {
		callback.Error("Home directory cannot be defined.")
		return
	}

	app_home_dir := filepath.Join(home_dir, Package_data_suffix())
	callback.Log("Detected app home dir: " + app_home_dir)

	callback.Log("Checking home dir...")
	file, err := os.Stat(app_home_dir)

	if err == nil {
		if !file.IsDir() {
			callback.Log("Removing FILE in place of app home dir...")
			err = os.Remove(app_home_dir)

			if err != nil {
				callback.Error("Cannot remove FILE in place of app home dir.")
				return
			}

			callback.Log("Removed FILE in place of app home dir.")
			callback.Log("Creating app home dir...")
			err = os.MkdirAll(app_home_dir, 0755)

			if err != nil {
				callback.Error("Cannot create: " + app_home_dir)
				return
			}

			callback.Log("Home DIR ready.")
		} else {
			callback.Log("Home dir exists.")
		}
	} else {
		callback.Log("Creating app home dir...")
		err = os.MkdirAll(app_home_dir, 0755)

		if err != nil {
			callback.Error("Cannot create: " + app_home_dir)
			return
		}

		callback.Log("Home DIR ready.")
	}

	callback.Log("Checking for target file...")
	file, err = os.Stat(pack)

	if err != nil {
		callback.Error("Target file does not exist.")
		return
	}

	callback.Log("Target temp stat completed.")
	if file.IsDir() {
		callback.Error("Given temp is not a file.")
		return
	}

	callback.Log("Target file exists.")
	callback.Log("Making temporary dir...")
	temp, err := os.MkdirTemp(app_home_dir, ".pket-install-")

	if err != nil {
		callback.Error("Cannot create temporary directory.")
		return
	}

	callback.Log("Created temporary dir.")
	install_pack(pack, callback, temp, app_home_dir)

	callback.Log("Removing temporary dir.")
	err = os.RemoveAll(temp)

	if err != nil {
		callback.Warn("Cannot remove: " + temp + ". Please manually delete this.")
	} else {
		callback.Success("Removed temporary files.")
	}
}

func extract_tar(pack string, dir string, call Callback) error {
	file, err := os.Open(pack)
	if err != nil {
		return fmt.Errorf("cannot open package: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat package: %w", err)
	}

	call.Log(fmt.Sprintf("Opening package archive: %s (%d bytes).", pack, stat.Size()))

	workers := max(runtime.NumCPU()-1, 1)
	gzipReader, err := pgzip.NewReaderN(file, 1<<20, workers)
	if err != nil {
		return fmt.Errorf("cannot open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	reader := tar.NewReader(gzipReader)
	count := 0
	seen := make(map[string]struct{})

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot read archive: %w", err)
		}

		name := filepath.Clean(header.Name)
		if name == "." || filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("invalid archive temp: %s", header.Name)
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate archive entry: %s", header.Name)
		}
		seen[name] = struct{}{}

		target := filepath.Join(dir, name)
		base := filepath.Clean(dir)

		relative, err := filepath.Rel(base, target)
		if err != nil {
			return fmt.Errorf("cannot validate archive temp %q: %w", header.Name, err)
		}

		if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("archive temp escapes extraction directory: %s", header.Name)
		}
		if err := validateExtractionParents(base, target); err != nil {
			return err
		}
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("duplicate or conflicting archive entry: %s", header.Name)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("cannot inspect archive target %s: %w", name, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("cannot create directory %s: %w", name, err)
			}

		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("cannot create parent directory for %s: %w", name, err)
			}

			output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("cannot create file %s: %w", name, err)
			}

			_, copyErr := io.Copy(output, reader)
			closeErr := output.Close()

			if copyErr != nil {
				return fmt.Errorf("cannot extract file %s: %w", name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("cannot close extracted file %s: %w", name, closeErr)
			}

		case tar.TypeSymlink:
			linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
			linkRelative, err := filepath.Rel(base, filepath.Clean(linkTarget))
			if err != nil || linkRelative == ".." || strings.HasPrefix(linkRelative, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("invalid symbolic link target in archive: %s", header.Name)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("cannot create parent directory for %s: %w", name, err)
			}

			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("cannot create symbolic link %s: %w", name, err)
			}

		case tar.TypeLink:
			linkTarget := filepath.Join(dir, filepath.Clean(header.Linkname))
			relative, err := filepath.Rel(base, linkTarget)
			if err != nil {
				return fmt.Errorf("cannot validate hard link %s: %w", name, err)
			}

			if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("hard link escapes extraction directory: %s", name)
			}

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("cannot create parent directory for %s: %w", name, err)
			}

			if err := os.Link(linkTarget, target); err != nil {
				return fmt.Errorf("cannot create hard link %s: %w", name, err)
			}

		default:
			return fmt.Errorf("unsupported archive entry type for %s", name)
		}

		count++
		if count%1000 == 0 {
			call.Log(fmt.Sprintf("Extracted %d archive entries...", count))
		}
	}

	call.Success(fmt.Sprintf("Extracted %d archive entries.", count))
	return nil
}

func validateExtractionParents(base string, target string) error {
	relative, err := filepath.Rel(base, filepath.Dir(target))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("archive entry parent escapes extraction directory: %s", target)
	}

	current := base
	for _, part := range strings.Split(relative, string(os.PathSeparator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("cannot inspect archive parent %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("archive entry parent is not a directory: %s", current)
		}
	}
	return nil
}

func install_pack(pack string, call Callback, temp string, app_home_dir string) {
	call.Log("Starting package installation...")

	call.Info("Extracting packet...")
	err := extract_tar(pack, temp, call)

	if err != nil {
		call.Error("Cannot extract packet: " + err.Error())
		return
	}
	call.Log("Package extraction completed.")

	manifest_file := filepath.Join(temp, "pket-manifest", "pket-config.toml")
	shafile := filepath.Join(temp, "pket-manifest", "sha-512.sums")

	if _, err = os.Stat(manifest_file); err != nil {
		call.Error("No manifest file found in the package.")
		return
	}
	call.Log("Manifest file found: " + manifest_file)

	var sums_to_verify bool
	_, err = os.Stat(shafile)
	sums_to_verify = err == nil
	call.Log(fmt.Sprintf("SHA-512 verification available: %t.", sums_to_verify))

	if sums_to_verify {
		result, err := os.ReadFile(shafile)

		if err != nil {
			call.Error("Cannot read SHA file.")
			return
		}

		hash_entries, err := make_install_hash_entries(filepath.Join(temp, "payload"))

		if err != nil {
			call.Error("Cannot prepare SHA inputs: " + err.Error())
			return
		}

		sha := string(result)
		real, err := Sha512Files(hash_entries, call)

		if err != nil {
			call.Error("Cannot extract SHA from input: " + err.Error())
			return
		}

		real += "\n"

		if sha != real {
			call.Error("Expected SHA do not match to real one.\n  Expected: " +
				strings.Trim(sha, "\n") + "\n  Real:     " + strings.Trim(real, "\n"))
			return
		}

		call.Success("SHA sums matched.")
	} else if !call.Prompt("SHA sums not found in this package.\nDo you still want to temporary this package?", false) {
		return
	}

	call.Log("Reading manifest file...")
	b, err := os.ReadFile(manifest_file)

	if err != nil {
		call.Error("Cannot read manifest file: " + err.Error())
		return
	}

	var config BuiltConfig
	_, err = toml.Decode(string(b), &config)

	if err != nil {
		call.Error("Cannot parse pket-config.toml: " + err.Error())
		return
	}

	call.Log("Manifest decoded successfully.")

	if err := config.Validate(); err != nil {
		call.Error("Invalid pket-config.toml: " + err.Error())
		return
	}

	call.Log("Manifest validation completed.")
	call.Success("Parsed pket-config.toml.")

	uid := strings.TrimSpace(config.Metadata.PackageUID)
	if uid == "" || uid == "." || uid == ".." || filepath.Base(uid) != uid || strings.ContainsAny(uid, `/\\`) {
		call.Error("Package UID is invalid.")
		return
	}

	package_name := strings.TrimSpace(config.Metadata.PackageName)
	if package_name == "" || package_name == "." || package_name == ".." || strings.ContainsAny(package_name, `/\\`) {
		call.Error("Package name is invalid.")
		return
	}

	install_dir := filepath.Join(app_home_dir, uid)
	call.Log("Package UID: " + uid)
	call.Log("Package name: " + package_name)
	call.Log("Installation directory: " + install_dir)

	_, err = os.Stat(install_dir)
	installed := err == nil
	if err != nil && !os.IsNotExist(err) {
		call.Error("Cannot inspect existing installation: " + err.Error())
		return
	}

	if installed {
		call.Warn("Package UID is already installed; updating existing package.")
		if !call.Prompt("Package is already installed. Do you want to update it?", false) {
			call.Log("Package update declined.")
			return
		}
	} else {
		call.Log("Package UID is not installed; creating new installation.")
	}

	staging_dir := filepath.Join(app_home_dir, "."+uid+".installing")
	call.Log("Preparing installation staging directory: " + staging_dir)
	if err := os.RemoveAll(staging_dir); err != nil {
		call.Error("Cannot clear installation staging directory: " + err.Error())
		return
	}
	staged_package_file := filepath.Join(temp, package_name+".pkt")
	if err := copy_file(pack, staged_package_file, call); err != nil {
		call.Error("Cannot save package archive: " + err.Error())
		os.RemoveAll(temp)
		return
	}

	call.Log("Applying executable permissions...")
	if err := make_executables_executable(config.Executables, temp, call); err != nil {
		call.Error("Cannot apply executable permissions: " + err.Error())
		os.RemoveAll(temp)
		return
	}

	if config.Install.PreInstall != "" && confirm_install_command("preinstall", config.Install.PreInstall, call) {
		if err := run_install_command(config.Install.PreInstall, temp, call); err != nil {
			call.Error("Preinstall command failed: " + err.Error())
			os.RemoveAll(temp)
			return
		}
	}

	backup_dir := ""
	if installed {
		backup_dir = filepath.Join(app_home_dir, "."+uid+".previous")
		call.Log("Moving existing installation to backup: " + backup_dir)
		if err := os.RemoveAll(backup_dir); err != nil {
			call.Error("Cannot clear update backup directory: " + err.Error())
			os.RemoveAll(temp)
			return
		}
		if err := os.Rename(install_dir, backup_dir); err != nil {
			call.Error("Cannot prepare package update: " + err.Error())
			os.RemoveAll(temp)
			return
		}
	}

	call.Log("Preparing verified package for activation...")
	if err := os.Rename(temp, staging_dir); err != nil {
		call.Error("Cannot prepare installation staging directory: " + err.Error())
		if installed {
			if restoreErr := os.Rename(backup_dir, install_dir); restoreErr != nil {
				call.Error("Cannot restore previous installation: " + restoreErr.Error())
			}
		}
		return
	}

	call.Log("Activating installation directory...")
	if err := os.Rename(staging_dir, install_dir); err != nil {
		call.Error("Cannot activate installation: " + err.Error())
		if installed {
			if restoreErr := os.Rename(backup_dir, install_dir); restoreErr != nil {
				call.Error("Cannot restore previous installation: " + restoreErr.Error())
			}
		}
		os.RemoveAll(staging_dir)
		return
	}

	if installed {
		call.Log("Removing externally created files from previous installation...")
		if err := remove_external_files(backup_dir, call, install_dir); err != nil {
			call.Error("Cannot remove previous external files: " + err.Error())
			return
		}
	}

	call.Log("Creating executable links...")
	external_files, err := create_executable_links(config.Executables, install_dir, app_home_dir, call)
	if err != nil {
		call.Error("Cannot create executable links: " + err.Error())
		return
	}

	files_list := filepath.Join(install_dir, "pket-manifest", "files.lst")
	if err := os.WriteFile(files_list, []byte(strings.Join(external_files, "\n")+"\n"), 0644); err != nil {
		call.Error("Cannot write files.lst: " + err.Error())
		return
	}
	call.Log("Wrote externally created files list: " + files_list)

	if installed {
		if err := os.RemoveAll(backup_dir); err != nil {
			call.Warn("Cannot remove previous installation backup: " + err.Error())
		}
		call.Success("Package updated successfully.")
	} else {
		call.Success("Package installed successfully.")
	}

	if config.Install.PostInstall != "" {
		if confirm_install_command("postinstall", config.Install.PostInstall, call) {
			if err := run_install_command(config.Install.PostInstall, install_dir, call); err != nil {
				call.Error("Postinstall command failed: " + err.Error())
				return
			}
		}
	}
}

func make_executables_executable(executables []ExecutableConfig, install_dir string, call Callback) error {
	for _, executable := range executables {
		source, err := resolve_resource_path(executable.Path, install_dir)
		if err != nil {
			return err
		}
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("cannot inspect executable %s: %w", source, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("executable source is not a regular file: %s", source)
		}

		mode := info.Mode().Perm() | 0111
		if err := os.Chmod(source, mode); err != nil {
			return fmt.Errorf("cannot make executable %s executable: %w", source, err)
		}
		call.Log("Set executable permissions: " + source)
	}

	return nil
}

func create_executable_links(executables []ExecutableConfig, install_dir string, app_home_dir string, call Callback) ([]string, error) {
	link_dir, err := executable_link_dir(app_home_dir)
	if err != nil {
		return nil, err
	}
	call.Log("Executable link directory: " + link_dir)

	created := make([]string, 0, len(executables))
	cleanup := func() {
		for _, created_path := range created {
			if removeErr := os.Remove(created_path); removeErr != nil && !os.IsNotExist(removeErr) {
				call.Warn("Cannot remove partial executable link: " + removeErr.Error())
			}
		}
	}

	for _, executable := range executables {
		link_name := strings.TrimSpace(executable.Link)
		if link_name == "" || link_name == "." || link_name == ".." || filepath.Base(link_name) != link_name || strings.ContainsAny(link_name, `/\\`) {
			cleanup()
			return nil, fmt.Errorf("invalid executable link name %q", executable.Link)
		}

		source, err := resolve_resource_path(executable.Path, install_dir)
		if err != nil {
			cleanup()
			return nil, err
		}
		if _, err := os.Lstat(source); err != nil {
			cleanup()
			return nil, fmt.Errorf("executable source %s is unavailable: %w", source, err)
		}

		link_path := filepath.Join(link_dir, link_name)
		if _, err := os.Lstat(link_path); err == nil {
			cleanup()
			return nil, fmt.Errorf("executable link already exists: %s", link_path)
		} else if !os.IsNotExist(err) {
			cleanup()
			return nil, fmt.Errorf("cannot inspect executable link %s: %w", link_path, err)
		}

		if err := os.Symlink(source, link_path); err != nil {
			cleanup()
			return nil, fmt.Errorf("cannot create executable link %s: %w", link_path, err)
		}
		created = append(created, link_path)
		call.Log("Created executable link: " + link_path + " -> " + source)
	}

	return created, nil
}

func resolve_resource_path(resource string, install_dir string) (string, error) {
	resource = strings.TrimSpace(resource)
	if !strings.HasPrefix(resource, "@res/") {
		return "", fmt.Errorf("resource path must start with @res/: %s", resource)
	}
	relative := strings.TrimPrefix(resource, "@res/")
	if !isSafeRelativePath(relative) {
		return "", fmt.Errorf("resource path escapes package payload: %s", resource)
	}
	return filepath.Join(install_dir, "payload", filepath.FromSlash(relative)), nil
}

func executable_link_dir(app_home_dir string) (string, error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	candidates := []string{}
	switch runtime.GOOS {
	case "linux":
		candidates = append(candidates, filepath.Join(home_dir, ".local", "bin"))
	case "darwin":
		candidates = append(candidates, "/usr/local/bin", filepath.Join(home_dir, "bin"))
	case "windows":
		candidates = strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	default:
		candidates = append(candidates, filepath.Join(app_home_dir, "bin"))
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if err := os.MkdirAll(candidate, 0755); err != nil {
			continue
		}
		test, err := os.CreateTemp(candidate, ".pket-write-*")
		if err != nil {
			continue
		}
		test_name := test.Name()
		if err := test.Close(); err != nil {
			os.Remove(test_name)
			continue
		}
		if err := os.Remove(test_name); err != nil {
			continue
		}
		return candidate, nil
	}

	return "", fmt.Errorf("no writable executable directory found in PATH")
}

func remove_external_files(install_dir string, call Callback, additional_roots ...string) error {
	files_list := filepath.Join(install_dir, "pket-manifest", "files.lst")
	data, err := os.ReadFile(files_list)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(data), "\n") {
		path := strings.TrimSpace(line)
		if path == "" {
			continue
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("refusing to remove non-absolute external file: %s", path)
		}
		info, statErr := os.Lstat(path)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("refusing to remove non-symlink external file: %s", path)
		}
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return err
		}
		resolvedTarget := linkTarget
		if !filepath.IsAbs(resolvedTarget) {
			resolvedTarget = filepath.Join(filepath.Dir(path), resolvedTarget)
		}
		owned := isPathWithin(install_dir, resolvedTarget)
		for _, root := range additional_roots {
			owned = owned || isPathWithin(root, resolvedTarget)
		}
		if !owned {
			return fmt.Errorf("refusing to remove external link not owned by package: %s", path)
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		call.Log("Removed external file: " + path)
	}
	return nil
}

func isPathWithin(base string, candidate string) bool {
	base, baseErr := filepath.Abs(base)
	candidate, candidateErr := filepath.Abs(candidate)
	if baseErr != nil || candidateErr != nil {
		return false
	}
	relative, err := filepath.Rel(base, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func confirm_install_command(name string, command string, call Callback) bool {
	call.Warn(name + " command is configured: " + command)
	if !call.Prompt("Run "+name+" command?", false) {
		call.Log(name + " command declined at first confirmation.")
		return false
	}

	if !call.Prompt("Confirm running "+name+" command?", false) {
		call.Log(name + " command declined at second confirmation.")
		return false
	}

	return true
}

func run_install_command(command string, install_dir string, call Callback) error {
	payload_dir := filepath.Join(install_dir, "payload")
	command = strings.ReplaceAll(command, "@res", payload_dir)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = install_dir

	call.Log("Running install command from " + install_dir + ": " + command)
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		for line := range strings.SplitSeq(strings.TrimRight(string(output), "\r\n"), "\n") {
			call.Info("Install command output: " + line)
		}
	}
	if err != nil {
		return err
	}

	call.Success("Install command completed successfully.")
	return nil
}

func copy_file(source string, target string, call Callback) error {
	if info, err := os.Stat(source); err == nil && info.Mode().IsRegular() {
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		if err := os.Link(source, target); err == nil {
			call.Log("Linked package archive: " + filepath.ToSlash(target))
			return nil
		}
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	info, err := input.Stat()
	if err != nil {
		return err
	}

	output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}

	if closeErr != nil {
		return closeErr
	}

	call.Log("Saved package archive: " + filepath.ToSlash(target))
	return nil
}

func (c BuiltConfig) Validate() error {
	if strings.TrimSpace(c.Metadata.PackageName) == "" {
		return fmt.Errorf("metadata.package_name is required")
	}

	if strings.TrimSpace(c.Metadata.PackageUID) == "" {
		return fmt.Errorf("metadata.package_uid is required")
	}

	for field, value := range map[string]string{
		"metadata.package_name": c.Metadata.PackageName,
		"metadata.package_uid":  c.Metadata.PackageUID,
	} {
		if value == "." || value == ".." || filepath.Base(value) != value || strings.ContainsAny(value, `/\\`) {
			return fmt.Errorf("%s must be a single path component", field)
		}
	}

	if len(c.Metadata.Version) != 3 {
		return fmt.Errorf("metadata.version must have exactly three parts")
	}

	for i, part := range c.Metadata.Version {
		if part < 0 {
			return fmt.Errorf("metadata.version[%d] must not be negative", i)
		}
	}

	if len(c.Executables) == 0 {
		return fmt.Errorf("at least one executable is required")
	}

	for i, executable := range c.Executables {
		if strings.TrimSpace(executable.Path) == "" {
			return fmt.Errorf("executable[%d].path is required", i)
		}

		if strings.TrimSpace(executable.Link) == "" {
			return fmt.Errorf("executable[%d].link is required", i)
		}
		if !strings.HasPrefix(strings.TrimSpace(executable.Path), "@res/") || !isSafeRelativePath(strings.TrimPrefix(strings.TrimSpace(executable.Path), "@res/")) {
			return fmt.Errorf("executable[%d].path must be inside the package payload", i)
		}
		link := strings.TrimSpace(executable.Link)
		if link == "." || link == ".." || filepath.Base(link) != link || strings.ContainsAny(link, `/\\`) {
			return fmt.Errorf("executable[%d].link must be a single path component", i)
		}
	}

	return nil
}

func make_install_hash_entries(temp string) ([]HashEntry, error) {
	var hash_entries []HashEntry

	err := filepath.Walk(temp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		relative, err := filepath.Rel(temp, path)
		if err != nil {
			return err
		}

		hash_entries = append(hash_entries, HashEntry{
			source: path,
			target: filepath.ToSlash(relative),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return hash_entries, nil
}
