package main

import (
	"os"

	"github.com/faustbrian/go-library-tools/internal/cli"
)

func main() {
	workingDirectory, err := os.Getwd()
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(cli.Execute(os.Args[1:], workingDirectory, os.Stdout, os.Stderr))
}
