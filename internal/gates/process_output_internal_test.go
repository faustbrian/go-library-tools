package gates

import "testing"

func TestBoundedProcessOutputNotifiesOnceOutsideLock(t *testing.T) {
	output := &boundedProcessOutput{limit: 3}
	notifications := 0
	output.setOverflowCallback(func() {
		if !output.didOverflow() {
			t.Error("callback observed output before overflow")
		}
		notifications++
	})
	for _, value := range []string{"ab", "c"} {
		if written, err := output.Write([]byte(value)); err != nil || written != len(value) {
			t.Fatalf("accepted write = %d/%v", written, err)
		}
	}
	if output.didOverflow() || notifications != 0 {
		t.Fatal("cumulative exact-limit writes notified overflow")
	}
	for _, value := range []string{"d", "ef"} {
		if written, err := output.Write([]byte(value)); err != nil || written != len(value) {
			t.Fatalf("overflow write = %d/%v", written, err)
		}
	}
	if !output.didOverflow() || notifications != 1 {
		t.Fatalf("overflow = %v, notifications = %d", output.didOverflow(), notifications)
	}
}

func TestBoundedProcessOutputNotifiesLateSubscriber(t *testing.T) {
	output := &boundedProcessOutput{limit: 1}
	if written, err := output.Write([]byte("ab")); err != nil || written != 2 || !output.didOverflow() {
		t.Fatalf("overflow write = %d/%v, overflow = %v", written, err, output.didOverflow())
	}
	// An absent subscriber must not clear a previously observed overflow.
	output.setOverflowCallback(nil)
	notifications := 0
	output.setOverflowCallback(func() {
		if !output.didOverflow() {
			t.Error("late callback lost overflow state")
		}
		notifications++
	})
	if notifications != 1 {
		t.Fatalf("late notifications = %d, want 1", notifications)
	}
	if _, err := output.Write([]byte("c")); err != nil {
		t.Fatal(err)
	}
	if notifications != 1 {
		t.Fatalf("repeated overflow notified late subscriber again: %d", notifications)
	}
}
