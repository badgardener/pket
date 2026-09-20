package core

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type BuilderConfig struct {
	Package     BuilderPackageConfig `toml:"package"`
	Files       BuilderFilesConfig   `toml:"files"`
	Executables []ExecutableConfig   `toml:"executable"`
	Install     InstallConfig        `toml:"install"`
}

type BuilderPackageConfig struct {
	Name        string   `toml:"name"`
	Pack        string   `toml:"pack"`
	Version     string   `toml:"version"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
}

type BuilderFilesConfig struct {
	Base   string   `toml:"base"`
	Assets []string `toml:"assets"`
}

type ExecutableConfig struct {
	Path string `toml:"path"`
	Link string `toml:"link"`
}

type InstallConfig struct {
	PreInstall  string `toml:"preinstall"`
	PostInstall string `toml:"postinstall"`
}

func (c BuilderConfig) Validate() error {
	if c.Package.Name == "" {
		return fmt.Errorf("package.name is required")
	}

	if c.Package.Pack == "" {
		return fmt.Errorf("package.pack is required")
	}

	if c.Package.Version == "" {
		return fmt.Errorf("package.version is required")
	}

	if c.Files.Base == "" {
		return fmt.Errorf("files.base is required")
	}

	if c.Files.Assets == nil {
		return fmt.Errorf("files.assets is required")
	}

	if len(c.Executables) == 0 {
		return fmt.Errorf("at least one executable is required")
	}

	for i, executable := range c.Executables {
		if executable.Path == "" {
			return fmt.Errorf("executable[%d].path is required", i)
		}

		if executable.Link == "" {
			return fmt.Errorf("executable[%d].link is required", i)
		}
	}

	return nil
}

func Build(project_path string, callback Callback) {
	callback.Log("Checking for target directory...")
	folder, err := os.Stat(project_path)

	if err != nil {
		callback.Error("Target path does not exists.")
		return
	}

	callback.Log("Target path stat completed.")
	if !folder.IsDir() {
		callback.Error("Given path is not a directory.")
		return
	}

	callback.Log("Target directory exists.")
	callback.Log("Checking for pket-package.toml...")
	toml_file := filepath.Join(project_path, "pket-package.toml")
	file, err := os.Stat(toml_file)

	if err != nil {
		callback.Error("pket-package.toml does not exists.")
		return
	}

	callback.Log("Manifest stat completed.")

	if file.IsDir() {
		callback.Error("pket-package.toml is a directory.")
		return
	}

	callback.Log("pket-package.toml exists.")
	callback.Log("Reading pket-package.toml...")
	data, err := os.ReadFile(toml_file)

	if err != nil {
		callback.Error("Cannot read pket-package.toml.")
		return
	}

	callback.Log(fmt.Sprintf("Read %d bytes from pket-package.toml.", len(data)))
	callback.Info("Parsing pket-package.toml...")
	var config BuilderConfig
	_, err = toml.Decode(string(data), &config)

	if err != nil {
		callback.Error("Cannot parse pket-package.toml: " + err.Error())
		return
	}

	callback.Log("Manifest decoded successfully.")

	if err := config.Validate(); err != nil {
		callback.Error("Invalid pket-package.toml: " + err.Error())
		return
	}

	callback.Log("Manifest validation completed.")
	callback.Success("Parsed pket-package.toml.")
	build_packet(project_path, config, callback)
}

type BuiltConfig struct {
	Metadata    BuiltMetaConfig    `toml:"metadata"`
	Executables []ExecutableConfig `toml:"executable"`
	Install     InstallConfig      `toml:"install"`
}

type BuiltMetaConfig struct {
	PackageName string   `toml:"package_name"`
	PackageUID  string   `toml:"package_uid"`
	Version     []int    `toml:"version"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
}

