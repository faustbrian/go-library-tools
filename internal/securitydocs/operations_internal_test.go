package securitydocs

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// These are private phase-entry assertions with genuinely canceled contexts,
// not attempts to schedule cancellation inside an OS operation or parser.
func TestSecurityAcquiredRootAndRegistryPhasesRespectCancellation(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	var reads []string
	read := securityDocumentReader(func(_ *os.Root, path string) ([]byte, error) {
		reads = append(reads, path)
		if path == "risk-register.json" {
			return []byte(`{"schema_version":1,"risks":[]}`), nil
		}
		return []byte(`{"schema_version":1,"modules":[]}`), nil
	})
	if err := validateRootContext(ctx, root, now, read); !errors.Is(err, context.Canceled) || len(reads) != 0 {
		t.Fatal("canceled acquired-root phase dispatched document work or lost refusal")
	}
	registry := map[string]risk{}
	if err := validateAdmittedRisksContext(ctx, root, registry, read); !errors.Is(err, context.Canceled) || len(reads) != 0 || len(registry) != 0 {
		t.Fatal("canceled admitted-registry phase dispatched matrix work or changed state")
	}
	if _, err := root.Stat("."); err != nil {
		t.Fatal("borrowed-root phase closed its caller-owned root")
	}
	if err := validateRootContext(t.Context(), root, now, read); err != nil || !slices.Equal(reads, []string{"risk-register.json", "security-matrix.json"}) {
		t.Fatal("ordinary acquired-root recovery changed validation or document order")
	}
}

func TestSecurityDecodedRecordPhasesRespectCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	ordinary := risk{ID: "SEC-ORDINARY", Module: "ordinary", Severity: "low", Status: "open", Owner: "owner", Rationale: "ordinary", Mitigation: "ordinary", ReviewCondition: "review"}
	risks := riskRegister{Risks: []risk{ordinary}}
	if admitted, err := validateDecodedRisksContext(ctx, risks, now); admitted != nil || !errors.Is(err, context.Canceled) || !reflect.DeepEqual(risks.Risks, []risk{ordinary}) {
		t.Fatal("canceled decoded-risk admission returned a registry or changed records")
	}
	admitted, err := validateDecodedRisksContext(t.Context(), risks, now)
	if err != nil || !reflect.DeepEqual(admitted, map[string]risk{"SEC-ORDINARY": ordinary}) {
		t.Fatal("ordinary decoded-risk recovery changed registry admission")
	}
	modules := matrix{Modules: []moduleRecord{}}
	if err := validateDecodedMatrixContext(ctx, modules, admitted, "ordinary"); !errors.Is(err, context.Canceled) || len(modules.Modules) != 0 || !reflect.DeepEqual(admitted, map[string]risk{"SEC-ORDINARY": ordinary}) {
		t.Fatal("canceled decoded-matrix admission lost refusal or changed registry")
	}
	if err := validateDecodedMatrixContext(t.Context(), modules, admitted, "ordinary"); err != nil {
		t.Fatal("ordinary decoded-matrix recovery changed admission")
	}
}

func TestSecurityAdmittedStreamRetainsInclusiveAllowanceAndCleanup(t *testing.T) {
	closeFailure := errors.New("ordinary close failure")
	stream := &securityOperationStream{reader: strings.NewReader(`{}`), closeFailure: closeFailure}
	data, err := consumeSecurityDocumentContext(t.Context(), stream, 2)
	if err != nil || string(data) != `{}` || stream.closes != 1 {
		t.Fatal("inclusive stream allowance changed data or historical close-error handling")
	}
	// The lookahead reads only two of three ordinary bytes at allowance one.
	remaining := strings.NewReader("{} ")
	stream = &securityOperationStream{reader: remaining}
	data, err = consumeSecurityDocumentContext(t.Context(), stream, 1)
	if data != nil || err == nil || err.Error() != "security document exceeds 1 bytes" || remaining.Len() != 1 || stream.closes != 1 {
		t.Fatal("retained-stream refusal lost its bound, lookahead or owned cleanup")
	}
	stream = &securityOperationStream{reader: strings.NewReader(`{}`)}
	data, err = consumeSecurityDocumentContext(t.Context(), stream, 2)
	if err != nil || string(data) != `{}` || stream.closes != 1 {
		t.Fatal("ordinary stream recovery retained refused state")
	}
}

func TestSecurityAdmittedStreamPreservesReadSizeCancellationPrecedence(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	failure := errors.New("ordinary read failure")
	cause := &os.PathError{Op: "read", Path: "ordinary", Err: failure}
	stream := &securityOperationStream{reader: securityOperationFailureReader{cause: cause}}
	data, err := consumeSecurityDocumentContext(ctx, stream, 1)
	var retained *os.PathError
	if data != nil || !errors.Is(err, failure) || !errors.As(err, &retained) || retained.Op != "read" || retained.Path != "ordinary" || errors.Is(err, context.Canceled) || stream.closes != 1 {
		t.Fatal("read refusal lost its typed cause, precedence or cleanup")
	}
	stream = &securityOperationStream{reader: strings.NewReader(`{}`)}
	data, err = consumeSecurityDocumentContext(ctx, stream, 1)
	if data != nil || err == nil || err.Error() != "security document exceeds 1 bytes" || errors.Is(err, context.Canceled) || stream.closes != 1 {
		t.Fatal("size refusal lost precedence over terminal cancellation or cleanup")
	}
	// Successful EOF is the actual operation boundary that cancels this context.
	ctx, cancel = context.WithCancel(t.Context())
	defer cancel()
	stream = &securityOperationStream{reader: &securityOperationEOFReader{Reader: strings.NewReader(`{}`), cancel: cancel}}
	data, err = consumeSecurityDocumentContext(ctx, stream, 2)
	if data != nil || !errors.Is(err, context.Canceled) || stream.closes != 1 {
		t.Fatal("completed admitted-stream read published data after cancellation")
	}
}

type securityOperationStream struct {
	reader       io.Reader
	closes       int
	closeFailure error
}

func (stream *securityOperationStream) Read(data []byte) (int, error) {
	return stream.reader.Read(data)
}

func (stream *securityOperationStream) Close() error {
	stream.closes++
	return stream.closeFailure
}

type securityOperationFailureReader struct{ cause error }

func (reader securityOperationFailureReader) Read([]byte) (int, error) {
	return 0, reader.cause
}

type securityOperationEOFReader struct {
	*strings.Reader
	cancel context.CancelFunc
}

func (reader *securityOperationEOFReader) Read(data []byte) (int, error) {
	n, err := reader.Reader.Read(data)
	if errors.Is(err, io.EOF) {
		reader.cancel()
	}
	return n, err
}
