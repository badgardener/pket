package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"pket/cli"
	"pket/core"
)

func List() bool {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(cli.Error.Render("Cannot determine home directory: " + err.Error()))
		return false
	}

	app_home_dir := filepath.Join(home_dir, core.Package_data_suffix())
	entries, err := os.ReadDir(app_home_dir)
	if os.IsNotExist(err) {
		fmt.Println(cli.Warning.Render("No packages installed."))
		return true
	}
	if err != nil {
		fmt.Println(cli.Error.Render("Cannot read package directory: " + err.Error()))
		return false
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
		fmt.Println(cli.Warning.Render("No packages installed."))
		return true
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].name < packages[j].name
	})

	name_style := cli.Title
	uid_style := cli.Command
	version_style := cli.Success
	for _, package_item := range packages {
		fmt.Printf("%s %s %s\n",
			name_style.Render(package_item.name),
			uid_style.Render(package_item.uid),
			version_style.Render(package_item.version),
		)
	}
	return true
}

func Info(pack string) bool {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		print_info_error("Cannot determine home directory: " + err.Error())
		return false
	}

	app_home_dir := filepath.Join(home_dir, core.Package_data_suffix())
	install_dir, err := core.Find_package(app_home_dir, pack)
	if err != nil {
		print_info_error(err.Error())
		return false
	}

	manifest_path := filepath.Join(install_dir, "pket-manifest", "pket-config.toml")
	data, err := os.ReadFile(manifest_path)
	if err != nil {
		print_info_error("Cannot read package manifest: " + err.Error())
		return false
	}

	var config core.BuiltConfig
	if _, err := toml.Decode(string(data), &config); err != nil {
		print_info_error("Cannot parse package manifest: " + err.Error())
		return false
	}
	if err := config.Validate(); err != nil {
		print_info_error("Invalid package manifest: " + err.Error())
		return false
	}

	heading_style := cli.Title
	label_style := cli.Info
	value_style := cli.Success
	path_style := cli.Command

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
		fmt.Println("  " + cli.Error.Render("Cannot read files.lst: "+err.Error()))
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
	return true
}

func print_info_error(message string) {
	style := cli.Error
	fmt.Println(style.Render(message))
}
