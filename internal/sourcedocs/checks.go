package sourcedocs

import (
	"go/ast"
	"go/token"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
)

// These intentionally narrow patterns supplement, never replace, the canonical
// pinned secret scanner. No matched value is copied into reports or diagnostics.
var secretPattern = regexp.MustCompile(`(?:AKIA[0-9A-Z]{16}|(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk_(?:live|test)_[A-Za-z0-9]{16,}|-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----)`)
var markerPattern = regexp.MustCompile(`\b(TODO|FIXME|HACK|NOTE|WARNING)\b`)
var issuePattern = regexp.MustCompile(`(?:https://[^\s]+/(?:issues|pull)/[0-9]+|\b[A-Z][A-Z0-9]+-[0-9]+\b|#[0-9]+\b)`)
var linkPattern = regexp.MustCompile(`(?m)\[[^\]\n]+\]\(([^)\s]+)\)|^\s*\[[^\]\n]+\]:\s*(\S+)`)
var externalURLPattern = regexp.MustCompile(`https?://[^\s<>]+`)
var deprecationMarkerPattern = regexp.MustCompile(`(?i)^deprecated\s*:`)
var generatorPathPattern = regexp.MustCompile(`\.(?:go|php|py|sh|rb)$`)
var headerProducerPattern = regexp.MustCompile(`^by ("(?:[^"\\]|\\.)*"|[^\s;]+)(?:;|\s|$)`)

func validPackageComment(comment *ast.CommentGroup, name string) bool {
	if comment == nil {
		return false
	}
	text := strings.TrimSpace(comment.Text())
	prefix := "Package " + name
	if strings.HasPrefix(text, prefix) && len(text) > len(prefix) {
		next, _ := utf8.DecodeRuneInString(text[len(prefix):])
		if unicode.IsSpace(next) {
			return true
		}
	}
	return name == "main" && strings.HasPrefix(text, "Command ")
}

func validDeclarationComment(comment *ast.CommentGroup, name string) bool {
	if comment == nil {
		return false
	}
	text := strings.TrimSpace(comment.Text())
	if !strings.HasPrefix(text, name) {
		return false
	}
	if len(text) <= len(name) {
		return false
	}
	next, _ := utf8.DecodeRuneInString(text[len(name):])
	return !unicode.IsLetter(next) && !unicode.IsDigit(next) && next != '_'
}

func inspectFile(root string, file parsedFile, all []parsedFile, surface publicSurface, report *Report) {
	if file.file.Generated {
		checkGenerator(root, file, all, report)
		// Externally generated declarations stay visible in the inventory, but
		// hand-maintained API comment rules must not rewrite generator output.
	} else if file.file.Kind == "source" {
		for _, declaration := range file.ast.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Name.IsExported() && (declaration.Recv == nil || surface.methods[declaration]) {
					checkDeclaration(file, declaration.Name, declaration.Doc, report)
				}
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					switch specification := specification.(type) {
					case *ast.TypeSpec:
						if specification.Name.IsExported() {
							comment := specification.Doc
							if comment == nil {
								comment = declaration.Doc
							}
							checkDeclaration(file, specification.Name, comment, report)
						}
					case *ast.ValueSpec:
						for _, name := range specification.Names {
							if !name.IsExported() {
								continue
							}
							comment := specification.Doc
							if comment == nil {
								comment = specification.Comment
							}
							// Go permits a shared group comment for grouped values.
							if comment == nil && declaration.Doc != nil && declaration.Lparen.IsValid() {
								continue
							}
							if comment == nil {
								comment = declaration.Doc
							}
							if len(specification.Names) > 1 && comment != nil && validDeclarationComment(comment, specification.Names[0].Name) {
								continue
							}
							checkDeclaration(file, name, comment, report)
						}
					}
				}
			}
		}
		checkMembers(file, file.ast, surface, report)
	}
	for _, group := range file.ast.Comments {
		text := group.Text()
		if secretPattern.MatchString(text) {
			addFinding(report, file, group.Pos(), "secret-pattern", "error", "comment contains a forbidden credential-shaped pattern; remove it and run the secret scanner")
		}
		checkDeprecation(file, group, report)
		checkMarkers(file, group, report)
		checkLinks(root, file, group, report)
		if !file.file.Generated {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "// Code generated") && !ast.IsGenerated(file.ast) {
					addFinding(report, file, comment.Pos(), "generated-header", "error", "generated-code marker does not use the canonical Go header before the package clause")
				}
			}
		}
	}
	if file.file.Kind == "test" {
		for _, declaration := range file.ast.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, "Example") {
				continue
			}
			ast.Inspect(function, func(node ast.Node) bool {
				literal, ok := node.(*ast.BasicLit)
				if ok && literal.Kind == token.STRING {
					value, err := strconv.Unquote(literal.Value)
					if err == nil && secretPattern.MatchString(value) {
						addFinding(report, file, literal.Pos(), "secret-pattern", "error", "example contains a forbidden credential-shaped pattern; remove it and run the secret scanner")
					}
				}
				return true
			})
		}
	}
}

