package cohesion

import (
	"bytes"
	"cmp"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	maximumAggregateInputsSize      int64 = 1 << 20
	maximumAggregateProjectionSize  int64 = 32 << 20
	maximumAggregateRepositories          = 256
	maximumAggregateProjectionBytes int64 = 256 << 20
	maximumAggregateModules               = 4096
	maximumAggregateArtifactBytes   int64 = 512 << 20
)

type aggregateLimits struct {
	maximumRepositories    int
	maximumProjectionBytes int64
	maximumModules         int
	maximumArtifactBytes   int64
}

func defaultAggregateLimits() aggregateLimits {
	return aggregateLimits{
		maximumRepositories:    maximumAggregateRepositories,
		maximumProjectionBytes: maximumAggregateProjectionBytes,
		maximumModules:         maximumAggregateModules,
		maximumArtifactBytes:   maximumAggregateArtifactBytes,
	}
}

type aggregateInputManifest struct {
	SchemaVersion  int                        `json:"schema_version"`
	DesignLanguage aggregateDesignLanguage    `json:"design_language"`
	Repositories   []aggregateRepositoryInput `json:"repositories"`
}

type aggregateDesignLanguage struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

type aggregateRepositoryInput struct {
	Repository string `json:"repository"`
	Projection string `json:"projection"`
	SHA256     string `json:"sha256"`
}

type aggregateEngineeringEnvelope struct {
	SchemaVersion  int                    `json:"schema_version"`
	View           string                 `json:"view"`
	Scope          string                 `json:"scope"`
	Repository     *string                `json:"repository"`
	DesignLanguage DesignLanguageIdentity `json:"design_language"`
	Tooling        ToolingIdentity        `json:"tooling"`
	Modules        []engineeringModule    `json:"modules"`
}

// Artifacts contains the four deterministic ecosystem catalog projections.
type Artifacts struct {
	ConsumerJSON        []byte
	ConsumerMarkdown    []byte
	EngineeringJSON     []byte
	EngineeringMarkdown []byte
}

var aggregateArtifactNames = []struct {
	name    string
	content func(Artifacts) []byte
}{
	{name: "catalog-consumer.json", content: func(artifacts Artifacts) []byte { return artifacts.ConsumerJSON }},
	{name: "catalog-consumer.md", content: func(artifacts Artifacts) []byte { return artifacts.ConsumerMarkdown }},
	{name: "catalog-engineering.json", content: func(artifacts Artifacts) []byte { return artifacts.EngineeringJSON }},
	{name: "catalog-engineering.md", content: func(artifacts Artifacts) []byte { return artifacts.EngineeringMarkdown }},
}

// GenerateAggregate atomically replaces each canonical catalog artifact after
// every input has been validated and all four outputs have rendered in memory.
func GenerateAggregate(inputsPath, outputDirectory string, identity Identity) error {
	return generateAggregateProtected(inputsPath, outputDirectory, "", identity, nil)
}

// GenerateAggregateProtected publishes through a stable output-directory
// handle and rejects a protected directory identity before writing.
func GenerateAggregateProtected(inputsPath, outputDirectory, protectedDirectory string, identity Identity) error {
	return generateAggregateProtected(inputsPath, outputDirectory, protectedDirectory, identity, nil)
}

func generateAggregateProtected(inputsPath, outputDirectory, protectedDirectory string, identity Identity, afterOpen func() error) error {
	operations := protectedAggregateOperations{
		mkdirAll:  os.MkdirAll,
		openRoot:  os.OpenRoot,
		statRoot:  func(root *os.Root) (os.FileInfo, error) { return root.Stat(".") },
		random:    rand.Reader,
		afterOpen: afterOpen,
	}
	return generateAggregateProtectedWithOperations(inputsPath, outputDirectory, protectedDirectory, identity, operations)
}

