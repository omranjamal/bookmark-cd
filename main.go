package main

import (
	"fmt"
	bcd "github.com/omranjamal/bookmark-cd/bcd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var version = "vvvv"

func main() {
	bcd.SetupTerminal()

	isPrintShellScript := 0
	isInstall := 0

	alias := "bcd"
	shellFile := ""

	search := make([]string, 0, 8)

	for _, arg := range os.Args[1:] {
		if arg == "--shell" {
			isPrintShellScript = 1
		} else if arg == "--install" {
			isInstall = 1
		} else if arg == "--version" || arg == "-v" {
			os.Stderr.WriteString("bookmark-cd " + version + "\n")
			return
		} else if arg == "--help" || arg == "-h" {
			os.Stderr.WriteString(bcd.HelpText + "\n")
			return
		} else {
			if isPrintShellScript == 1 {
				alias = arg
			} else if isInstall == 1 {
				if shellFile != "" {
					alias = arg
				} else {
					absolutePath, absolutePathError := filepath.Abs(arg)

					if absolutePathError != nil {
						log.Fatal(absolutePathError)
					}

					shellFile = absolutePath
				}
			} else {
				search = append(search, arg)
			}
		}
	}

	if (isPrintShellScript + isInstall) > 1 {
		os.Stderr.WriteString("ERROR: can't use --shell and --install together\n")
		os.Exit(1)
		return
	}

	if isInstall == 1 {
		if shellFile == "" {
			os.Stderr.WriteString("ERROR: must provide a shell file to modify\n")
			os.Exit(1)
			return
		} else {
			bcd.Install(shellFile, alias)
			return
		}
	}

	if isPrintShellScript == 1 {
		fmt.Fprintf(
			os.Stdout,
			"%s\n",
			strings.Replace(
				bcd.ShellFunction,
				"bcd",
				alias,
				1,
			),
		)

		return
	}

	initialSearchText := strings.Join(search, " ")
	bcd.Run(initialSearchText)
}
