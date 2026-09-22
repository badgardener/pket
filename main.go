package main

import (
	"fmt"
	"os"

	"pket/core"
	"pket/ui"
)

const VERSION = "26.1.8"

func printHelp() {
	fmt.Println(ui.Title.Render("pket") + " - Package Management Tool")

	fmt.Println("\nUsage:")
	fmt.Println("  " + ui.Command.Render("pket") + " " + ui.Muted.Render("[--verbose]") + " " + ui.Argument.Render("<ui.Command>") + " " + ui.Muted.Render("[ui.Arguments]"))

	fmt.Println("\nui.Commands:")
	fmt.Println("  " + ui.Command.Render("build") + " " + ui.Argument.Render("<path>") + "         Build a package from a directory.")
	fmt.Println("  " + ui.Command.Render("install") + " " + ui.Argument.Render("<package>") + "    Install a package to a directory.")
	fmt.Println("  " + ui.Command.Render("list") + "                 List installed packages.")
	fmt.Println("  " + ui.Command.Render("info") + " " + ui.Argument.Render("<package>") + "       Show information about an installed ui.Package.")
	fmt.Println("  " + ui.Command.Render("uninstall") + " " + ui.Argument.Render("<package>") + "  Remove an installed ui.Package.")

	fmt.Println("\nOptions:")
	fmt.Println("  " + ui.Flag.Render("--verbose") + "           Enable verbose output.")
	fmt.Println("  " + ui.Flag.Render("--help") + ", " + ui.Flag.Render("-h") + "          Show this help message.")
	fmt.Println("  " + ui.Flag.Render("--version") + ", " + ui.Flag.Render("-v") + "       Show version information.")

	fmt.Println("\nExamples:")
	fmt.Println("  " + ui.Command.Render("pket build") + " " + ui.Package.Render("./my-package"))
	fmt.Println("  " + ui.Command.Render("pket install") + " " + ui.Package.Render("./my-ui.Package.pkt") + " " + ui.Package.Render("/opt/packages"))
	fmt.Println("  " + ui.Command.Render("pket list"))
	fmt.Println("  " + ui.Command.Render("pket info") + " " + ui.Package.Render("my-package"))
	fmt.Println("  " + ui.Command.Render("pket repair") + " " + ui.Package.Render("my-package"))
	fmt.Println("  " + ui.Command.Render("pket uninstall") + " " + ui.Package.Render("my-package"))
}

func main() {
	argc := len(os.Args)
	argv := os.Args

	if argc == 1 {
		fmt.Println("pket is " + ui.Success.Render("installed.") + " Use " + ui.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd := argv[1]

	verbose := cmd == "--verbose"
	var callback core.Callback

	if verbose {
		argv = append([]string{argv[0]}, argv[2:]...)
		argc--
		callback = ui.VerboseCallback{}
	} else {
		callback = ui.NormalCallback{}
	}

	if argc == 1 {
		fmt.Println("pket is " + ui.Success.Render("installed.") + " Use " + ui.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd = argv[1]

	switch cmd {
	case "build":
		if argc != 3 {
			fmt.Println(ui.Error.Render("Error:") + " Path is required and ui.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Build(argv[2], callback)

	case "install":
		if argc != 3 {
			fmt.Println(ui.Error.Render("Error:") + " Path is required and ui.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Install(argv[2], callback)

	case "list":
		if verbose {
			fmt.Println(ui.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		List()

	case "info":
		if verbose {
			fmt.Println(ui.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		if argc != 3 {
			fmt.Println(ui.Error.Render("Error:") + " Package name is required and ui.Argument count should be exactly 2.")
			os.Exit(1)
		}

		Info(argv[2])

	case "uninstall":
		if argc != 3 {
			fmt.Println(ui.Error.Render("Error:") + " Package name is required and ui.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Uninstall(argv[2], callback)

	case "--help", "-h":
		printHelp()

	case "--version", "-v":
		fmt.Println(ui.Title.Render("pket") + " version " + ui.Info.Render(VERSION))

	default:
		fmt.Println(ui.Error.Render("Error:") + " invalid ui.Command " + ui.Argument.Render("'"+cmd+"'") + ".")
		fmt.Println("Use " + ui.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}
}