type protectedAggregateOperations struct {
	mkdirAll  func(string, os.FileMode) error
	openRoot  func(string) (*os.Root, error)
	statRoot  func(*os.Root) (os.FileInfo, error)
	random    io.Reader
	afterOpen func() error
}

func generateAggregateProtectedWithOperations(inputsPath, outputDirectory, protectedDirectory string, identity Identity, operations protectedAggregateOperations) error {
	artifacts, err := Aggregate(inputsPath, identity)
	if err != nil {
		return err
	}
	if err := operations.mkdirAll(outputDirectory, 0o755); err != nil {
		return fmt.Errorf("create cohesion aggregation output: %w", err)
	}
	root, err := operations.openRoot(outputDirectory)
	if err != nil {
		return fmt.Errorf("open cohesion aggregation output: %w", err)
	}
	defer root.Close()
	if operations.afterOpen != nil {
		if err := operations.afterOpen(); err != nil {
			return err
		}
	}
	if protectedDirectory != "" {
		protected, err := operations.openRoot(protectedDirectory)
		if err == nil {
			defer protected.Close()
			outputInfo, outputError := operations.statRoot(root)
			protectedInfo, protectedError := operations.statRoot(protected)
			if outputError != nil || protectedError != nil {
				return errors.Join(outputError, protectedError)
			}
			if os.SameFile(outputInfo, protectedInfo) {
				return errors.New("source build cannot publish cohesion catalogs; use a released checksummed binary")
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("open protected cohesion output: %w", err)
		}
	}
	files := &rootedCatalogFiles{root: root, directory: outputDirectory, random: operations.random}
	return publishAggregateArtifacts(artifacts, outputDirectory, files)
}

func generateAggregateWithFiles(inputsPath, outputDirectory string, identity Identity, files catalogFiles) error {
	artifacts, err := Aggregate(inputsPath, identity)
	if err != nil {
		return err
	}
	return publishAggregateArtifacts(artifacts, outputDirectory, files)
}

func publishAggregateArtifacts(artifacts Artifacts, outputDirectory string, files catalogFiles) error {
	if err := files.MkdirAll(outputDirectory, 0o755); err != nil {
		return fmt.Errorf("create cohesion aggregation output: %w", err)
	}
	staged := make([]stagedCatalog, 0, len(aggregateArtifactNames))
	for _, artifact := range aggregateArtifactNames {
		target := filepath.Join(outputDirectory, artifact.name)
		info, err := files.Lstat(target)
		if err == nil && !info.Mode().IsRegular() {
			cleanupStagedCatalogs(staged, files)
			return fmt.Errorf("publish cohesion catalog: %s must be a regular file", artifact.name)
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupStagedCatalogs(staged, files)
			return err
		}
		entry, err := stageCatalog(target, artifact.content(artifacts), err == nil, files)
		if err != nil {
			cleanupStagedCatalogs(staged, files)
			return err
		}
		staged = append(staged, entry)
	}
	if err := publishCatalogSet(staged, files); err != nil {
		return err
	}
	return nil
}

// CheckAggregate renders every artifact in memory and reports any byte-level
// difference without modifying the output directory.
func CheckAggregate(inputsPath, outputDirectory string, identity Identity) error {
	artifacts, err := Aggregate(inputsPath, identity)
	if err != nil {
		return err
	}
	return checkAggregateArtifacts(outputDirectory, artifacts, maximumAggregateArtifactBytes, readBoundedAggregateFile)
}

func checkAggregateArtifacts(outputDirectory string, artifacts Artifacts, maximum int64, read func(string, int64, string) ([]byte, error)) error {
	stale := make([]string, 0)
	for _, artifact := range aggregateArtifactNames {
		actual, err := read(filepath.Join(outputDirectory, artifact.name), maximum, "cohesion catalog artifact")
		if err != nil || !bytes.Equal(actual, artifact.content(artifacts)) {
			stale = append(stale, artifact.name)
		}
	}
	if len(stale) != 0 {
		return fmt.Errorf("cohesion aggregate artifacts are stale: %s", strings.Join(stale, ", "))
	}
	return nil
}

// Aggregate validates repository engineering projections and derives both
// ecosystem catalog views in memory.
func Aggregate(inputsPath string, identity Identity) (Artifacts, error) {
	return aggregateWithOperations(inputsPath, identity, defaultAggregateOperations())
}

type aggregateOperations struct {
	limits             aggregateLimits
	read               func(string, int64, string) ([]byte, error)
	readProjection     func(string, string, int64, string) ([]byte, error)
	validateInputs     func([]byte) error
	decodeInputs       func([]byte, *aggregateInputManifest) error
	validateCatalog    func([]byte) error
	decodeProjection   func([]byte, *aggregateEngineeringEnvelope) error
	validateProjection func(aggregateRepositoryInput, aggregateEngineeringEnvelope, Identity) error
	marshal            func(Envelope) ([]byte, error)
	render             func(Envelope) ([]byte, error)
}

func defaultAggregateOperations() aggregateOperations {
	return aggregateOperations{
		limits: defaultAggregateLimits(),
		read:   readBoundedAggregateFile, readProjection: readBoundedAggregateFileInRoot,
		validateInputs:     validateInputsSchema,
		decodeInputs:       func(data []byte, value *aggregateInputManifest) error { return json.Unmarshal(data, value) },
		validateCatalog:    validateCatalogSchema,
		decodeProjection:   func(data []byte, value *aggregateEngineeringEnvelope) error { return json.Unmarshal(data, value) },
		validateProjection: validateAggregateProjection,
		marshal:            marshalCatalog, render: RenderMarkdown,
	}
}

func aggregateWithOperations(inputsPath string, identity Identity, operations aggregateOperations) (Artifacts, error) {
	if err := validateIdentity(identity); err != nil {
		return Artifacts{}, err
	}
	data, err := operations.read(inputsPath, maximumAggregateInputsSize, "cohesion aggregation inputs")
	if err != nil {
		return Artifacts{}, fmt.Errorf("read cohesion aggregation inputs: %w", err)
	}
	if err := operations.validateInputs(data); err != nil {
		return Artifacts{}, fmt.Errorf("cohesion aggregation input schema is invalid: %w", err)
	}
	var inputs aggregateInputManifest
	if err := operations.decodeInputs(data, &inputs); err != nil {
		return Artifacts{}, fmt.Errorf("decode cohesion aggregation inputs: %w", err)
	}
	if inputs.SchemaVersion != 1 {
		return Artifacts{}, errors.New("cohesion aggregation inputs require schema_version 1")
	}
	if inputs.DesignLanguage.Version != identity.DesignLanguageVersion || inputs.DesignLanguage.SHA256 != identity.DesignLanguageSHA256 {
		return Artifacts{}, errors.New("cohesion aggregation design-language identity does not match the generator")
	}
	if len(inputs.Repositories) == 0 {
		return Artifacts{}, errors.New("cohesion aggregation inputs require at least one repository")
	}
	if len(inputs.Repositories) > operations.limits.maximumRepositories {
		return Artifacts{}, fmt.Errorf("cohesion aggregation inputs exceed maximum of %d repositories", operations.limits.maximumRepositories)
	}
	for index, input := range inputs.Repositories {
		if input.Repository == "" {
			return Artifacts{}, errors.New("cohesion aggregation repository identity must not be empty")
		}
		if index > 0 && strings.Compare(inputs.Repositories[index-1].Repository, input.Repository) >= 0 {
			return Artifacts{}, errors.New("cohesion aggregation repositories must be unique and sorted")
		}
	}

	modules := make([]engineeringModule, 0)
	seenModules := make(map[string]struct{})
	manifestDirectory := filepath.Dir(inputsPath)
	var projectionBytes int64
	for _, input := range inputs.Repositories {
		if !safeRelativePath(input.Projection) {
			return Artifacts{}, fmt.Errorf("cohesion projection for %s must use a safe manifest-relative path", input.Repository)
		}
		projectionData, err := operations.readProjection(manifestDirectory, input.Projection, maximumAggregateProjectionSize, "cohesion projection")
		if err != nil {
			return Artifacts{}, fmt.Errorf("read cohesion projection for %s: %w", input.Repository, err)
		}
		if int64(len(projectionData)) > operations.limits.maximumProjectionBytes-projectionBytes {
			return Artifacts{}, fmt.Errorf("cohesion aggregation projections exceed maximum total size of %d bytes", operations.limits.maximumProjectionBytes)
		}
		projectionBytes += int64(len(projectionData))
		digest := sha256.Sum256(projectionData)
		if hex.EncodeToString(digest[:]) != input.SHA256 {
			return Artifacts{}, fmt.Errorf("cohesion projection digest does not match for %s", input.Repository)
		}
		if err := operations.validateCatalog(projectionData); err != nil {
			return Artifacts{}, fmt.Errorf("cohesion projection schema is invalid for %s: %w", input.Repository, err)
		}
		var projection aggregateEngineeringEnvelope
		if err := operations.decodeProjection(projectionData, &projection); err != nil {
			return Artifacts{}, fmt.Errorf("decode cohesion projection for %s: %w", input.Repository, err)
		}
		if err := operations.validateProjection(input, projection, identity); err != nil {
			return Artifacts{}, err
		}
		if len(projection.Modules) == 0 {
			return Artifacts{}, fmt.Errorf("cohesion projection for %s contains no modules", input.Repository)
		}
		if len(projection.Modules) > operations.limits.maximumModules-len(modules) {
			return Artifacts{}, fmt.Errorf("cohesion aggregation projections exceed maximum of %d modules", operations.limits.maximumModules)
		}
		for _, module := range projection.Modules {
			if module.Repository != input.Repository {
				return Artifacts{}, fmt.Errorf("cohesion module repository identity does not match for %s", module.ModulePath)
			}
			if module.Releasable && module.ModulePath != input.Repository && !strings.HasPrefix(module.ModulePath, input.Repository+"/") {
				return Artifacts{}, fmt.Errorf("cohesion module path identity does not match repository %s", input.Repository)
			}
			if _, exists := seenModules[module.ModulePath]; exists {
				return Artifacts{}, fmt.Errorf("duplicate cohesion module %s", module.ModulePath)
			}
			seenModules[module.ModulePath] = struct{}{}
			modules = append(modules, module)
		}
	}
	slices.SortFunc(modules, compareEngineeringModules)

	base := Envelope{
		SchemaVersion: 1,
		Scope:         "ecosystem",
		Repository:    nil,
		DesignLanguage: DesignLanguageIdentity{
			Version:        identity.DesignLanguageVersion,
			SHA256:         identity.DesignLanguageSHA256,
			SourceIdentity: identity.SourceIdentity,
		},
		Tooling: ToolingIdentity{Version: identity.ToolingVersion, PublicationStatus: identity.PublicationStatus},
	}
	engineering := base
	engineering.View = "engineering"
	engineering.Modules = modules
	consumer := base
	consumer.View = "consumer"
	consumer.Modules = aggregateConsumerModules(modules)

	consumerJSON, err := operations.marshal(consumer)
	if err != nil {
		return Artifacts{}, err
	}
	engineeringJSON, err := operations.marshal(engineering)
	if err != nil {
		return Artifacts{}, err
	}
	consumerMarkdown, err := operations.render(consumer)
	if err != nil {
		return Artifacts{}, err
	}
	engineeringMarkdown, err := operations.render(engineering)
	if err != nil {
		return Artifacts{}, err
	}
	artifacts := Artifacts{
		ConsumerJSON: consumerJSON, ConsumerMarkdown: consumerMarkdown,
		EngineeringJSON: engineeringJSON, EngineeringMarkdown: engineeringMarkdown,
	}
	for _, artifact := range aggregateArtifactNames {
		if int64(len(artifact.content(artifacts))) > operations.limits.maximumArtifactBytes {
			return Artifacts{}, fmt.Errorf("cohesion aggregate artifact %s exceeds maximum size of %d bytes", artifact.name, operations.limits.maximumArtifactBytes)
		}
	}
	return artifacts, nil
}

func readBoundedAggregateFile(path string, maximum int64, label string) ([]byte, error) {
	return readBoundedAggregateFileWithFiles(path, maximum, label, operatingAggregateInputFiles{})
}

type aggregateReadCloser interface {
	io.Reader
	Close() error
	Stat() (os.FileInfo, error)
}

type aggregateInputFiles interface {
	Lstat(string) (os.FileInfo, error)
	Open(string) (aggregateReadCloser, error)
	SameFile(os.FileInfo, os.FileInfo) bool
}

type operatingAggregateInputFiles struct{}

func (operatingAggregateInputFiles) Lstat(path string) (os.FileInfo, error) { return os.Lstat(path) }
func (operatingAggregateInputFiles) Open(path string) (aggregateReadCloser, error) {
	return os.Open(path)
}
func (operatingAggregateInputFiles) SameFile(left, right os.FileInfo) bool {
	return os.SameFile(left, right)
}

type rootedAggregateInputFiles struct{ root *os.Root }

func (files rootedAggregateInputFiles) Lstat(path string) (os.FileInfo, error) {
	return files.root.Lstat(path)
}
func (files rootedAggregateInputFiles) Open(path string) (aggregateReadCloser, error) {
	return files.root.Open(path)
}
func (rootedAggregateInputFiles) SameFile(left, right os.FileInfo) bool {
	return os.SameFile(left, right)
}

func readBoundedAggregateFileInRoot(directory, path string, maximum int64, label string) ([]byte, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return readBoundedAggregateFileWithFiles(path, maximum, label, rootedAggregateInputFiles{root: root})
}

func readBoundedAggregateFileWithFiles(path string, maximum int64, label string, files aggregateInputFiles) ([]byte, error) {
	info, err := files.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be a regular non-symlink file", label)
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("%s exceed maximum size", label)
	}
	file, err := files.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !openedInfo.Mode().IsRegular() || !files.SameFile(info, openedInfo) {
		return nil, fmt.Errorf("%s changed while it was being opened", label)
	}
	if openedInfo.Size() > maximum {
		return nil, fmt.Errorf("%s exceed maximum size", label)
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("%s exceed maximum size", label)
	}
	return data, nil
}

