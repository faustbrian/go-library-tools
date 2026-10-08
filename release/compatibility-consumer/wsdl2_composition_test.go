package compatibilityconsumer_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	wsdl "github.com/faustbrian/go-wsdl/v2"
	_ "github.com/faustbrian/go-wsdl/v2/builder"
	_ "github.com/faustbrian/go-wsdl/v2/codegen"
	"github.com/faustbrian/go-wsdl/v2/compile"
	_ "github.com/faustbrian/go-wsdl/v2/compose"
	_ "github.com/faustbrian/go-wsdl/v2/diff"
	_ "github.com/faustbrian/go-wsdl/v2/resolve"
)

func TestPublicWSDL2Composition(t *testing.T) {
	source := []byte(`<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" name="Inventory" targetNamespace="urn:inventory"/>`)
	document, err := wsdl.Parse(context.Background(), source, wsdl.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model, ok := document.Definitions11()
	if !ok || model.Name != "Inventory" || model.TargetNamespace != "urn:inventory" {
		t.Fatal("published model mismatch")
	}
	encoded, err := wsdl.Marshal(document, wsdl.MarshalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	copy, err := wsdl.Parse(context.Background(), encoded, wsdl.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	copyModel, ok := copy.Definitions11()
	if !ok || copyModel.Name != model.Name || copyModel.TargetNamespace != model.TargetNamespace {
		t.Fatal("semantic round trip mismatch")
	}
	if result, err := wsdl.Parse(context.Background(), source, wsdl.ParseOptions{MaxDocumentBytes: 1}); result != nil || !errors.Is(err, wsdl.ErrLimitExceeded) {
		t.Fatalf("byte refusal: %v %v", result, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := wsdl.Parse(ctx, source, wsdl.ParseOptions{}); result != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled parse: %v %v", result, err)
	}
	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := compiler.Compile(ctx, compile.Source{URI: "urn:inventory", Content: source}); result != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled compiler: %v %v", result, err)
	}
	if result, err := compiler.Compile(context.Background(), compile.Source{URI: "urn:inventory", Content: source}); result == nil || err != nil {
		t.Fatalf("compiler reuse: %v %v", result, err)
	}
}

func TestPublicWSDL2DefaultErrorPrivacy(t *testing.T) {
	const private = "private-resource-identity"
	source := []byte(`<definitions xmlns="http://schemas.xmlsoap.org/wsdl/" name="&` + private + `;"/>`)
	result, err := wsdl.Parse(context.Background(), source, wsdl.ParseOptions{})
	if result != nil || err == nil {
		t.Fatalf("malformed document admitted: %v %v", result, err)
	}
	if strings.Contains(err.Error(), private) {
		t.Fatal("default error exposes input-derived text")
	}
}
