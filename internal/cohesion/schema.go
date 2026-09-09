package cohesion

import (
	"bytes"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const modulesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules.schema.json"
const catalogSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-catalog.schema.json"
const inputsSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-inputs.schema.json"
const sourcesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/cohesion-sources.schema.json"

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