func checkDeclaration(file parsedFile, name *ast.Ident, comment *ast.CommentGroup, report *Report) {
	if !validDeclarationComment(comment, name.Name) {
		addFinding(report, file, name.Pos(), "exported-comment", "error", "exported declaration needs documentation beginning with its name")
	}
}

func checkMembers(file parsedFile, expression ast.Node, surface publicSurface, report *Report) {
	ast.Inspect(expression, func(node ast.Node) bool {
		var members *ast.FieldList
		switch node := node.(type) {
		case *ast.StructType:
			members = node.Fields
		case *ast.InterfaceType:
			members = node.Methods
		default:
			return true
		}
		for _, member := range members.List {
			if !surface.members[member] {
				continue
			}
			comment := member.Doc
			if comment == nil {
				comment = member.Comment
			}
			for _, name := range member.Names {
				if name.IsExported() {
					if len(member.Names) > 1 && validDeclarationComment(comment, member.Names[0].Name) {
						continue
					}
					checkDeclaration(file, name, comment, report)
				}
			}
		}
		return true
	})
}

type publicSurface struct {
	methods map[*ast.FuncDecl]bool
	members map[*ast.Field]bool
}

// consumerSurface follows type expressions from public declarations, never
// function bodies. Private concrete types can expose members through public
// signatures or embedding; private wire layouts and implementation seams cannot.
func consumerSurface(files []parsedFile) publicSurface {
	surface := publicSurface{methods: map[*ast.FuncDecl]bool{}, members: map[*ast.Field]bool{}}
	types := map[string][]ast.Expr{}
	methods := map[string][]*ast.FuncDecl{}
	functions := map[string][]*ast.FuncType{}
	roots := []ast.Expr{}
	for _, file := range files {
		if file.file.Kind != "source" || file.file.Generated {
			continue
		}
		for _, declaration := range file.ast.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Recv != nil {
					if name := receiverName(declaration.Recv.List[0].Type); name != "" {
						methods[name] = append(methods[name], declaration)
					}
				} else {
					functions[declaration.Name.Name] = append(functions[declaration.Name.Name], declaration.Type)
					if declaration.Name.IsExported() {
						roots = append(roots, declaration.Type)
					}
				}
			case *ast.GenDecl:
				for _, specification := range declaration.Specs {
					switch specification := specification.(type) {
					case *ast.TypeSpec:
						types[specification.Name.Name] = append(types[specification.Name.Name], specification.Type)
						if specification.TypeParams != nil {
							for _, parameter := range specification.TypeParams.List {
								types[specification.Name.Name] = append(types[specification.Name.Name], parameter.Type)
							}
						}
						if specification.Name.IsExported() {
							roots = append(roots, specification.Name)
						}
					case *ast.ValueSpec:
						if slices.ContainsFunc(specification.Names, func(name *ast.Ident) bool { return name.IsExported() }) {
							if specification.Type != nil {
								roots = append(roots, specification.Type)
							}
							roots = append(roots, specification.Values...)
						}
					}
				}
			}
		}
	}
	seen := map[ast.Expr]bool{}
	for len(roots) > 0 {
		expression := roots[len(roots)-1]
		roots = roots[:len(roots)-1]
		if expression == nil || seen[expression] {
			continue
		}
		seen[expression] = true
		switch expression := expression.(type) {
		case *ast.Ident:
			roots = append(roots, types[expression.Name]...)
			for _, method := range methods[expression.Name] {
				if method.Name.IsExported() {
					surface.methods[method] = true
					roots = append(roots, method.Type)
				}
			}
		case *ast.StarExpr:
			roots = append(roots, expression.X)
		case *ast.ParenExpr:
			roots = append(roots, expression.X)
		case *ast.ArrayType:
			roots = append(roots, expression.Elt)
		case *ast.MapType:
			roots = append(roots, expression.Key, expression.Value)
		case *ast.ChanType:
			roots = append(roots, expression.Value)
		case *ast.Ellipsis:
			roots = append(roots, expression.Elt)
		case *ast.IndexExpr:
			roots = append(roots, expression.X, expression.Index)
		case *ast.IndexListExpr:
			roots = append(roots, expression.X)
			roots = append(roots, expression.Indices...)
		case *ast.UnaryExpr:
			roots = append(roots, expression.X)
		case *ast.BinaryExpr:
			roots = append(roots, expression.X, expression.Y)
		case *ast.CompositeLit:
			roots = append(roots, expression.Type)
		case *ast.FuncLit:
			roots = append(roots, expression.Type)
		case *ast.CallExpr:
			callee := expression.Fun
			for {
				parenthesized, ok := callee.(*ast.ParenExpr)
				if !ok {
					break
				}
				callee = parenthesized.X
			}
			results := []*ast.FieldList{}
			switch callee := callee.(type) {
			case *ast.Ident:
				roots = append(roots, callee)
				for _, function := range functions[callee.Name] {
					results = append(results, function.Results)
				}
				if (callee.Name == "make" || callee.Name == "new") && len(expression.Args) > 0 {
					roots = append(roots, expression.Args[0])
				}
			case *ast.FuncLit:
				results = append(results, callee.Type.Results)
			}
			// Calling a private factory exposes its result, not its parameter
			// layouts. A function value exposes both via FuncLit/FuncType above.
			for _, fields := range results {
				if fields != nil {
					for _, field := range fields.List {
						roots = append(roots, field.Type)
					}
				}
			}
		case *ast.FuncType:
			for _, fields := range []*ast.FieldList{expression.TypeParams, expression.Params, expression.Results} {
				if fields != nil {
					for _, field := range fields.List {
						roots = append(roots, field.Type)
					}
				}
			}
		case *ast.StructType:
			for _, field := range expression.Fields.List {
				if len(field.Names) == 0 || slices.ContainsFunc(field.Names, func(name *ast.Ident) bool { return name.IsExported() }) {
					surface.members[field] = true
					roots = append(roots, field.Type)
				}
			}
		case *ast.InterfaceType:
			for _, field := range expression.Methods.List {
				if len(field.Names) == 0 || slices.ContainsFunc(field.Names, func(name *ast.Ident) bool { return name.IsExported() }) {
					surface.members[field] = true
					roots = append(roots, field.Type)
				}
			}
		}
	}
	return surface
}

