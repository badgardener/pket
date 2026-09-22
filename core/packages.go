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