func validateAggregateProjection(input aggregateRepositoryInput, projection aggregateEngineeringEnvelope, identity Identity) error {
	if projection.SchemaVersion != 1 || projection.View != "engineering" || projection.Scope != "repository" || projection.Repository == nil || *projection.Repository != input.Repository {
		return fmt.Errorf("cohesion projection identity does not match for %s", input.Repository)
	}
	if projection.DesignLanguage.Version != identity.DesignLanguageVersion ||
		projection.DesignLanguage.SHA256 != identity.DesignLanguageSHA256 ||
		projection.DesignLanguage.SourceIdentity != identity.SourceIdentity ||
		projection.Tooling.Version != identity.ToolingVersion ||
		projection.Tooling.PublicationStatus != identity.PublicationStatus {
		return fmt.Errorf("cohesion projection generator identity does not match for %s", input.Repository)
	}
	return nil
}

func compareEngineeringModules(left, right engineeringModule) int {
	leftOrder := len(familyOrder)
	rightOrder := len(familyOrder)
	if left.Cohesion != nil {
		if order, exists := familyOrder[left.Cohesion.Family]; exists {
			leftOrder = order
		}
	}
	if right.Cohesion != nil {
		if order, exists := familyOrder[right.Cohesion.Family]; exists {
			rightOrder = order
		}
	}
	if leftOrder != rightOrder {
		return cmp.Compare(leftOrder, rightOrder)
	}
	return strings.Compare(left.ModulePath, right.ModulePath)
}