func parse_version(v string) ([]int, error) {
	parts := strings.Split(v, ".")

	if len(parts) != 3 {
		return nil, fmt.Errorf("version must have exactly three parts")
	}

	result := make([]int, 3)

	for i, part := range parts {
		n, err := strconv.Atoi(part)

		if err != nil {
			return nil, fmt.Errorf("invalid version part %q", part)
		}

		result[i] = n
	}

	return result, nil
}

func build_packet(base_path string, config BuilderConfig, call Callback) {
	call.Info("Building final config...")
	final_config := BuiltConfig{}

	final_config.Metadata = BuiltMetaConfig{
		PackageName: config.Package.Name,
		PackageUID:  config.Package.Pack,
		Description: config.Package.Description,
		Authors:     config.Package.Authors,
	}

	call.Log("Verifying version...")
	version, err := parse_version(config.Package.Version)

	if err != nil {
		call.Error("Field package.version is invalid in pket-package.toml: " + err.Error())
		return
	}

	call.Log(fmt.Sprintf("Parsed version: %v.", version))
	final_config.Metadata.Version = version
	final_config.Install = config.Install
	call.Log(fmt.Sprintf("Processing %d executable(s)...", len(config.Executables)))

	for i := range config.Executables {
		call.Log(fmt.Sprintf("Processing executable %d: %s -> %s.", i+1, config.Executables[i].Path,
			config.Executables[i].Link))

		config.Executables[i].Path = "@res/" + strings.Trim(config.Executables[i].Path, "/")
		call.Log(fmt.Sprintf("Normalized executable path: %s.", config.Executables[i].Path))
	}

	final_config.Executables = config.Executables
	call.Success("Final config built successfully.")
	call.Log("Creating temp dir...")
	temp_path := filepath.Join(base_path, ".pket")
	payload_path := filepath.Join(temp_path, "payload")
	assets_path := filepath.Join(payload_path, "pket-assets")
	call.Log("Checking existing temp dir...")
	_, err = os.Stat(temp_path)

	if err == nil {
		call.Log("Removing old temp dir...")
		err = os.RemoveAll(temp_path)

		if err != nil {
			call.Error("Cannot remove old temp dir: " + err.Error())
			return
		}

		call.Log("Old temp dir removed.")
	}

	call.Log("Creating temporary directory structure...")

	if err := os.MkdirAll(assets_path, 0755); err != nil {
		call.Error("Cannot create temp directory structure: " + err.Error())
		return
	}

	call.Log("Temporary directory structure created.")
	call.Log("Creating final metadata file...")
	manifest_path := filepath.Join(temp_path, "pket-config.toml")
	file, err := os.Create(manifest_path)

	if err != nil {
		call.Error("Cannot create file in temp dir: " + err.Error())
		return
	}

	call.Log("Final metadata file created.")

	defer func() {
		call.Log("Closing final metadata file...")
		file.Close()
	}()

	encoder := toml.NewEncoder(file)
	call.Log("Encoding final metadata...")
	err = encoder.Encode(final_config)

	if err != nil {
		call.Error("Cannot write final manifest: " + err.Error())
		return
	}

	call.Log("Final metadata encoded.")
	call.Info("Copying files...")
	base_include := filepath.Join(base_path, config.Files.Base)
	call.Log("Reading base include directory: " + base_include)
	entries, err := os.ReadDir(base_include)

	if err != nil {
		call.Error("Cannot read the base include dir: " + err.Error())
		return
	}

	call.Log(fmt.Sprintf("Found %d base entries.", len(entries)))

	for _, f := range entries {
		target := filepath.Join(base_path, config.Files.Base, f.Name())
		output := filepath.Join(payload_path, f.Name())
		call.Log(fmt.Sprintf("Preparing base entry '%s'...", f.Name()))

		if f.IsDir() {
			call.Log("Entry is a directory.")
			err = copy_directory(target, output, call)
		} else {
			call.Log("Entry is a file.")
			err = copy_file(target, output, call)
		}

		if err != nil {
			call.Warn("Cannot copy '" + f.Name() + "': " + err.Error())
		} else {
			call.Log("Copied '" + f.Name() + "'.")
		}
	}

	call.Log(fmt.Sprintf("Processing %d asset(s)...", len(config.Files.Assets)))

	for _, f := range config.Files.Assets {
		target := filepath.Join(base_path, f)
		call.Log("Checking asset: " + target)
		file, err := os.Stat(target)

		if err != nil {
			call.Warn("Cannot copy '" + f + "': File/Folder does not exists.")
			continue
		}

		call.Log(fmt.Sprintf("Asset '%s' found.", f))
		output := filepath.Join(assets_path, file.Name())
		call.Log("Copying '" + f + "'...")

		if file.IsDir() {
			call.Log("Asset is a directory.")
			err = copy_directory(target, output, call)
		} else {
			call.Log("Asset is a file.")
			err = copy_file(target, output, call)
		}

		if err != nil {
			call.Warn("Cannot copy '" + f + "': " + err.Error())
		} else {
			call.Log("Copied '" + f + "'.")
		}
	}

	call.Success("Copying of files done.")
	call.Info("Write SHA-512 sums...")
	shafile := filepath.Join(temp_path, "sha-512.sums")
	call.Log("Creating SHA-512 output file...")
	hash_file, err := os.Create(shafile)

	if err != nil {
		call.Warn("Cannot create hash file: " + err.Error())
	} else {
		hash_file.Close()
		call.Log("SHA-512 output file created.")
	}

	call.Log("Calculating SHA-512 hash for payload...")
	hash, err := sha512_folder(payload_path, call)

	if err != nil {
		call.Warn("Cannot obtain hash: " + err.Error())
	} else {
		call.Log("SHA-512 calculation completed.")
		err = os.WriteFile(shafile, []byte(hash+"\n"), 0644)
		if err != nil {
			call.Warn("Cannot write hash: " + err.Error())
		} else {
			call.Success("Hash written successfully.")
		}
	}

	call.Info("Building .pkt...")
	output_packet := filepath.Join(base_path, config.Package.Pack+"-"+config.Package.Version+".pkt")
	call.Log("Creating package archive: " + output_packet)

	if err := make_tar(temp_path, output_packet, call); err != nil {
		call.Error("Cannot make packet: " + err.Error())
		return
	}

	call.Log("Package archive created.")
	call.Success("Package built successfully.")
}

