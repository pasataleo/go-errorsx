package errorsx

import (
	"errors"
	"fmt"
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

// An error annotated or wrapped after it was created still reports the stack
// from where it was created.
func TestFormatStack_InsideTheTree(t *testing.T) {
	tcs := map[string]error{
		"annotated":  Annotate(New(Internal, nil, "boom"), "key", "value"),
		"fmt wrap":   fmt.Errorf("context: %w", New(Internal, nil, "boom")),
		"aggregated": Append(errors.New("plain"), New(Internal, nil, "boom")),
	}
	for name, err := range tcs {
		t.Run(name, func(t *testing.T) {
			if s := FormatStack(err); !strings.Contains(s, "stack_test.go") {
				t.Errorf("FormatStack = %q, want the stack from stack_test.go", s)
			}
		})
	}
}

// The outermost stack wins, as Wrap keeps the original call site.
func TestFormatStack_Outermost(t *testing.T) {
	inner := New(Internal, nil, "inner")
	outer := Annotate(fmt.Errorf("outer: %w", inner), "key", "value")
	if FormatStack(outer) != FormatStack(inner) {
		t.Errorf("FormatStack(outer) = %q, want inner's stack %q", FormatStack(outer), FormatStack(inner))
	}

	created := New(NotFound, nil, "created here")
	wrapped := Annotate(New(Internal, created, "wrapped here"), "key", "value")
	if FormatStack(wrapped) == FormatStack(created) {
		t.Error("FormatStack chose the inner stack over the outer one")
	}
}
