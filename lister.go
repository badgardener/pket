package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"pket/core"
	"pket/ui"
)

func List() {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(ui.Error.Render("Cannot determine home directory: " + err.Error()))
		return
	}

	app_home_dir := filepath.Join(home_dir, core.Package_data_suffix())
	entries, err := os.ReadDir(app_home_dir)
	if os.IsNotExist(err) {
		fmt.Println(ui.Warning.Render("No packages installed."))
		return
	}
	if err != nil {
		fmt.Println(ui.Error.Render("Cannot read package directory: " + err.Error()))
		return
	}

	type package_info struct {
		name    string
		uid     string
		version string
	}

	packages := []package_info{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		manifest_path := filepath.Join(app_home_dir, entry.Name(), "pket-manifest", "pket-config.toml")
		data, err := os.ReadFile(manifest_path)
		if err != nil {
			continue
		}

		var config core.BuiltConfig
		if _, err := toml.Decode(string(data), &config); err != nil {
			continue
		}
		if err := config.Validate(); err != nil {
			continue
		}

		version := make([]string, len(config.Metadata.Version))
		for i, part := range config.Metadata.Version {
			version[i] = fmt.Sprint(part)
		}
		packages = append(packages, package_info{
			name:    config.Metadata.PackageName,
			uid:     config.Metadata.PackageUID,
			version: strings.Join(version, "."),
		})
	}

	if len(packages) == 0 {
		fmt.Println(ui.Warning.Render("No packages installed."))
		return
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].name < packages[j].name
	})

	name_style := ui.Title
	uid_style := ui.Command
	version_style := ui.Success
	for _, package_item := range packages {
		fmt.Printf("%s %s %s\n",
			name_style.Render(package_item.name),
			uid_style.Render(package_item.uid),
			version_style.Render(package_item.version),
		)
	}
}

func Info(pack string) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		print_info_error("Cannot determine home directory: " + err.Error())
		return
	}

	app_home_dir := filepath.Join(home_dir, core.Package_data_suffix())
	install_dir, err := core.Find_package(app_home_dir, pack)
	if err != nil {
		print_info_error(err.Error())
		return
	}

	manifest_path := filepath.Join(install_dir, "pket-manifest", "pket-config.toml")
	data, err := os.ReadFile(manifest_path)
	if err != nil {
		print_info_error("Cannot read package manifest: " + err.Error())
		return
	}

	var config core.BuiltConfig
	if _, err := toml.Decode(string(data), &config); err != nil {
		print_info_error("Cannot parse package manifest: " + err.Error())
		return
	}
	if err := config.Validate(); err != nil {
		print_info_error("Invalid package manifest: " + err.Error())
		return
	}

	heading_style := ui.Title
	label_style := ui.Info
	value_style := ui.Success
	path_style := ui.Command

	version := make([]string, len(config.Metadata.Version))
	for i, part := range config.Metadata.Version {
		version[i] = fmt.Sprint(part)
	}

	fmt.Println(heading_style.Render(config.Metadata.PackageName))
	fmt.Println(label_style.Render("Package UID:") + " " + value_style.Render(config.Metadata.PackageUID))
	fmt.Println(label_style.Render("Version:") + " " + value_style.Render(strings.Join(version, ".")))
	fmt.Println(label_style.Render("Description:") + " " + config.Metadata.Description)
	fmt.Println(label_style.Render("Authors:") + " " + strings.Join(config.Metadata.Authors, ", "))
	fmt.Println(label_style.Render("Install directory:") + " " + path_style.Render(install_dir))

	fmt.Println(heading_style.Render("Executables"))
	for _, executable := range config.Executables {
		resolved := strings.ReplaceAll(executable.Path, "@res", filepath.Join(install_dir, "payload"))
		fmt.Println("  " + path_style.Render(resolved) + " -> " + value_style.Render(executable.Link))
	}

	fmt.Println(heading_style.Render("Install commands"))
	if config.Install.PreInstall == "" && config.Install.PostInstall == "" {
		fmt.Println("  " + "None")
	} else {
		if config.Install.PreInstall != "" {
			fmt.Println("  " + label_style.Render("Preinstall:") + " " + config.Install.PreInstall)
		}
		if config.Install.PostInstall != "" {
			fmt.Println("  " + label_style.Render("Postinstall:") + " " + config.Install.PostInstall)
		}
	}

	fmt.Println(heading_style.Render("External files"))
	files_path := filepath.Join(install_dir, "pket-manifest", "files.lst")
	files_data, err := os.ReadFile(files_path)
	if os.IsNotExist(err) {
		fmt.Println("  " + "None")
	} else if err != nil {
		fmt.Println("  " + ui.Error.Render("Cannot read files.lst: "+err.Error()))
	} else {
		found := false
		for _, line := range strings.Split(string(files_data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			found = true
			fmt.Println("  " + path_style.Render(line))
		}
		if !found {
			fmt.Println("  " + "None")
		}
	}
}

func print_info_error(message string) {
	style := ui.Error
	fmt.Println(style.Render(message))
}
