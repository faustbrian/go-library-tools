package gates

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOrdinaryBoundedProcessTerminalResult(t *testing.T) {
	childError := errors.New("ordinary child failure")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, cancelDeadline := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancelDeadline()
	for _, test := range []struct {
		name            string
		ctx             context.Context
		child, ctxError error
	}{
		{"uncanceled success", context.Background(), nil, nil},
		{"uncanceled failure", context.Background(), childError, nil},
		{"canceled success", canceled, nil, context.Canceled},
		{"canceled failure", canceled, childError, context.Canceled},
		{"expired success", expired, nil, context.DeadlineExceeded},
		{"expired failure", expired, childError, context.DeadlineExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := boundedProcessResult(test.ctx, "ordinary-child", test.child)
			if test.ctxError == nil && test.child == nil {
				if got != nil {
					t.Fatalf("uncanceled success = %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("terminal failure was reported as success")
			}
			if test.ctxError != nil && !errors.Is(got, test.ctxError) {
				t.Fatalf("terminal error = %v; missing caller classification %v", got, test.ctxError)
			}
			if test.child != nil && !errors.Is(got, test.child) {
				t.Fatalf("terminal error = %v; missing child identity", got)
			}
		})
	}
}
