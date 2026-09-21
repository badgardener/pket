package core

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

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

	if err := os.MkdirAll(temp_path, 0755); err != nil {
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
	base_include := filepath.Join(base_path, config.Files.Base)
	call.Log("Reading base include directory: " + base_include)
	entries, err := os.ReadDir(base_include)

	if err != nil {
		call.Error("Cannot read the base include dir: " + err.Error())
		return
	}

	call.Log(fmt.Sprintf("Found %d base entries.", len(entries)))
	files_to_pack := []string{}
	assets_to_pack := []string{}

	for _, f := range entries {
		target := filepath.Join(base_path, config.Files.Base, f.Name())
		files_to_pack = append(files_to_pack, target)
	}

	call.Log(fmt.Sprintf("Processing %d asset(s)...", len(config.Files.Assets)))

	for _, f := range config.Files.Assets {
		target := filepath.Join(base_path, f)
		assets_to_pack = append(assets_to_pack, target)
	}

	call.Info("Write SHA-512 sums...")
	shafile := []string{filepath.Join(temp_path, "sha-512.sums")}
	call.Log("Creating SHA-512 output file...")
	hash_file, err := os.Create(shafile[0])

	if err != nil {
		call.Warn("Cannot create hash file: " + err.Error())
		shafile = []string{}

		if !call.Prompt("Continue?", true) {
			return
		}
	} else {
		hash_file.Close()
		call.Log("SHA-512 output file created.")

		call.Log("Calculating SHA-512 hash for payload...")

		var hash_entries []HashEntry

		for _, target := range files_to_pack {
			target_name := filepath.Base(filepath.Clean(target))

			err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.Mode().IsRegular() {
					return nil
				}

				relative, err := filepath.Rel(target, path)
				if err != nil {
					return err
				}

				archive_path := filepath.Join(target_name, relative)

				if relative == "." {
					archive_path = target_name
				}

				hash_entries = append(hash_entries, HashEntry{
					source: path,
					target: filepath.ToSlash(archive_path),
				})

				return nil
			})

			if err != nil {
				call.Warn("Cannot prepare hash input: " + err.Error())
				if !call.Prompt("Continue?", true) {
					return
				}
				break
			}
		}

		for _, target := range assets_to_pack {
			target_name := filepath.Base(filepath.Clean(target))

			err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if !info.Mode().IsRegular() {
					return nil
				}

				relative, err := filepath.Rel(target, path)
				if err != nil {
					return err
				}

				archive_path := filepath.Join("pket-assets", target_name, relative)

				if relative == "." {
					archive_path = filepath.Join("pket-assets", target_name)
				}

				hash_entries = append(hash_entries, HashEntry{
					source: path,
					target: filepath.ToSlash(archive_path),
				})

				return nil
			})

			if err != nil {
				call.Warn("Cannot prepare asset hash input: " + err.Error())
				if !call.Prompt("Continue?", true) {
					return
				}
				break
			}
		}

		hash, err := Sha512Files(hash_entries, call)

		if err != nil {
			call.Warn("Cannot obtain hash: " + err.Error())

			if !call.Prompt("Continue?", true) {
				return
			}
		} else {
			call.Log("SHA-512 calculation completed.")
			err = os.WriteFile(shafile[0], []byte(hash+"\n"), 0644)
			if err != nil {
				call.Warn("Cannot write hash: " + err.Error())

				if !call.Prompt("Continue?", true) {
					return
				}
			} else {
				call.Success("Hash written successfully.")
			}
		}
	}

	call.Info("Building .pkt...")
	output_packet := filepath.Join(temp_path, config.Package.Pack+"-"+config.Package.Version+".pkt")
	call.Log("Creating package archive: " + output_packet)

	if err := make_tar(files_to_pack, assets_to_pack, append(shafile, manifest_path), output_packet, call); err != nil {
		call.Error("Cannot make packet: " + err.Error())
		return
	}

	call.Log("Package archive created.")
	call.Success("Package built successfully.")
}

