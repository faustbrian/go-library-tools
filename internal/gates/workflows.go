package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	actionlintVersion     = "v1.7.12"
	maximumWorkflowOutput = 4 << 20
	maximumWorkflowFiles  = 512
)

var immutableWorkflowRef = regexp.MustCompile(`^[0-9a-f]{40}$`)
var immutableWorkflowImage = regexp.MustCompile(`^docker://[^@\s]+@sha256:[0-9a-f]{64}$`)
var immutableContainerImage = regexp.MustCompile(`^[^@\s]+@sha256:[0-9a-f]{64}$`)

type boundedWorkflowBuffer struct {
	data     bytes.Buffer
	overflow bool
}

func (buffer *boundedWorkflowBuffer) Write(value []byte) (int, error) {
	remaining := maximumWorkflowOutput - buffer.data.Len()
	if len(value) <= remaining {
		return buffer.data.Write(value)
	}
	buffer.overflow = true
	_, _ = buffer.data.Write(value[:remaining])
	return len(value), nil
}

// Workflows validates every GitHub Actions workflow with the centrally pinned
// Actionlint release.
func (runner Runner) Workflows(ctx context.Context) error {
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	var standardOutput, diagnostics boundedWorkflowBuffer
	err := runner.Executor.Run(ctx, Command{
		Name: "go", Dir: runner.Root, Env: map[string]string{"GOWORK": "off"},
		Args: []string{
			"run", "github.com/rhysd/actionlint/cmd/actionlint@" + actionlintVersion,
			"-no-color", "-oneline", "-shellcheck=", "-pyflakes=",
		},
		Stdout: &standardOutput,
		Stderr: &diagnostics,
	})
	if standardOutput.overflow || diagnostics.overflow {
		return fmt.Errorf("actionlint output exceeded %d bytes", maximumWorkflowOutput)
	}
	if err != nil {
		message := strings.TrimSpace(standardOutput.data.String() + "\n" + diagnostics.data.String())
		if message != "" {
			return fmt.Errorf("actionlint failed: %w: %s", err, message)
		}
		return fmt.Errorf("actionlint failed: %w", err)
	}
	if err := checkWorkflowSecurity(runner.Root); err != nil {
		return err
	}
	_, _ = io.WriteString(output, "workflow contract passed\n")
	return nil
}

func checkWorkflowSecurity(root string) error {
	directory := filepath.Join(root, ".github", "workflows")
	workflowRoot, err := os.OpenRoot(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("workflow security policy: %w", err)
	}
	defer workflowRoot.Close()
	var findings []string
	files := 0
	err = filepath.WalkDir(directory, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("workflow security policy: symbolic link %s", filepath.Base(path))
		}
		if extension := filepath.Ext(entry.Name()); extension != ".yml" && extension != ".yaml" {
			return nil
		}
		files++
		if files > maximumWorkflowFiles {
			return errors.New("workflow security policy: workflow file limit exceeded")
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > maximumWorkflowOutput {
			return fmt.Errorf("workflow security policy: %s exceeds size limit", entry.Name())
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		file, err := workflowRoot.Open(relative)
		if err != nil {
			return err
		}
		content, readErr := io.ReadAll(io.LimitReader(file, maximumWorkflowOutput+1))
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			return errors.Join(readErr, closeErr)
		}
		if len(content) > maximumWorkflowOutput {
			return fmt.Errorf("workflow security policy: %s exceeds size limit", entry.Name())
		}
		workflowFindings, err := inspectWorkflow(content)
		if err != nil {
			return fmt.Errorf("%s: %w", entry.Name(), err)
		}
		for _, finding := range workflowFindings {
			findings = append(findings, entry.Name()+": "+finding)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("workflow security policy: %w", err)
	}
	if len(findings) > 0 {
		sort.Strings(findings)
		return fmt.Errorf("workflow security policy failed: %s", strings.Join(findings, "; "))
	}
	return nil
}

func inspectWorkflow(content []byte) ([]string, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode workflow: %w", err)
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("workflow must contain exactly one YAML document")
		}
		return nil, fmt.Errorf("decode workflow: %w", err)
	}
	var findings []string
	if err := inspectWorkflowNode(&document, &findings); err != nil {
		return nil, err
	}
	return findings, nil
}

