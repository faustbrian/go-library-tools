package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCommandArgumentsOmitsExecutable(t *testing.T) {
	original := os.Args
	os.Args = []string{"golib", "inventory", "--json"}
	t.Cleanup(func() { os.Args = original })

	arguments := commandArguments()
	if strings.Join(arguments, " ") != "inventory --json" {
		t.Fatalf("commandArguments() = %#v", arguments)
	}
}

func TestRunDelegatesToCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, func() (string, error) { return t.TempDir(), nil }, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "golib check") {
		t.Fatalf("run() = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestRunReportsWorkingDirectoryFailure(t *testing.T) {
	var stdout, stderr bytes.Buffer
	failure := errors.New("working directory unavailable")
	code := run(nil, func() (string, error) { return "", failure }, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || stderr.String() != failure.Error()+"\n" {
		t.Fatalf("run() = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestMainUsesProcessBoundaries(t *testing.T) {
	originalArguments := commandArguments
	originalWorkingDirectory := getWorkingDirectory
	originalOutput := standardOutput
	originalError := standardError
	originalExit := exitProcess
	t.Cleanup(func() {
		commandArguments = originalArguments
		getWorkingDirectory = originalWorkingDirectory
		standardOutput = originalOutput
		standardError = originalError
		exitProcess = originalExit
	})

	var stdout, stderr bytes.Buffer
	exitCode := -1
	commandArguments = func() []string { return []string{"--help"} }
	getWorkingDirectory = func() (string, error) { return t.TempDir(), nil }
	standardOutput = &stdout
	standardError = &stderr
	exitProcess = func(code int) { exitCode = code }

	main()
	if exitCode != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "golib check") {
		t.Fatalf("main() exit = %d, stdout = %q, stderr = %q", exitCode, stdout.String(), stderr.String())
	}
}