func copy_file(target string, output string, call Callback) error {
	call.Log("Opening source file: " + target)
	src, err := os.Open(target)

	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}

	call.Log("Source file opened.")

	defer func() {
		call.Log("Closing source file: " + target)
		src.Close()
	}()

	call.Log("Reading source file metadata...")
	srcInfo, err := src.Stat()

	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	call.Log(fmt.Sprintf("Source file size: %d bytes.", srcInfo.Size()))
	call.Log("Creating destination directory: " + filepath.Dir(output))

	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	call.Log("Destination directory ready.")
	call.Log("Creating destination file: " + output)
	dst, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())

	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	call.Log("Destination file created.")

	defer func() {
		call.Log("Closing destination file: " + output)
		dst.Close()
	}()

	call.Log("Copying file data...")

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	call.Log("File data copied.")
	return nil
}

func copy_directory(target string, output string, call Callback) error {
	call.Log("Reading source directory metadata: " + target)
	srcInfo, err := os.Stat(target)

	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("target %s is not a directory", target)
	}

	call.Log("Creating destination directory: " + output)

	if err := os.MkdirAll(output, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	call.Log("Destination directory created.")
	call.Log("Reading directory entries: " + target)
	entries, err := os.ReadDir(target)

	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	call.Log(fmt.Sprintf("Found %d entries in '%s'.", len(entries), target))

	for _, entry := range entries {
		srcPath := filepath.Join(target, entry.Name())
		dstPath := filepath.Join(output, entry.Name())
		call.Log(fmt.Sprintf("Processing directory entry '%s'...", entry.Name()))

		if entry.IsDir() {
			call.Log("Entry is a directory; descending...")

			if err := copy_directory(srcPath, dstPath, call); err != nil {
				return err
			}
		} else {
			call.Log("Entry is a file; copying...")

			if err := copy_file(srcPath, dstPath, call); err != nil {
				return err
			}
		}
	}

	call.Log("Finished directory: " + target)
	return nil
}

