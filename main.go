package main

import (
	"fmt"
	"os"
	"strings"

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
	fmt.Println("  " + Command.Render("pket") + Muted.Render("[--silent|--verbose]") +
		" " + Argument.Render("<command>") + " " + Muted.Render("[arguments]"))
	fmt.Println("\nCommands:")
	fmt.Println("  " + Command.Render("build") + "      Build a package.")
	fmt.Println("  " + Command.Render("install") + "    Install a package.")
	fmt.Println("  " + Command.Render("list") + "       List installed packages.")
	fmt.Println("  " + Command.Render("info") + "       Info of the installed package.")
	fmt.Println("  " + Command.Render("repair") + "     Repair an installed package.")
	fmt.Println("  " + Command.Render("uninstall") + "   Remove an installed package.")
	fmt.Println("\nOptions:")
	fmt.Println("  " + Flag.Render("--help") + "       Show help.")
	fmt.Println("  " + Flag.Render("--version") + "    Show version.")
	fmt.Println("\nExamples:")
	fmt.Println("  " + Command.Render("pket build") + " " + Package.Render("."))
	fmt.Println("  " + Command.Render("pket install") + " " + Package.Render("package.pkt"))
	fmt.Println("  " + Command.Render("pket list"))
	fmt.Println("  " + Command.Render("pket repair") + " " + Package.Render("<package>"))
	fmt.Println("  " + Command.Render("pket uninstall") + " " + Package.Render("<package>"))
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

	if verbose {
		argv = append([]string{argv[0]}, argv[2:]...)
	}

	switch cmd {
	case "build":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Path is required.")
			os.Exit(1)
		}

		fmt.Println("Building " + Package.Render(strings.Join(argv[2:], " ")) + "...")

	case "install":
		if argc < 4 {
			fmt.Println(Error.Render("Error:") + " Not enough arguments.")
			os.Exit(1)
		}

		fmt.Println("Installing " + Package.Render(strings.Join(argv[2:], " ")) + "...")

	case "list":

	case "repair":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

		fmt.Println("Repairing " + Package.Render(strings.Join(argv[2:], " ")) + "...")

	case "info":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

	case "uninstall":
		if argc < 3 {
			fmt.Println(Error.Render("Error:") + " Package name is required.")
			os.Exit(1)
		}

		fmt.Println("Uninstalling " + Package.Render(strings.Join(argv[2:], " ")) + "...")

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
