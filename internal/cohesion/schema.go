package cohesion

import (
	"bytes"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const modulesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules.schema.json"
const catalogSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog.schema.json"
const inputsSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-inputs.schema.json"
const sourcesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-sources.schema.json"
const contractReviewV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-contract-review-v1.schema.json"
const schemaV3DecisionReviewV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-v3-decision-review-v1.schema.json"
const contractFreezeV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-contract-freeze-v1.schema.json"
const schemaV3DecisionFreezeV1SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-v3-decision-freeze-v1.schema.json"
const modulesV2SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules-v2.schema.json"
const modulesV3SchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules-v3.schema.json"

var compiledModulesSchema = func() *jsonschema.Schema {
	schemaDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(modulesSchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(modulesSchemaIdentity, schemaDocument))
	return must(compiler.Compile(modulesSchemaIdentity))
}()

var compiledCatalogSchema = func() *jsonschema.Schema {
	modulesDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(modulesSchemaJSON))))
	catalogDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(catalogSchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(modulesSchemaIdentity, modulesDocument))
	must(struct{}{}, compiler.AddResource(catalogSchemaIdentity, catalogDocument))
	return must(compiler.Compile(catalogSchemaIdentity))
}()

var compiledInputsSchema = func() *jsonschema.Schema {
	document := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(inputsSchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(inputsSchemaIdentity, document))
	return must(compiler.Compile(inputsSchemaIdentity))
}()

var compiledSourcesSchema = func() *jsonschema.Schema {
	document := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(sourcesSchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(sourcesSchemaIdentity, document))
	return must(compiler.Compile(sourcesSchemaIdentity))
}()

var compiledSchemaV3BootstrapSchemas = func() map[string]*jsonschema.Schema {
	contractReviewDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(contractReviewV1SchemaJSON))))
	contractFreezeDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(contractFreezeV1SchemaJSON))))
	decisionReviewDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(schemaV3DecisionReviewV1SchemaJSON))))
	decisionFreezeDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(schemaV3DecisionFreezeV1SchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(contractReviewV1SchemaIdentity, contractReviewDocument))
	must(struct{}{}, compiler.AddResource(contractFreezeV1SchemaIdentity, contractFreezeDocument))
	must(struct{}{}, compiler.AddResource(schemaV3DecisionReviewV1SchemaIdentity, decisionReviewDocument))
	must(struct{}{}, compiler.AddResource(schemaV3DecisionFreezeV1SchemaIdentity, decisionFreezeDocument))
	return map[string]*jsonschema.Schema{
		contractReviewV1SchemaIdentity:         must(compiler.Compile(contractReviewV1SchemaIdentity)),
		contractFreezeV1SchemaIdentity:         must(compiler.Compile(contractFreezeV1SchemaIdentity)),
		schemaV3DecisionReviewV1SchemaIdentity: must(compiler.Compile(schemaV3DecisionReviewV1SchemaIdentity)),
		schemaV3DecisionFreezeV1SchemaIdentity: must(compiler.Compile(schemaV3DecisionFreezeV1SchemaIdentity)),
	}
}()

var compiledSchemaV3DecisionReviewSchema = compiledSchemaV3BootstrapSchemas[schemaV3DecisionReviewV1SchemaIdentity]

var compiledManifestDeliverySchema = func() *jsonschema.Schema {
	compiler := jsonschema.NewCompiler()
	for _, schema := range []struct {
		identity string
		document string
	}{
		{modulesV2SchemaIdentity, modulesV2SchemaJSON},
		{modulesV3SchemaIdentity, modulesV3SchemaJSON},
	} {
		document := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(schema.document))))
		must(struct{}{}, compiler.AddResource(schema.identity, document))
	}
	return must(compiler.Compile(modulesV3SchemaIdentity + "#/$defs/delivery"))
}()

var compiledSchemaV3Schemas = func() map[string]*jsonschema.Schema {
	embedded := []struct {
		identity string
		document string
	}{
		{"https://github.com/faustbrian/go-library-tools/schema/modules-v1.schema.json", modulesV1SchemaJSON},
		{modulesV2SchemaIdentity, modulesV2SchemaJSON},
		{modulesV3SchemaIdentity, modulesV3SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v1.schema.json", catalogV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog-v2.schema.json", catalogV2SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-inputs-v1.schema.json", inputsV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-inputs-v2.schema.json", inputsV2SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-sources-v1.schema.json", sourcesV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-sources-v2.schema.json", sourcesV2SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-delivery-evidence-v1.schema.json", deliveryEvidenceV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-authorization-registry-v1.schema.json", authorizationRegistryV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-authorization-record-v1.schema.json", authorizationRecordV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-inventory-v1.schema.json", residualInventoryV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-residual-register-v1.schema.json", residualRegisterV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-gate-policy-v1.schema.json", gatePolicyV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-go-toolchains-v1.schema.json", goToolchainsV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-v1.schema.json", goOfficialDownloadsOracleV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-go-official-downloads-oracle-review-v1.schema.json", goOfficialDownloadsOracleReviewV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-engineering-identities-v1.schema.json", engineeringIdentitiesV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-schema-provenance-v1.schema.json", schemaProvenanceV1SchemaJSON},
		{schemaV3DecisionReviewV1SchemaIdentity, schemaV3DecisionReviewV1SchemaJSON},
		{schemaV3DecisionFreezeV1SchemaIdentity, schemaV3DecisionFreezeV1SchemaJSON},
		{contractFreezeV1SchemaIdentity, contractFreezeV1SchemaJSON},
		{contractReviewV1SchemaIdentity, contractReviewV1SchemaJSON},
		{"https://github.com/faustbrian/go-library-tools/schema/cohesion-diagnostic-v1.schema.json", diagnosticV1SchemaJSON},
	}
	compiler := jsonschema.NewCompiler()
	for _, schema := range embedded {
		document := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(schema.document))))
		must(struct{}{}, compiler.AddResource(schema.identity, document))
	}
	compiled := make(map[string]*jsonschema.Schema, len(embedded))
	for _, schema := range embedded {
		compiled[schema.identity] = must(compiler.Compile(schema.identity))
	}
	return compiled
}()

func validateModulesSchema(data []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return compiledModulesSchema.Validate(document)
}

func validateCatalogSchema(data []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return compiledCatalogSchema.Validate(document)
}

func validateInputsSchema(data []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return compiledInputsSchema.Validate(document)
}

func validateSourcesSchema(data []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return compiledSourcesSchema.Validate(document)
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