func inspectWorkflowNode(document *yaml.Node, findings *[]string) error {
	type item struct {
		node  *yaml.Node
		path  []string
		depth int
	}
	stack := []item{{node: document}}
	visited := 0
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		visited++
		if visited > 100_000 || current.depth > 100 {
			return errors.New("workflow YAML structure limit exceeded")
		}
		if current.node.Kind == yaml.AliasNode {
			if current.node.Alias == nil {
				return errors.New("workflow contains an unresolved YAML alias")
			}
			stack = append(stack, item{node: current.node.Alias, path: current.path, depth: current.depth + 1})
			continue
		}
		if current.node.Kind == yaml.DocumentNode || current.node.Kind == yaml.SequenceNode {
			path := current.path
			if current.node.Kind == yaml.SequenceNode {
				path = appendPath(path, "[]")
			}
			for _, child := range current.node.Content {
				stack = append(stack, item{node: child, path: path, depth: current.depth + 1})
			}
			continue
		}
		if current.node.Kind != yaml.MappingNode {
			continue
		}
		if workflowStepPath(current.path) {
			persists, err := checkoutCredentialsPersist(current.node)
			if err != nil {
				return err
			}
			if persists {
				*findings = append(*findings, "actions/checkout requires literal persist-credentials: false")
			}
		}
		keys := map[string]struct{}{}
		for index := 0; index+1 < len(current.node.Content); index += 2 {
			key, value := current.node.Content[index], current.node.Content[index+1]
			if _, exists := keys[key.Value]; exists {
				return fmt.Errorf("workflow contains duplicate key %q", key.Value)
			}
			keys[key.Value] = struct{}{}
			resolved, err := resolveWorkflowAlias(value)
			if err != nil {
				return err
			}
			switch key.Value {
			case "on":
				if len(current.path) == 0 && workflowEventPresent(resolved, "pull_request_target") {
					*findings = append(*findings, "pull_request_target is forbidden")
				}
			case "permissions":
				if workflowPermissionsPath(current.path) && resolved.Kind == yaml.ScalarNode && resolved.Value == "write-all" {
					*findings = append(*findings, "permissions write-all is forbidden")
				}
			case "uses":
				if workflowUsesPath(current.path) && (resolved.Kind != yaml.ScalarNode || !immutableActionReference(resolved.Value)) {
					*findings = append(*findings, "remote actions require immutable SHA references")
				}
			case "container":
				if workflowJobPath(current.path) && resolved.Kind == yaml.ScalarNode && !immutableContainerImage.MatchString(resolved.Value) {
					*findings = append(*findings, "job container images require immutable digest references")
				}
			case "image":
				if workflowContainerImagePath(current.path) && (resolved.Kind != yaml.ScalarNode || !immutableContainerImage.MatchString(resolved.Value)) {
					*findings = append(*findings, "container images require immutable digest references")
				}
			}
			childPath := appendPath(current.path, key.Value)
			if key.Value == "<<" {
				childPath = current.path
			}
			stack = append(stack, item{node: value, path: childPath, depth: current.depth + 1})
		}
	}
	return nil
}

func resolveWorkflowAlias(node *yaml.Node) (*yaml.Node, error) {
	seen := map[*yaml.Node]struct{}{}
	for depth := 0; node != nil && node.Kind == yaml.AliasNode; depth++ {
		if depth >= 100 {
			return nil, errors.New("workflow YAML alias limit exceeded")
		}
		if _, exists := seen[node]; exists {
			return nil, errors.New("workflow YAML alias cycle detected")
		}
		seen[node] = struct{}{}
		if node.Alias == nil {
			return nil, errors.New("workflow contains an unresolved YAML alias")
		}
		node = node.Alias
	}
	if node == nil {
		return nil, errors.New("workflow contains an unresolved YAML alias")
	}
	return node, nil
}

