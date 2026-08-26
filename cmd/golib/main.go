package main

import (
	"io"
	"os"

	"github.com/faustbrian/go-library-tools/internal/cli"
)

var (
	commandArguments              = func() []string { return os.Args[1:] }
	getWorkingDirectory           = os.Getwd
	standardOutput      io.Writer = os.Stdout
	standardError       io.Writer = os.Stderr
	exitProcess                   = os.Exit
)

func main() {
	exitProcess(run(commandArguments(), getWorkingDirectory, standardOutput, standardError))
}

func run(args []string, getwd func() (string, error), stdout, stderr io.Writer) int {
	workingDirectory, err := getwd()
	if err != nil {
		_, _ = io.WriteString(stderr, err.Error()+"\n")
		return 1
	}
	return cli.Execute(args, workingDirectory, stdout, stderr)
}
