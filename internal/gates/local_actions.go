package gates

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// One inspection owns the entire workflow/local-descriptor graph budget.
type localActionInspection struct {
	ctx              context.Context
	root             string
	active, complete map[string]bool
	files, bytes     int
}

func (inspection *localActionInspection) inspect(content []byte, depth int, action bool) error {
	if inspection.ctx != nil {
		if err := inspection.ctx.Err(); err != nil {
			return err
		}
	}
	if depth > maximumLocalActionDepth {
		return errors.New("local action depth limit exceeded")
	}
	if action {
		findings, err := inspectWorkflow(content)
		if err != nil {
			return errors.New("invalid local action descriptor")
		}
		if len(findings) != 0 {
			return errors.New("local action security policy failed")
		}
	}
	var document yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(content)).Decode(&document); err != nil {
		return errors.New("invalid action YAML")
	}
	visited := 0
	var walk func(*yaml.Node, []string, int) error
	walk = func(node *yaml.Node, path []string, nesting int) error {
		visited++
		if visited > 100_000 || nesting > 100 {
			return errors.New("action YAML structure limit exceeded")
		}
		if node.Kind == yaml.AliasNode {
			resolved, err := resolveWorkflowAlias(node)
			if err != nil {
				return err
			}
			return walk(resolved, path, nesting+1)
		}
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				resolved, err := resolveWorkflowAlias(value)
				if err != nil {
					return err
				}
				if key.Value == "uses" && workflowUsesPath(path) && strings.HasPrefix(resolved.Value, "./") {
					if err := inspection.local(resolved.Value, depth+1); err != nil {
						return err
					}
				}
				if action && key.Value == "image" && len(path) == 1 && path[0] == "runs" {
					if resolved.Kind != yaml.ScalarNode || !immutableContainerImage.MatchString(resolved.Value) {
						return errors.New("docker action images require immutable digest references")
					}
				}
				child := appendPath(path, key.Value)
				if key.Value == "<<" {
					child = path
				}
				if err := walk(value, child, nesting+1); err != nil {
					return err
				}
			}
		} else {
			if node.Kind == yaml.SequenceNode {
				path = appendPath(path, "[]")
			}
			for _, child := range node.Content {
				if err := walk(child, path, nesting+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(&document, nil, 0)
}

func (inspection *localActionInspection) local(reference string, depth int) error {
	name := strings.TrimPrefix(reference, "./")
	if name == "" || !filepath.IsLocal(name) || filepath.ToSlash(filepath.Clean(name)) != name || strings.Contains(name, "\\") {
		return errors.New("local action path is not repository-contained and canonical")
	}
	if slices.Contains(strings.Split(name, "/"), "..") {
		return errors.New("local action traversal is forbidden")
	}
	if depth > maximumLocalActionDepth {
		return errors.New("local action depth limit exceeded")
	}
	if inspection.active[name] {
		return errors.New("local action cycle detected")
	}
	if inspection.complete[name] {
		return nil
	}
	current := inspection.root
	for part := range strings.SplitSeq(name, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("local action path missing or symbolic")
		}
	}
	info, err := os.Stat(current)
	if err != nil {
		return errors.New("local action path unavailable")
	}
	if info.IsDir() {
		name = filepath.Join(name, "action.yml")
		if _, err := os.Lstat(filepath.Join(inspection.root, name)); errors.Is(err, os.ErrNotExist) {
			name = strings.TrimSuffix(name, ".yml") + ".yaml"
		}
	}
	info, err = os.Lstat(filepath.Join(inspection.root, name))
	if err != nil || !info.Mode().IsRegular() || info.Size() > maximumWorkflowOutput {
		return errors.New("local action descriptor missing, symbolic, or oversized")
	}
	inspection.files++
	if inspection.files > maximumWorkflowFiles {
		return errors.New("local action descriptor count limit exceeded")
	}
	root, err := os.OpenRoot(inspection.root)
	if err != nil {
		return errors.New("local action root unavailable")
	}
	defer root.Close()
	file, err := root.Open(name)
	if err != nil {
		return errors.New("local action descriptor unavailable")
	}
	content, readErr := io.ReadAll(io.LimitReader(file, maximumWorkflowOutput+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(content) > maximumWorkflowOutput {
		return errors.New("local action descriptor read failed or oversized")
	}
	inspection.bytes += len(content)
	if inspection.bytes > maximumWorkflowBytes {
		return errors.New("workflow descriptor total byte limit exceeded")
	}
	key := strings.TrimPrefix(reference, "./")
	inspection.active[key] = true
	err = inspection.inspect(content, depth, true)
	delete(inspection.active, key)
	if err == nil {
		inspection.complete[key] = true
	}
	return err
}
