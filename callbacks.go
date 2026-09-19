package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type NormalCallback struct{}
type VerboseCallback struct{}

func (l NormalCallback) Log(msg string) {}
func (l NormalCallback) Info(msg string) {
	fmt.Println(Info.Render("[ INFO  ]") + " " + msg)
}

func (l NormalCallback) Success(msg string) {
	fmt.Println(Success.Render("[SUCCESS]") + " " + msg)
}

func (l NormalCallback) Warn(msg string) {
	fmt.Println(Warning.Render("[WARNING]") + " " + msg)
}

func (l NormalCallback) Error(msg string) {
	fmt.Println(Error.Render("[ ERROR ]") + " " + msg)
}

func (l NormalCallback) Fatal(msg string) {
	fmt.Println(Error.Render("[ FATAL ] " + msg))
}

func (l NormalCallback) Prompt(msg string, def bool) bool {
	var token string
	if def {
		token = " [Y/n] "
	} else {
		token = " [y/N] "
	}

	fmt.Print(Info.Render(msg) + token)

	scanner := bufio.NewScanner(os.Stdin)
	usr := scanner.Text()
	usr = strings.Trim(strings.ToLower(usr), " ")

	if usr == "y" || usr == "yes" {
		return true
	}

	if usr == "n" || usr == "no" {
		return false
	}

	return def
}

func (l VerboseCallback) Log(msg string) {
	fmt.Println(Muted.Render("[  LOG  ] " + msg))
}

func (l VerboseCallback) Info(msg string) {
	fmt.Println(Info.Render("[ INFO  ]") + " " + msg)
}

func (l VerboseCallback) Success(msg string) {
	fmt.Println(Success.Render("[SUCCESS]") + " " + msg)
}

func (l VerboseCallback) Warn(msg string) {
	fmt.Println(Warning.Render("[WARNING]") + " " + msg)
}

func (l VerboseCallback) Error(msg string) {
	fmt.Println(Error.Render("[ ERROR ]") + " " + msg)
}

func (l VerboseCallback) Fatal(msg string) {
	fmt.Println(Error.Render("[ FATAL ] " + msg))
}

func (l VerboseCallback) Prompt(msg string, def bool) bool {
	var token string
	if def {
		token = " [Y/n] "
	} else {
		token = " [y/N] "
	}

	fmt.Print(Info.Render(msg) + token)

	scanner := bufio.NewScanner(os.Stdin)
	usr := scanner.Text()
	usr = strings.Trim(strings.ToLower(usr), " ")

	if usr == "y" || usr == "yes" {
		return true
	}

	if usr == "n" || usr == "no" {
		return false
	}

	return def
}
