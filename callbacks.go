package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type BaseCallback struct{}

func (l BaseCallback) Log(msg string) {}

func (l BaseCallback) Info(msg string) {
	fmt.Println(Info.Render("[ INFO  ]") + " " + msg)
}

func (l BaseCallback) Success(msg string) {
	fmt.Println(Success.Render("[SUCCESS]") + " " + msg)
}

func (l BaseCallback) Warn(msg string) {
	fmt.Println(Warning.Render("[WARNING]") + " " + msg)
}

func (l BaseCallback) Error(msg string) {
	fmt.Println(Error.Render("[ ERROR ]") + " " + msg)
}

func (l BaseCallback) Fatal(msg string) {
	fmt.Println(Error.Render("[ FATAL ] " + msg))
}

func (l BaseCallback) Prompt(msg string, def bool) bool {
	token := " [y/N] "
	if def {
		token = " [Y/n] "
	}

	fmt.Print(Info.Render(msg) + token)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	usr := strings.TrimSpace(strings.ToLower(scanner.Text()))

	switch usr {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return def
	}
}

type NormalCallback struct {
	BaseCallback
}

type VerboseCallback struct {
	BaseCallback
}

func (l VerboseCallback) Log(msg string) {
	fmt.Println(Muted.Render("[  LOG  ] " + msg))
}