func aggregateConsumerModules(modules []engineeringModule) []consumerModule {
	consumer := make([]consumerModule, 0, len(modules))
	for _, module := range modules {
		if !module.Releasable || module.Cohesion == nil ||
			(module.Kind != "public library" && module.Kind != "adapter") ||
			(module.Cohesion.LifecycleStatus != "active" && module.Cohesion.LifecycleStatus != "deprecated") {
			continue
		}
		consumer = append(consumer, consumerModule{
			Repository: module.Repository, Directory: module.Directory,
			ModulePath: module.ModulePath, GoVersion: module.GoVersion, Kind: module.Kind,
			Releasable: module.Releasable, Version: module.Version,
			Specifications: module.Specifications, OwnedDependencies: module.OwnedDependencies,
			Cohesion: module.Cohesion,
		})
	}
	return consumer
}

func marshalCatalog(envelope Envelope) ([]byte, error) {
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode cohesion catalog: %w", err)
	}
	return append(data, '\n'), nil
}

type stagedCatalog struct {
	target    string
	temporary string
	backup    string
	existed   bool
	published bool
}

type catalogFile interface {
	Name() string
	Chmod(os.FileMode) error
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

type catalogFiles interface {
	MkdirAll(string, os.FileMode) error
	Lstat(string) (os.FileInfo, error)
	CreateTemp(string, string) (catalogFile, error)
	Rename(string, string) error
	Remove(string) error
}

type rootedCatalogFiles struct {
	root      *os.Root
	directory string
	random    io.Reader
}

func (files *rootedCatalogFiles) relative(path string) (string, error) {
	return rootedCatalogRelative(files.directory, path, filepath.Rel)
}

func rootedCatalogRelative(directory, path string, relativePath func(string, string) (string, error)) (string, error) {
	relative, err := relativePath(directory, path)
	if err != nil {
		return "", err
	}
	if filepath.Dir(relative) != "." {
		return "", fmt.Errorf("cohesion catalog path escapes output directory: %s", filepath.Base(path))
	}
	if !safeRelativePath(filepath.ToSlash(relative)) {
		return "", fmt.Errorf("cohesion catalog path escapes output directory: %s", filepath.Base(path))
	}
	return filepath.ToSlash(relative), nil
}

func (files *rootedCatalogFiles) MkdirAll(path string, _ os.FileMode) error {
	if filepath.Clean(path) != filepath.Clean(files.directory) {
		return errors.New("cohesion catalog output directory changed")
	}
	return nil
}

func (files *rootedCatalogFiles) Lstat(path string) (os.FileInfo, error) {
	relative, err := files.relative(path)
	if err != nil {
		return nil, err
	}
	return files.root.Lstat(relative)
}

func (files *rootedCatalogFiles) CreateTemp(directory, pattern string) (catalogFile, error) {
	if filepath.Clean(directory) != filepath.Clean(files.directory) {
		return nil, errors.New("temporary cohesion catalog directory changed")
	}
	prefix := strings.TrimSuffix(pattern, "*")
	for range 100 {
		suffix := make([]byte, 16)
		if _, err := io.ReadFull(files.random, suffix); err != nil {
			return nil, err
		}
		name := prefix + hex.EncodeToString(suffix)
		file, err := files.root.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return &rootedCatalogFile{File: file, name: filepath.Join(files.directory, name)}, nil
	}
	return nil, errors.New("create unique temporary cohesion catalog")
}

func (files *rootedCatalogFiles) Rename(oldPath, newPath string) error {
	oldRelative, err := files.relative(oldPath)
	if err != nil {
		return err
	}
	newRelative, err := files.relative(newPath)
	if err != nil {
		return err
	}
	return files.root.Rename(oldRelative, newRelative)
}

func (files *rootedCatalogFiles) Remove(path string) error {
	relative, err := files.relative(path)
	if err != nil {
		return err
	}
	return files.root.Remove(relative)
}

type rootedCatalogFile struct {
	*os.File
	name string
}

func (file *rootedCatalogFile) Name() string { return file.name }

func stageCatalog(target string, content []byte, existed bool, files catalogFiles) (entry stagedCatalog, returnError error) {
	temporary, err := files.CreateTemp(filepath.Dir(target), ".golib-cohesion-*")
	if err != nil {
		return stagedCatalog{}, fmt.Errorf("create temporary cohesion catalog: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if returnError != nil {
			_ = files.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return stagedCatalog{}, fmt.Errorf("set cohesion catalog permissions: %w", err)
	}
	written, err := temporary.Write(content)
	if err != nil {
		_ = temporary.Close()
		return stagedCatalog{}, fmt.Errorf("write cohesion catalog: %w", err)
	}
	if written != len(content) {
		_ = temporary.Close()
		return stagedCatalog{}, fmt.Errorf("write cohesion catalog: %w", io.ErrShortWrite)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return stagedCatalog{}, fmt.Errorf("sync cohesion catalog: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return stagedCatalog{}, fmt.Errorf("close cohesion catalog: %w", err)
	}
	return stagedCatalog{target: target, temporary: temporaryPath, existed: existed}, nil
}

func publishCatalogSet(staged []stagedCatalog, files catalogFiles) error {
	for index := range staged {
		if !staged[index].existed {
			continue
		}
		backup, err := files.CreateTemp(filepath.Dir(staged[index].target), ".golib-cohesion-backup-*")
		if err != nil {
			return rollbackCatalogSet(staged, files, fmt.Errorf("reserve cohesion catalog backup: %w", err))
		}
		backupPath := backup.Name()
		if err := backup.Close(); err != nil {
			_ = files.Remove(backupPath)
			return rollbackCatalogSet(staged, files, fmt.Errorf("close cohesion catalog backup: %w", err))
		}
		if err := files.Remove(backupPath); err != nil {
			return rollbackCatalogSet(staged, files, fmt.Errorf("prepare cohesion catalog backup: %w", err))
		}
		if err := files.Rename(staged[index].target, backupPath); err != nil {
			return rollbackCatalogSet(staged, files, fmt.Errorf("back up cohesion catalog: %w", err))
		}
		staged[index].backup = backupPath
	}
	for index := range staged {
		if err := files.Rename(staged[index].temporary, staged[index].target); err != nil {
			return rollbackCatalogSet(staged, files, fmt.Errorf("publish cohesion catalog set: %w", err))
		}
		staged[index].temporary = ""
		staged[index].published = true
	}
	for index := range staged {
		if staged[index].backup != "" {
			if err := files.Remove(staged[index].backup); err != nil {
				return fmt.Errorf("remove cohesion catalog backup: %w", err)
			}
			staged[index].backup = ""
		}
	}
	return nil
}

func rollbackCatalogSet(staged []stagedCatalog, files catalogFiles, cause error) error {
	rollbackErrors := []error{cause}
	for index := len(staged) - 1; index >= 0; index-- {
		if staged[index].published {
			if err := files.Remove(staged[index].target); err != nil && !errors.Is(err, os.ErrNotExist) {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("remove partially published cohesion catalog: %w", err))
			}
		}
		if staged[index].backup != "" {
			if err := files.Rename(staged[index].backup, staged[index].target); err != nil {
				rollbackErrors = append(rollbackErrors, fmt.Errorf("restore cohesion catalog backup: %w", err))
			} else {
				staged[index].backup = ""
			}
		}
	}
	cleanupStagedCatalogs(staged, files)
	return errors.Join(rollbackErrors...)
}

func cleanupStagedCatalogs(staged []stagedCatalog, files catalogFiles) {
	for _, entry := range staged {
		if entry.temporary != "" {
			_ = files.Remove(entry.temporary)
		}
	}
}
