package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

func Package_data_suffix() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join("AppData", "Roaming", "pket")
	case "darwin":
		return filepath.Join("Library", "Application Support", "pket")
	default:
		return filepath.Join(".local", "share", "pket")
	}
}

func Find_package(app_home_dir string, query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", fmt.Errorf("package name or UID is required")
	}

	entries, err := os.ReadDir(app_home_dir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no packages installed")
	}
	if err != nil {
		return "", fmt.Errorf("cannot read package directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		install_dir := filepath.Join(app_home_dir, entry.Name())
		if entry.Name() == query {
			return install_dir, nil
		}

		manifest_path := filepath.Join(install_dir, "pket-manifest", "pket-config.toml")
		data, err := os.ReadFile(manifest_path)
		if err != nil {
			continue
		}

		var config BuiltConfig
		if _, err := toml.Decode(string(data), &config); err == nil && config.Metadata.PackageName == query {
			return install_dir, nil
		}
	}

	return "", fmt.Errorf("package not found: %s", query)
}

func Uninstall(pack string, callback Callback) {
	callback.Log("Starting package removal...")

	home_dir, err := os.UserHomeDir()
	if err != nil {
		callback.Error("Home directory cannot be defined: " + err.Error())
		return
	}

	app_home_dir := filepath.Join(home_dir, Package_data_suffix())
	callback.Log("Searching installed packages in: " + app_home_dir)
	install_dir, err := Find_package(app_home_dir, pack)
	if err != nil {
		callback.Error(err.Error())
		return
	}

	manifest_path := filepath.Join(install_dir, "pket-manifest", "pket-config.toml")
	manifest_data, err := os.ReadFile(manifest_path)
	if err != nil {
		callback.Error("Cannot read package manifest: " + err.Error())
		return
	}

	var config BuiltConfig
	if _, err := toml.Decode(string(manifest_data), &config); err != nil {
		callback.Error("Cannot parse package manifest: " + err.Error())
		return
	}
	if err := config.Validate(); err != nil {
		callback.Error("Invalid package manifest: " + err.Error())
		return
	}

	uid := config.Metadata.PackageUID
	callback.Log("Resolved package: " + config.Metadata.PackageName + " (" + uid + ")")
	callback.Log("Installation directory: " + install_dir)

	if !callback.Prompt("Remove external files created by this package?", false) {
		callback.Log("External file removal declined.")
		return
	}

	if !callback.Prompt("Permanently uninstall package "+config.Metadata.PackageName+"?", false) {
		callback.Log("Package removal declined.")
		return
	}

	callback.Log("Removing externally created files...")
	if err := remove_external_files(install_dir, callback); err != nil {
		callback.Error("Cannot remove external files: " + err.Error())
		return
	}

	callback.Log("Removing package directory...")
	if err := os.RemoveAll(install_dir); err != nil {
		callback.Error("Cannot remove package directory: " + err.Error())
		return
	}

	for _, leftover := range []string{
		filepath.Join(app_home_dir, "."+uid+".installing"),
		filepath.Join(app_home_dir, "."+uid+".previous"),
	} {
		if err := os.RemoveAll(leftover); err != nil && !os.IsNotExist(err) {
			callback.Warn("Cannot remove leftover package data: " + err.Error())
		} else {
			callback.Log("Removed leftover package data: " + leftover)
		}
	}

	callback.Success(fmt.Sprintf("Package %s uninstalled successfully.", config.Metadata.PackageName))
}
