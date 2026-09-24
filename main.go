package main

import (
	"fmt"
	"os"

	"pket/core"
	"pket/cli"
)

const VERSION = "26.2.1"

func printHelp() {
	fmt.Println(cli.Title.Render("pket") + " - Package Management Tool")

	fmt.Println("\nUsage:")
	fmt.Println("  " + cli.Command.Render("pket") + " " + cli.Muted.Render("[--verbose]") + " " + cli.Argument.Render("<command>") + " " + cli.Muted.Render("[arguments]"))

	fmt.Println("\nCommands:")
	fmt.Println("  " + cli.Command.Render("build") + " " + cli.Argument.Render("<path>") + "         Build a package from a directory.")
	fmt.Println("  " + cli.Command.Render("install") + " " + cli.Argument.Render("<package>") + "    Install a package to a directory.")
	fmt.Println("  " + cli.Command.Render("list") + "                 List installed packages.")
	fmt.Println("  " + cli.Command.Render("info") + " " + cli.Argument.Render("<package>") + "       Show information about an installed package.")
	fmt.Println("  " + cli.Command.Render("uninstall") + " " + cli.Argument.Render("<package>") + "  Remove an installed package.")

	fmt.Println("\nOptions:")
	fmt.Println("  " + cli.Flag.Render("--verbose") + "           Enable verbose output.")
	fmt.Println("  " + cli.Flag.Render("--help") + ", " + cli.Flag.Render("-h") + "          Show this help message.")
	fmt.Println("  " + cli.Flag.Render("--version") + ", " + cli.Flag.Render("-v") + "       Show version information.")

	fmt.Println("\nExamples:")
	fmt.Println("  " + cli.Command.Render("pket build") + " " + cli.Package.Render("./my-package"))
	fmt.Println("  " + cli.Command.Render("pket install") + " " + cli.Package.Render("./my-cli.Package.pkt"))
	fmt.Println("  " + cli.Command.Render("pket list"))
	fmt.Println("  " + cli.Command.Render("pket info") + " " + cli.Package.Render("my-package"))
	fmt.Println("  " + cli.Command.Render("pket uninstall") + " " + cli.Package.Render("my-package"))
}

func main() {
	argc := len(os.Args)
	argv := os.Args

	if argc == 1 {
		fmt.Println("pket is " + cli.Success.Render("installed.") + " Use " + cli.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd := argv[1]

	verbose := cmd == "--verbose"
	var callback core.Callback

	if verbose {
		argv = append([]string{argv[0]}, argv[2:]...)
		argc--
		callback = cli.VerboseCallback{}
	} else {
		callback = cli.NormalCallback{}
	}

	if argc == 1 {
		fmt.Println("pket is " + cli.Success.Render("installed.") + " Use " + cli.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd = argv[1]

	switch cmd {
	case "build":
		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Path is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Build(argv[2], callback)

	case "install":
		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Path is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Install(argv[2], callback)

	case "list":
		if verbose {
			fmt.Println(cli.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		List()

	case "info":
		if verbose {
			fmt.Println(cli.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Package name is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		Info(argv[2])

	case "uninstall":
		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Package name is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Uninstall(argv[2], callback)

	case "--help", "-h":
		printHelp()

	case "--version", "-v":
		fmt.Println(cli.Title.Render("pket") + " version " + cli.Info.Render(VERSION))

	default:
		fmt.Println(cli.Error.Render("Error:") + " invalid command " + cli.Argument.Render("'"+cmd+"'") + ".")
		fmt.Println("Use " + cli.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}
}
