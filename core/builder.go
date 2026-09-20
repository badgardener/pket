package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type BuilderConfig struct {
	Package     BuilderPackageConfig `toml:"package"`
	Files       BuilderFilesConfig   `toml:"files"`
	Executables []ExecutableConfig   `toml:"executable"`
	Install     *InstallConfig       `toml:"install"`
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

	callback.Info("Parsing pket-package.toml...")
	var config BuilderConfig
	_, err = toml.Decode(string(data), &config)

	if err != nil {
		callback.Error("Cannot parse pket-package.toml: " + err.Error())
		return
	}

	if err := config.Validate(); err != nil {
		callback.Error("Invalid pket-package.toml: " + err.Error())
		return
	}

	callback.Success("Parsed pket-package.toml.")
	build_packet(project_path, config, callback)
}

type BuiltConfig struct {
	Metadata    BuiltMetaConfig    `toml:"metadata"`
	Executables []ExecutableConfig `toml:"executable"`
	Install     *InstallConfig     `toml:"install"`
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

	final_config.Metadata.Version = version
	final_config.Install = config.Install

	for _, exe := range config.Executables {
		exe.Path = "@res/" + strings.Trim(exe.Path, "/")
	}

	final_config.Executables = config.Executables
	call.Success("Final config built successfully.")

	call.Log("Creating temp dir...")

	temp_path := filepath.Join(base_path, ".pket")
	payload_path := filepath.Join(temp_path, "payload")

	folder, err := os.Stat(temp_path)

	if err == nil {
		call.Log("Removing old temp dir...")

		if folder.IsDir() {
			err = os.RemoveAll(temp_path)
		} else {
			err = os.Remove(temp_path)
		}

		if err != nil {
			call.Error("Cannot remove old temp dir: " + err.Error())
			return
		}

		return
	}

	os.MkdirAll(payload_path, 0755)

	call.Log("Creating final metadata file...")
	file, err := os.Create(filepath.Join(temp_path, "pket-config.toml"))

	if err != nil {
		call.Error("Cannot create file in temp dir: " + err.Error())
		return
	}

	defer file.Close()
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(final_config)

	if err != nil {
		call.Error("Cannot write final manifest: " + err.Error())
		return
	}
}
