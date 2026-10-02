package gates

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// One inspection owns the entire workflow/local-descriptor graph budget.
type localActionInspection struct {
	ctx              context.Context
	root             string
	active, complete map[string]bool
	files, bytes     int
	fileOperations   *workflowDescriptorFiles
}

func (inspection *localActionInspection) inspect(content []byte, depth int, action bool) error {
	if inspection.ctx != nil {
		if err := inspection.ctx.Err(); err != nil {
			return err
		}
	}
	document, findings, err := inspectWorkflowDocument(content)
	if err != nil {
		return errors.New("invalid local action descriptor")
	}
	if action && len(findings) != 0 {
		return errors.New("local action security policy failed")
	}
	return inspection.walkDocument(document, depth, action)
}

func (inspection *localActionInspection) inspectValidated(document *yaml.Node, depth int, action bool) error {
	if inspection.ctx != nil {
		if err := inspection.ctx.Err(); err != nil {
			return err
		}
	}
	return inspection.walkDocument(document, depth, action)
}

// The shared workflow owner has validated this parser-owned graph. Its aliases
// target non-alias nodes, and its full traversal already enforces YAML bounds.
func (inspection *localActionInspection) walkDocument(document *yaml.Node, depth int, action bool) error {
	var walk func(*yaml.Node, []string) error
	walk = func(node *yaml.Node, path []string) error {
		if node.Kind == yaml.AliasNode {
			return walk(node.Alias, path)
		}
		if node.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				resolved := value
				if resolved.Kind == yaml.AliasNode {
					resolved = resolved.Alias
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
				if err := walk(value, child); err != nil {
					return err
				}
			}
		} else {
			if node.Kind == yaml.SequenceNode {
				path = appendPath(path, "[]")
			}
			for _, child := range node.Content {
				if err := walk(child, path); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return walk(document, nil)
}

func (inspection *localActionInspection) local(reference string, depth int) error {
	name := strings.TrimPrefix(reference, "./")
	if name == "" || !filepath.IsLocal(name) || filepath.ToSlash(filepath.Clean(name)) != name || strings.Contains(name, "\\") {
		return errors.New("local action path is not repository-contained and canonical")
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
	files := operatingWorkflowDescriptorFiles()
	if inspection.fileOperations != nil {
		files = *inspection.fileOperations
		// Existing scoped workflow readers need no stat override.
		if files.stat == nil {
			files.stat = os.Stat
		}
	}
	info, err := files.stat(current)
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
	root, err := files.openRoot(inspection.root)
	if err != nil {
		return errors.New("local action root unavailable")
	}
	defer root.Close()
	file, err := files.open(root, name)
	if err != nil {
		return errors.New("local action descriptor unavailable")
	}
	content, readErr := io.ReadAll(io.LimitReader(file, maximumWorkflowOutput+1))
	closeErr := files.close(file)
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