func receiverName(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.StarExpr:
		return receiverName(expression.X)
	case *ast.IndexExpr:
		return receiverName(expression.X)
	case *ast.IndexListExpr:
		return receiverName(expression.X)
	}
	return ""
}

func checkDeprecation(file parsedFile, group *ast.CommentGroup, report *Report) {
	lines := strings.Split(strings.TrimSpace(group.Text()), "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		// A declaration may itself be named Deprecated. Only colon-form markers
		// carry Go deprecation syntax; ordinary name-leading prose is not a marker.
		if !deprecationMarkerPattern.MatchString(trimmed) {
			continue
		}
		if !strings.HasPrefix(trimmed, "Deprecated: ") || strings.TrimSpace(strings.TrimPrefix(trimmed, "Deprecated: ")) == "" || (index > 0 && strings.TrimSpace(lines[index-1]) != "") {
			addFinding(report, file, group.Pos(), "deprecated-syntax", "error", "deprecation must start a separate paragraph with Deprecated: and migration guidance")
		}
	}
}

func checkMarkers(file parsedFile, group *ast.CommentGroup, report *Report) {
	for line := range strings.SplitSeq(group.Text(), "\n") {
		location := markerPattern.FindStringIndex(line)
		if location == nil {
			continue
		}
		marker := line[location[0]:location[1]]
		body := strings.TrimSpace(strings.TrimLeft(line[location[1]:], ": "))
		severity, code, message := "review", "marker-review", "review marker semantics against implementation, owner, hazard, and removal condition"
		if body == "" {
			severity, code, message = "error", "marker-context", "marker needs an explanation of its missing behavior, constraint, or hazard"
		} else if (marker == "TODO" || marker == "FIXME" || marker == "HACK") && !issuePattern.MatchString(body) && (!strings.Contains(body, "owner=") || !strings.Contains(body, "trigger=")) {
			code, message = "marker-context", "verify debt marker has a tracked issue or enough owner and trigger context; free-form semantics need human review"
		}
		addFinding(report, file, group.Pos(), code, severity, message)
	}
}