func make_tar(target_files_or_folders_base []string,
	target_files_or_folders_assets []string, manifest_files []string,
	output_file string, call Callback) error {
	call.Log("Creating archive output directory...")

	if err := os.MkdirAll(filepath.Dir(output_file), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	call.Log("Creating archive file: " + output_file)
	file, err := os.Create(output_file)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
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

	add_targets := func(targets []string, archive_prefix string) error {
		for _, target := range targets {
			target_path, err := filepath.Abs(target)
			if err != nil {
				return fmt.Errorf("failed to resolve %s: %w", target, err)
			}

			call.Log("Reading archive target: " + target_path)

			_, err = os.Stat(target_path)
			if err != nil {
				return fmt.Errorf("failed to stat %s: %w", target_path, err)
			}

			target_name := filepath.Base(filepath.Clean(target_path))

			err = filepath.Walk(target_path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relative, err := filepath.Rel(target_path, path)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				if relative == "." {
					relative = target_name
				} else {
					relative = filepath.Join(target_name, relative)
				}

				if archive_prefix != "" {
					relative = filepath.Join(archive_prefix, relative)
				}

				relative = filepath.ToSlash(relative)

				call.Log("Adding '" + relative + "'...")

				header, err := tar.FileInfoHeader(info, "")
				if err != nil {
					return fmt.Errorf("failed to create tar header for %s: %w", relative, err)
				}

				header.Name = relative

				if err := tar_writer.WriteHeader(header); err != nil {
					return fmt.Errorf("failed to write tar header for %s: %w", relative, err)
				}

				if info.Mode().IsRegular() {
					call.Log("Opening archive source file: " + path)

					src, err := os.Open(path)
					if err != nil {
						return fmt.Errorf("failed to open %s: %w", path, err)
					}

					_, copy_err := io.Copy(tar_writer, src)
					close_err := src.Close()

					if copy_err != nil {
						return fmt.Errorf("failed to write %s: %w", relative, copy_err)
					}

					if close_err != nil {
						return fmt.Errorf("failed to close %s: %w", path, close_err)
					}
				}

				call.Log("Finished archive entry: " + relative)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	}

	call.Log("Adding base files and folders...")
	if err := add_targets(target_files_or_folders_base, ""); err != nil {
		return err
	}

	call.Log("Adding asset files and folders...")
	if err := add_targets(target_files_or_folders_assets, "pket-assets"); err != nil {
		return err
	}

	call.Log("Adding manifest files...")
	if err := add_targets(manifest_files, "pket-manifest"); err != nil {
		return err
	}

	call.Log("Finished creating archive.")
	return nil
}

type HashEntry struct {
	source string
	target string
}

func Sha512Files(entries []HashEntry, call Callback) (string, error) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].target < entries[j].target
	})

	type result struct {
		index int
		data  []byte
		err   error
	}

	workers := runtime.NumCPU() - 1
	jobs := make(chan int, workers)
	results := make(chan result, len(entries))
	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			for i := range jobs {
				entry := entries[i]
				call.Log("Hashing path: " + entry.target)
				hash := sha512.New()

				if _, err := hash.Write([]byte(entry.target)); err != nil {
					results <- result{index: i, err: err}
					continue
				}

				if _, err := hash.Write([]byte{0}); err != nil {
					results <- result{index: i, err: err}
					continue
				}

				f, err := os.Open(entry.source)

				if err != nil {
					results <- result{index: i, err: err}
					continue
				}

				_, copyErr := io.Copy(hash, f)
				closeErr := f.Close()

				if copyErr != nil {
					results <- result{index: i, err: fmt.Errorf("failed to hash %s: %w", entry.source, copyErr)}
					continue
				}

				if closeErr != nil {
					results <- result{index: i, err: fmt.Errorf("failed to close %s: %w", entry.source, closeErr)}
					continue
				}

				if _, err := hash.Write([]byte{0}); err != nil {
					results <- result{index: i, err: err}
					continue
				}

				results <- result{index: i, data: []byte(fmt.Sprintf("%x", hash.Sum(nil)))}
			}
		})
	}

	go func() {
		defer close(jobs)
		for i := range entries {
			jobs <- i
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	hash := sha512.New()
	outputs := make([][]byte, len(entries))

	for result := range results {
		if result.err != nil {
			return "", result.err
		}

		outputs[result.index] = result.data
	}

	for _, output := range outputs {
		hash.Write(output)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
