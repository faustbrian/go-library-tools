package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/faustbrian/go-library-tools/internal/cohesion"
)

type cohesionV2Invocation struct {
	action        string
	view          string
	schemaVersion string
	mode          string
	repository    string
	inputs        string
	resolutionMap string
	output        string
}

// Function variables keep command dispatch independently testable while the
// production defaults remain the cohesion validators.
var checkSourcesV2 = cohesion.CheckSourcesV2
var verifySourcesV2 = cohesion.VerifySourcesV2

func isCohesionV2Invocation(args []string) bool {
	if len(args) >= 2 && args[0] == "catalog" && args[1] == "project" {
		return true
	}
	for _, arg := range args {
		if arg == "--schema-version" {
			return true
		}
	}
	return false
}

func executeCohesionV2(args []string, root string, stdout, stderr io.Writer) int {
	invocation, err := parseCohesionV2Invocation(args)
	if err != nil {
		return writeCohesionV2Diagnostic(stderr, 2, cohesion.DiagnosticV3{Code: "invalid-invocation", Message: "invalid cohesion schema-v2 invocation"})
	}
	resolve := func(path string) string {
		if filepath.IsAbs(path) {
			clean := filepath.Clean(path)
			if canonical, err := filepath.EvalSymlinks(clean); err == nil {
				return canonical
			}
			return clean
		}
		return filepath.Join(root, path)
	}
	switch invocation.action {
	case "sources-check":
		if err := checkSourcesV2(resolve(invocation.inputs)); err != nil {
			if diagnostic, ok := cohesion.DiagnosticFromV3Error(err); ok {
				return writeCohesionV2Diagnostic(stderr, 1, diagnostic)
			}
			return writeCohesionV2Diagnostic(stderr, 1, cohesion.DiagnosticV3{Code: "resolution-failed", Message: "source input resolution failed"})
		}
		return 0
	case "sources-verify":
		if err := verifySourcesV2(resolve(invocation.inputs), invocation.repository, resolve(invocation.resolutionMap)); err != nil {
			if diagnostic, ok := cohesion.DiagnosticFromV3Error(err); ok {
				return writeCohesionV2Diagnostic(stderr, 1, diagnostic)
			}
			return writeCohesionV2Diagnostic(stderr, 1, cohesion.DiagnosticV3{Code: "resolution-failed", Message: "source verification failed"})
		}
		return 0
	default:
		return writeCohesionV2Diagnostic(stderr, 1, cohesion.DiagnosticV3{Code: "unsupported", Message: "cohesion schema-v2 operation is not implemented"})
	}
}

func writeCohesionV2Diagnostic(stderr io.Writer, exitCode int, diagnostic cohesion.DiagnosticV3) int {
	document := map[string]any{
		"diagnostics": []any{map[string]any{
			"code":    diagnostic.Code,
			"message": diagnostic.Message,
			"path":    diagnostic.Path,
		}},
		"schema_id":      "urn:golib:cohesion:diagnostic:v1",
		"schema_version": 1,
		"status":         "failed",
	}
	// The closed diagnostic document contains only marshal-safe primitives.
	encoded, _ := json.Marshal(document)
	encoded = append(encoded, '\n')
	_, _ = stderr.Write(encoded)
	return exitCode
}

func parseCohesionV2Invocation(args []string) (cohesionV2Invocation, error) {
	var invocation cohesionV2Invocation
	var flagArgs []string
	var allowed, required map[string]bool

	switch {
	case len(args) >= 3 && args[0] == "catalog" && args[1] == "project" && (args[2] == "consumer" || args[2] == "engineering"):
		invocation.action, invocation.view = "project", args[2]
		flagArgs = args[3:]
		allowed = flagSet("--schema-version", "--mode", "--repository", "--inputs", "--resolution-map", "--output")
		required = allowed
	case len(args) >= 4 && args[0] == "catalog" && args[1] == "project" && args[2] == "recover" && (args[3] == "consumer" || args[3] == "engineering"):
		invocation.action, invocation.view = "project-recover", args[3]
		flagArgs = args[4:]
		allowed = flagSet("--schema-version", "--mode", "--repository", "--inputs", "--resolution-map", "--output")
		required = allowed
	case len(args) >= 2 && args[0] == "aggregate" && (args[1] == "generate" || args[1] == "check" || args[1] == "recover"):
		invocation.action = "aggregate-" + args[1]
		flagArgs = args[2:]
		allowed = flagSet("--schema-version", "--mode", "--inputs", "--resolution-map", "--output")
		required = allowed
	case len(args) >= 2 && args[0] == "sources" && args[1] == "check":
		invocation.action = "sources-check"
		flagArgs = args[2:]
		allowed = flagSet("--schema-version", "--inputs")
		required = allowed
	case len(args) >= 2 && args[0] == "sources" && args[1] == "verify":
		invocation.action = "sources-verify"
		flagArgs = args[2:]
		allowed = flagSet("--schema-version", "--inputs", "--repository", "--resolution-map")
		required = allowed
	default:
		return cohesionV2Invocation{}, errors.New("unknown cohesion schema-v2 command")
	}

	values, err := parseSingletonFlags(flagArgs, allowed, required)
	if err != nil {
		return cohesionV2Invocation{}, err
	}
	if values["--schema-version"] != "2" {
		return cohesionV2Invocation{}, errors.New("--schema-version must be 2")
	}
	if mode := values["--mode"]; mode != "" && mode != "preview" && mode != "final" {
		return cohesionV2Invocation{}, errors.New("--mode must be preview or final")
	}
	invocation.schemaVersion = values["--schema-version"]
	invocation.mode = values["--mode"]
	invocation.repository = values["--repository"]
	invocation.inputs = values["--inputs"]
	invocation.resolutionMap = values["--resolution-map"]
	invocation.output = values["--output"]
	return invocation, nil
}

func flagSet(names ...string) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = true
	}
	return result
}

func parseSingletonFlags(args []string, allowed, required map[string]bool) (map[string]string, error) {
	if len(args)%2 != 0 {
		return nil, errors.New("flag is missing its value")
	}
	values := make(map[string]string)
	for index := 0; index < len(args); index += 2 {
		name, value := args[index], args[index+1]
		if !allowed[name] {
			return nil, fmt.Errorf("unknown flag %s", name)
		}
		if value == "" {
			return nil, fmt.Errorf("flag %s has an empty value", name)
		}
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("flag %s is repeated", name)
		}
		values[name] = value
	}
	for name := range required {
		if values[name] == "" {
			return nil, fmt.Errorf("required flag %s is missing", name)
		}
	}
	return values, nil
}