func checkLinks(root string, file parsedFile, group *ast.CommentGroup, report *Report) {
	text := group.Text()
	for _, match := range linkPattern.FindAllStringSubmatch(text, maximumFindings+1) {
		target := match[1]
		if target == "" {
			target = match[2]
		}
		parsed, err := url.Parse(target)
		valid := err == nil
		if valid && parsed.Scheme != "" {
			valid = (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != "" && parsed.User == nil
			if parsed.Scheme == "mailto" {
				valid = parsed.Opaque != ""
			}
		} else if valid {
			valid = parsed.Host == "" && !path.IsAbs(parsed.Path) && !strings.Contains(parsed.Path, "\\")
			if valid && parsed.Path != "" {
				name := path.Clean(path.Join(path.Dir(file.relative), parsed.Path))
				valid = name != ".." && !strings.HasPrefix(name, "../")
				if valid {
					valid = repositoryfile.ValidateRegularFile(root, filepath.FromSlash(name)) == nil || repositoryfile.ValidateDirectory(root, filepath.FromSlash(name)) == nil
				}
			}
		}
		if !valid {
			addFinding(report, file, group.Pos(), "comment-link", "error", "comment link is malformed, unsupported, missing, unsafe, or outside the repository")
		}
	}
	for _, target := range externalURLPattern.FindAllString(text, maximumFindings+1) {
		// Sentence punctuation is not part of a bare specification URL.
		parsed, err := url.Parse(strings.TrimRight(target, ".,;:)]}"))
		if err != nil || parsed.Host == "" || parsed.User != nil {
			addFinding(report, file, group.Pos(), "comment-link", "error", "specification URL is malformed or contains embedded credentials")
		}
	}
}

func checkGenerator(root string, file parsedFile, all []parsedFile, report *Report) {
	headerAvailable, headerMissing, headerReview := false, false, false
	for _, group := range file.ast.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "// Code generated ") && strings.HasSuffix(comment.Text, " DO NOT EDIT.") {
				identity := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(comment.Text, "// Code generated "), " DO NOT EDIT."))
				if strings.Trim(identity, ";. ") == "" {
					addFinding(report, file, comment.Pos(), "generated-identity", "error", "canonical generated header must identify its generator or regeneration source")
				}
				// A canonical header may identify a real local producer directly;
				// A generation directive is one format, not the only ownership path.
				producer := headerProducerPattern.FindStringSubmatch(identity)
				if len(producer) == 0 {
					headerReview = true
					continue
				}
				word := producer[1]
				if strings.HasPrefix(word, "\"") {
					var err error
					word, err = strconv.Unquote(word)
					if err != nil {
						headerReview = true
						continue
					}
				}
				if !generatorPathPattern.MatchString(word) && !strings.Contains(word, "/") || strings.Contains(word, ":") || strings.ContainsAny(word, "$@") {
					headerReview = true
					continue
				}
				name := path.Clean(path.Join(path.Dir(file.relative), word))
				if path.IsAbs(word) || strings.Contains(word, "\\") || !file.tracked[name] || repositoryfile.ValidateRegularFile(root, filepath.FromSlash(name)) != nil {
					headerMissing = true
				} else {
					headerAvailable = true
				}
			}
		}
	}
	hasDeclaration, needsReview, unavailable := false, false, false
	for _, candidate := range all {
		if path.Dir(candidate.relative) != path.Dir(file.relative) {
			continue
		}
		for _, group := range candidate.ast.Comments {
			for _, comment := range group.List {
				body, isDirective := strings.CutPrefix(comment.Text, "//go:generate")
				if !isDirective || len(body) == 0 || (body[0] != ' ' && body[0] != '\t') {
					continue
				}
				hasDeclaration = true
				args, understood := generatorArguments(body)
				if !understood {
					needsReview = true
					continue
				}
				switch inspectGeneratorArguments(root, path.Dir(candidate.relative), args, file.tracked) {
				case generatorUnknown:
					needsReview = true
				case generatorMissing:
					unavailable = true
				case generatorAvailable:
					continue
				}
			}
		}
	}
	if (!hasDeclaration && !headerAvailable && !headerReview) || unavailable || headerMissing {
		addFinding(report, file, file.ast.Package, "generator-unavailable", "error", "generated file needs an available local header path or regeneration declaration; audit never executes it")
	}
	if needsReview || (!hasDeclaration && headerReview) {
		addFinding(report, file, file.ast.Package, "generator-review", "review", "unsupported, external, or ambiguous generator declaration needs manual availability and output verification; audit never expands or executes it")
	}
}

