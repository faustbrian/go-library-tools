package cohesion

import (
	"bytes"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const modulesSchemaIdentity = "https://github.com/faustbrian/go-library-tools/schema/modules.schema.json"

var compiledModulesSchema = func() *jsonschema.Schema {
	schemaDocument := must(jsonschema.UnmarshalJSON(bytes.NewReader([]byte(modulesSchemaJSON))))
	compiler := jsonschema.NewCompiler()
	must(struct{}{}, compiler.AddResource(modulesSchemaIdentity, schemaDocument))
	return must(compiler.Compile(modulesSchemaIdentity))
}()

func validateModulesSchema(data []byte) error {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return compiledModulesSchema.Validate(document)
}

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
