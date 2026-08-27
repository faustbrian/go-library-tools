// Command golib validates and executes the shared Go library contract.
package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/faustbrian/go-library-tools/internal/cli"
)

var (
	commandArguments              = func() []string { return os.Args[1:] }
	getWorkingDirectory           = os.Getwd
	standardOutput      io.Writer = os.Stdout
	standardError       io.Writer = os.Stderr
	exitProcess                   = os.Exit
	notifyContext                 = signal.NotifyContext
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
	ctx, stop := notifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cli.ExecuteContext(ctx, args, workingDirectory, stdout, stderr)
}
