package assets

// Spec describes one historical asset generation target.
type Spec struct {
	Base             string
	Commit           string
	SchemaPath       string
	OriginalPath     string
	OriginalSHA256   string
	LoaderPaths      []string
	ParserPaths      []string
	CorpusEntrypoint string
	Runner           string
	RunnerDirectory  string
	RunnerTest       string
	DefaultValueKind string
	RuntimeRunners   []RuntimeRunner
}

// RuntimeRunner describes an additional runtime characterization test.
type RuntimeRunner struct {
	Runner, Directory, Test string
}

// Specs lists the historical asset generation targets.
var Specs = []Spec{
	{
		Base: "modules-v1", Commit: "368b815e2e1b444b388cf405823afd0daf031ab6",
		SchemaPath:       "schema/modules-v1.schema.json",
		LoaderPaths:      []string{"internal/inventory/inventory.go#Load(root string, policy config.Config) (Inventory, error)"},
		CorpusEntrypoint: "internal/inventory/inventory.go#Load(root string, policy config.Config) (Inventory, error)", Runner: "modules_v1_historical_test.go", RunnerDirectory: "internal/inventory", RunnerTest: "TestHistoricalModulesV1ObservationRunner", DefaultValueKind: "json",
	},
	{
		Base: "modules-v2", Commit: "b06e903bfef69efe0b21cad07117555d8373d5ff",
		SchemaPath:   "schema/modules-v2.schema.json",
		OriginalPath: "schema/modules.schema.json", OriginalSHA256: "sha256:a72d03ea77f9134b516ef61426ba7fa1a95fd9505c68d5d2288aee6472b42cd7",
		LoaderPaths: []string{
			"internal/inventory/inventory.go#Load(root string, policy config.Config) (Inventory, error)",
			"internal/inventory/inventory.go#LoadSnapshot(root string, policy config.Config) (Inventory, []byte, error)",
		},
		CorpusEntrypoint: "internal/inventory/inventory.go#LoadSnapshot(root string, policy config.Config) (Inventory, []byte, error)", Runner: "modules_v2_historical_test.go", RunnerDirectory: "internal/inventory", RunnerTest: "TestHistoricalModulesV2ObservationRunner", DefaultValueKind: "json",
	},
	{
		Base: "cohesion-catalog-v1", Commit: "b06e903bfef69efe0b21cad07117555d8373d5ff",
		SchemaPath:   "schema/cohesion-catalog-v1.schema.json",
		OriginalPath: "schema/cohesion-catalog.schema.json", OriginalSHA256: "sha256:674fcd9c3abd6c985de6b681883b4f822a510e89a84dd01e8d6249a56af3af49",
		ParserPaths:      []string{"internal/cohesion/catalog.go#Project(catalog inventory.Inventory, view string, identity Identity) (Envelope, error)"},
		CorpusEntrypoint: "internal/cohesion/catalog.go#Project(catalog inventory.Inventory, view string, identity Identity) (Envelope, error)", Runner: "catalog_v1_historical_runner_test.go", RunnerDirectory: "internal/cohesion", RunnerTest: "TestHistoricalCatalogV1Runner",
		RuntimeRunners: []RuntimeRunner{{Runner: "catalog_v1_markdown_historical_test.go", Directory: "internal/cli", Test: "TestHistoricalCatalogV1MarkdownRuntimeAndCLI"}},
	},
	{
		Base: "cohesion-inputs-v1", Commit: "903eb53b12eb52510e07aaf193e8361c25537606",
		SchemaPath:   "schema/cohesion-inputs-v1.schema.json",
		OriginalPath: "schema/cohesion-inputs.schema.json", OriginalSHA256: "sha256:b8d7977be639117e349156bdd9d37ec475832151e0d56fc5159ea940cf4f5fca",
		LoaderPaths:      []string{"internal/cohesion/aggregate.go#Aggregate(inputsPath string, identity Identity) (Artifacts, error)"},
		CorpusEntrypoint: "internal/cohesion/aggregate.go#Aggregate(inputsPath string, identity Identity) (Artifacts, error)", Runner: "inputs_v1_historical_runner_test.go", RunnerDirectory: "internal/cohesion", RunnerTest: "TestHistoricalInputsV1Runner", DefaultValueKind: "json",
	},
	{
		Base: "cohesion-sources-v1", Commit: "89536ffa95c8d665537c23e05566e1225b0e36a6",
		SchemaPath:   "schema/cohesion-sources-v1.schema.json",
		OriginalPath: "schema/cohesion-sources.schema.json", OriginalSHA256: "sha256:bb86d9335a5362000111ba38aecda145bc3ac0e3329719f003fe4c8b4684288a",
		LoaderPaths:      []string{"internal/cohesion/source_lock.go#LoadSourceLock(path string) (SourceLock, error)"},
		ParserPaths:      []string{"internal/cohesion/source_lock.go#SourceLock.VerifyPolicy(repository, version, checksumsSHA256 string) error"},
		CorpusEntrypoint: "internal/cohesion/source_lock.go#LoadSourceLock(path string) (SourceLock, error)", Runner: "historical_source_lock_runner_test.go", RunnerDirectory: "internal/cohesion", RunnerTest: "TestHistoricalSourceLockObservationRunner", DefaultValueKind: "json",
	},
}

// AssetNames returns the generated filenames for a specification.
func AssetNames(spec Spec) []string {
	return []string{
		spec.Base + "-historical-baseline.json",
		spec.Base + "-accepted-corpus.json",
		spec.Base + "-rejected-corpus.json",
		spec.Base + "-resolved-reference-graph.json",
	}
}
