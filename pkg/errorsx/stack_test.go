package errorsx

import (
	"errors"
	"strings"
	"testing"
)

func TestNew_CapturesStack(t *testing.T) {
	err := New(NotFound, nil, "not found")
	st, ok := err.(StackTracer)
	if !ok {
		t.Fatal("expected error to implement StackTracer")
	}
	if len(st.StackTrace()) == 0 {
		t.Error("expected non-empty stack trace")
	}
}

func TestWrap_ReusesExistingStack(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	innerST := inner.(StackTracer).StackTrace() //nolint:errorlint // testing concrete type

	outer := Wrap(inner, "wrapped")
	outerST := outer.(StackTracer).StackTrace() //nolint:errorlint // testing concrete type

	if len(outerST) == 0 {
		t.Fatal("expected non-empty stack trace on wrapped error")
	}
	if &outerST[0] != &innerST[0] {
		t.Error("expected Wrap to reuse the existing stack trace, not capture a new one")
	}
}

func TestWrap_CapturesStackForPlainError(t *testing.T) {
	inner := errors.New("plain")
	outer := Wrap(inner, "wrapped")
	st, ok := outer.(StackTracer)
	if !ok {
		t.Fatal("expected wrapped plain error to implement StackTracer")
	}
	if len(st.StackTrace()) == 0 {
		t.Error("expected non-empty stack trace when wrapping a plain error")
	}
}

func TestFormatStack_WithStack(t *testing.T) {
	err := New(NotFound, nil, "not found")
	s := FormatStack(err)
	if s == "" {
		t.Error("expected non-empty FormatStack output")
	}
	if !strings.Contains(s, "stack_test.go") {
		t.Errorf("expected stack to reference stack_test.go, got: %s", s)
	}
}

func TestFormatStack_PlainError(t *testing.T) {
	err := errors.New("plain")
	if s := FormatStack(err); s != "" {
		t.Errorf("FormatStack(plain) = %q, want empty string", s)
	}
}

func TestFormatStack_Nil(t *testing.T) {
	if s := FormatStack(nil); s != "" {
		t.Errorf("FormatStack(nil) = %q, want empty string", s)
	}
}