func make_tar(target_dir string, output_file string, call Callback) error {
	call.Log("Reading archive target directory: " + target_dir)
	info, err := os.Stat(target_dir)

	if err != nil {
		return fmt.Errorf("failed to stat target directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory")
	}

	call.Log("Creating archive output directory...")

	if err := os.MkdirAll(filepath.Dir(output_file), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	call.Log("Creating archive file: " + output_file)
	file, err := os.Create(output_file)

	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	call.Log("Archive file created.")

	defer func() {
		call.Log("Closing archive file...")
		file.Close()
	}()

	call.Log("Creating gzip writer...")
	gzip_writer := gzip.NewWriter(file)

	defer func() {
		call.Log("Closing gzip writer...")
		gzip_writer.Close()
	}()

	call.Log("Creating tar writer...")
	tar_writer := tar.NewWriter(gzip_writer)

	defer func() {
		call.Log("Closing tar writer...")
		tar_writer.Close()
	}()

	call.Log("Walking archive source directory...")

	err = filepath.Walk(target_dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(target_dir, path)

		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		if relative == "." {
			call.Log("Skipping archive root directory.")
			return nil
		}

		relative = filepath.ToSlash(relative)
		call.Log("Adding '" + relative + "'...")
		call.Log("Creating tar header for '" + relative + "'...")
		header, err := tar.FileInfoHeader(info, "")

		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", relative, err)
		}

		header.Name = relative
		call.Log("Writing tar header for '" + relative + "'...")

		if err := tar_writer.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", relative, err)
		}

		if info.Mode().IsRegular() {
			call.Log("Opening archive source file: " + path)
			src, err := os.Open(path)

			if err != nil {
				return fmt.Errorf("failed to open %s: %w", path, err)
			}

			call.Log("Writing file contents to archive: " + relative)
			_, err = io.Copy(tar_writer, src)
			src.Close()
			call.Log("Closed archive source file: " + path)

			if err != nil {
				return fmt.Errorf("failed to write %s: %w", relative, err)
			}

			call.Log("File contents written: " + relative)
		}

		call.Log("Finished archive entry: " + relative)
		return nil
	})

	if err != nil {
		return err
	}

	call.Log("Finished walking archive source directory.")
	return nil
}

func sha512_folder(path string, call Callback) (string, error) {
	call.Log("Scanning files for SHA-512...")
	var files []string

	err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			call.Log("Found hash input: " + file)
			files = append(files, file)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	call.Log(fmt.Sprintf("Found %d files for SHA-512.", len(files)))
	call.Log("Sorting SHA-512 input files...")
	sort.Strings(files)
	call.Log("Creating SHA-512 hash state...")
	hash := sha512.New()

	for _, file := range files {
		relative, err := filepath.Rel(path, file)

		if err != nil {
			return "", err
		}

		relative = filepath.ToSlash(relative)
		call.Log("Hashing path: " + relative)
		hash.Write([]byte(relative))
		hash.Write([]byte{0})
		call.Log("Opening hash input: " + file)
		f, err := os.Open(file)

		if err != nil {
			return "", err
		}

		call.Log("Hashing file contents: " + relative)

		if _, err := io.Copy(hash, f); err != nil {
			f.Close()
			return "", err
		}

		f.Close()
		call.Log("Finished hashing: " + relative)
		hash.Write([]byte{0})
	}

	call.Log("Finalizing SHA-512 digest...")
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