func appendPath(path []string, value string) []string {
	result := make([]string, len(path)+1)
	copy(result, path)
	result[len(path)] = value
	return result
}

func workflowPermissionsPath(path []string) bool {
	return len(path) == 0 || len(path) == 2 && path[0] == "jobs"
}

func workflowUsesPath(path []string) bool {
	return len(path) == 2 && path[0] == "jobs" ||
		len(path) == 4 && path[0] == "jobs" && path[2] == "steps" && path[3] == "[]"
}

func workflowJobPath(path []string) bool {
	return len(path) == 2 && path[0] == "jobs"
}

func workflowContainerImagePath(path []string) bool {
	return len(path) == 3 && path[0] == "jobs" && path[2] == "container" ||
		len(path) == 4 && path[0] == "jobs" && path[2] == "services"
}

func workflowStepPath(path []string) bool {
	return len(path) == 4 && path[0] == "jobs" && path[2] == "steps" && path[3] == "[]"
}

func checkoutCredentialsPersist(step *yaml.Node) (bool, error) {
	uses := workflowMappingValue(step, "uses", 0)
	if uses == nil {
		return false, nil
	}
	resolvedUses, err := resolveWorkflowAlias(uses)
	if err != nil {
		return false, err
	}
	if resolvedUses.Kind != yaml.ScalarNode {
		return false, nil
	}
	ownerAction, _, _ := strings.Cut(resolvedUses.Value, "@")
	if !strings.EqualFold(ownerAction, "actions/checkout") {
		return false, nil
	}
	with := workflowMappingValue(step, "with", 0)
	persist := workflowMappingValue(with, "persist-credentials", 0)
	if persist == nil {
		return true, nil
	}
	resolvedPersist, err := resolveWorkflowAlias(persist)
	if err != nil {
		return false, err
	}
	return resolvedPersist.Kind != yaml.ScalarNode || resolvedPersist.Value != "false", nil
}

func workflowMappingValue(node *yaml.Node, key string, depth int) *yaml.Node {
	if node == nil || depth > 100 {
		return nil
	}
	if node.Kind == yaml.AliasNode {
		return workflowMappingValue(node.Alias, key, depth+1)
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value != "<<" {
			continue
		}
		merged := node.Content[index+1]
		if merged.Kind == yaml.SequenceNode {
			for _, candidate := range merged.Content {
				if value := workflowMappingValue(candidate, key, depth+1); value != nil {
					return value
				}
			}
			continue
		}
		if value := workflowMappingValue(merged, key, depth+1); value != nil {
			return value
		}
	}
	return nil
}

func workflowEventPresent(node *yaml.Node, event string) bool {
	stack := []*yaml.Node{node}
	for visited := 0; len(stack) > 0 && visited < 100_000; visited++ {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current.Kind == yaml.AliasNode && current.Alias != nil {
			stack = append(stack, current.Alias)
			continue
		}
		switch current.Kind {
		case yaml.ScalarNode:
			if current.Value == event {
				return true
			}
		case yaml.SequenceNode:
			stack = append(stack, current.Content...)
		case yaml.MappingNode:
			for index := 0; index+1 < len(current.Content); index += 2 {
				if current.Content[index].Value == event {
					return true
				}
			}
		case yaml.DocumentNode, yaml.AliasNode:
		}
	}
	return false
}

func immutableActionReference(value string) bool {
	if strings.HasPrefix(value, "./") {
		return true
	}
	if strings.HasPrefix(value, "docker://") {
		return immutableWorkflowImage.MatchString(value)
	}
	at := strings.LastIndexByte(value, '@')
	return at > 0 && immutableWorkflowRef.MatchString(value[at+1:])
}
