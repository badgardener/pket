package main

import (
	"fmt"
	"os"

	"pket/cli"
	"pket/core"
)

const VERSION = "26.2.2"

type commandCallback struct {
	core.Callback
	failed bool
}

func (c *commandCallback) Error(msg string) {
	c.failed = true
	c.Callback.Error(msg)
}

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
	var callback *commandCallback

	if verbose {
		argv = append([]string{argv[0]}, argv[2:]...)
		argc--
		callback = &commandCallback{Callback: cli.VerboseCallback{}}
	} else {
		callback = &commandCallback{Callback: cli.NormalCallback{}}
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
		if callback.failed {
			os.Exit(1)
		}

	case "install":
		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Path is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Install(argv[2], callback)
		if callback.failed {
			os.Exit(1)
		}

	case "list":
		if verbose {
			fmt.Println(cli.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		if !List() {
			os.Exit(1)
		}

	case "info":
		if verbose {
			fmt.Println(cli.Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Package name is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		if !Info(argv[2]) {
			os.Exit(1)
		}

	case "uninstall":
		if argc != 3 {
			fmt.Println(cli.Error.Render("Error:") + " Package name is required and cli.Argument count should be exactly 2.")
			os.Exit(1)
		}

		core.Uninstall(argv[2], callback)
		if callback.failed {
			os.Exit(1)
		}

	case "--help", "-h":
		printHelp()

	case "--version":
		fmt.Println(cli.Title.Render("pket") + " version " + cli.Info.Render(VERSION))

	case "-v":
		fmt.Println(VERSION)

	default:
		fmt.Println(cli.Error.Render("Error:") + " invalid command " + cli.Argument.Render("'"+cmd+"'") + ".")
		fmt.Println("Use " + cli.Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}
}
