package main

import (
	"fmt"
	"os"

	"pket/core"

	"charm.land/lipgloss/v2"
)

const VERSION = "26.1.1"

var (
	Success  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	Error    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	Warning  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF"))
	Info     = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	Muted    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	Title    = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true)
	Command  = lipgloss.NewStyle().Foreground(lipgloss.Color("#94E2D5")).Bold(true)
	Flag     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387"))
	Argument = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5C2E7"))
	Package  = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
)

func printHelp() {
	fmt.Println(Title.Render("pket") + " - Package Management Tool")

	fmt.Println("\nUsage:")
	fmt.Println("  " + Command.Render("pket") + " " + Muted.Render("[--verbose]") + " " + Argument.Render("<command>") + " " + Muted.Render("[arguments]"))

	fmt.Println("\nCommands:")
	fmt.Println("  " + Command.Render("build") + " " + Argument.Render("<path>") + "       Build a package from a directory.")
	fmt.Println("  " + Command.Render("install") + " " + Argument.Render("<package>") + " " + Argument.Render("<directory>"))
	fmt.Println("                               Install a package to a directory.")
	fmt.Println("  " + Command.Render("list") + "                  List installed packages.")
	fmt.Println("  " + Command.Render("info") + " " + Argument.Render("<package>") + "       Show information about an installed package.")
	fmt.Println("  " + Command.Render("repair") + " " + Argument.Render("<package>") + "     Repair an installed package.")
	fmt.Println("  " + Command.Render("uninstall") + " " + Argument.Render("<package>") + "  Remove an installed package.")

	fmt.Println("\nOptions:")
	fmt.Println("  " + Flag.Render("--verbose") + "              Enable verbose output.")
	fmt.Println("  " + Flag.Render("--help") + ", " + Flag.Render("-h") + "          Show this help message.")
	fmt.Println("  " + Flag.Render("--version") + ", " + Flag.Render("-v") + "       Show version information.")

	fmt.Println("\nExamples:")
	fmt.Println("  " + Command.Render("pket build") + " " + Package.Render("./my-package"))
	fmt.Println("  " + Command.Render("pket install") + " " + Package.Render("./my-package.pkt") + " " + Package.Render("/opt/packages"))
	fmt.Println("  " + Command.Render("pket list"))
	fmt.Println("  " + Command.Render("pket info") + " " + Package.Render("my-package"))
	fmt.Println("  " + Command.Render("pket repair") + " " + Package.Render("my-package"))
	fmt.Println("  " + Command.Render("pket uninstall") + " " + Package.Render("my-package"))
}

func main() {
	argc := len(os.Args)
	argv := os.Args

	if argc == 1 {
		fmt.Println("pket is " + Success.Render("installed.") + " Use " + Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd := argv[1]

	verbose := cmd == "--verbose"
	var callback core.Callback

	if verbose {
		argv = append([]string{argv[0]}, argv[2:]...)
		argc--
		callback = VerboseCallback{}
	} else {
		callback = NormalCallback{}
	}

	if argc == 1 {
		fmt.Println("pket is " + Success.Render("installed.") + " Use " + Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}

	cmd = argv[1]

	switch cmd {
	case "build":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Path is required.")
			os.Exit(1)
		}

		core.Build(argv[2], callback)

	case "install":
		if argc < 4 {
			fmt.Println(Error.Render("Error:") + " Not enough arguments.")
			os.Exit(1)
		}

		core.Install(argv[2], argv[3], callback)

	case "list":
		if verbose {
			fmt.Println(Error.Render("Error:") + " Verbose not allowed here.")
			os.Exit(1)
		}

		core.List()

	case "repair":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

		core.Repair(argv[2], callback)

	case "info":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

		core.Info(argv[2], verbose)

	case "uninstall":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

		core.Uninstall(argv[2], callback)

	case "--help", "-h":
		printHelp()

	case "--version", "-v":
		fmt.Println(Title.Render("pket") + " version " + Info.Render(VERSION))

	default:
		fmt.Println(Error.Render("Error:") + " invalid command " + Argument.Render("'"+cmd+"'") + ".")
		fmt.Println("Use " + Info.Render("pket --help") + " for usage.")
		os.Exit(1)
	}
}