type generatorDisposition uint8

const (
	generatorUnknown generatorDisposition = iota
	generatorAvailable
	generatorMissing
)

// generatorArguments follows go generate's lexical double-quote rules, without
// evaluating its aliases or environment. Unknown or excessive forms remain
// explicit manual-review items rather than a fabricated availability result.
func generatorArguments(line string) ([]string, bool) {
	if len(line) > 64<<10 {
		return nil, false
	}
	arguments := []string{}
	for {
		line = strings.TrimLeft(line, " \t")
		if line == "" {
			return arguments, true
		}
		if len(arguments) >= 256 {
			return nil, false
		}
		if line[0] != '"' {
			index := strings.IndexAny(line, " \t")
			if index < 0 {
				index = len(line)
			}
			arguments, line = append(arguments, line[:index]), line[index:]
			continue
		}
		end := 0
		for index := 1; index < len(line); index++ {
			if line[index] == '\\' {
				index++
				continue
			}
			if line[index] == '"' {
				end = index + 1
				break
			}
		}
		if end == 0 {
			return nil, false
		}
		argument, err := strconv.Unquote(line[:end])
		if err != nil {
			return nil, false
		}
		line = line[end:]
		if line != "" && line[0] != ' ' && line[0] != '\t' {
			return nil, false
		}
		arguments = append(arguments, argument)
	}
}

func inspectGeneratorArguments(root, directory string, args []string, tracked map[string]bool) generatorDisposition {
	if len(args) < 3 || args[0] != "go" || args[1] != "run" || strings.HasPrefix(args[2], "-") {
		return generatorUnknown
	}
	for _, argument := range args {
		if strings.Contains(argument, "$") {
			return generatorUnknown
		}
	}
	first := args[2]
	if strings.HasSuffix(first, ".go") {
		// go run consumes consecutive .go arguments as source files. Once an
		// application argument is reached, later .go words belong to that program.
		for _, source := range args[2:] {
			if !strings.HasSuffix(source, ".go") {
				break
			}
			if path.IsAbs(source) || strings.Contains(source, "\\") {
				return generatorMissing
			}
			name := path.Clean(path.Join(directory, source))
			if !tracked[name] || repositoryfile.ValidateRegularFile(root, filepath.FromSlash(name)) != nil {
				return generatorMissing
			}
		}
		return generatorAvailable
	}
	if strings.HasPrefix(first, "./") || strings.HasPrefix(first, "../") {
		name := path.Clean(path.Join(directory, first))
		found := false
		for entry := range tracked {
			if strings.HasPrefix(entry, name+"/") {
				found = true
				break
			}
		}
		if !found || repositoryfile.ValidateDirectory(root, filepath.FromSlash(name)) != nil {
			return generatorMissing
		}
		return generatorAvailable
	}
	return generatorUnknown
}
