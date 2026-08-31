package errorsx

import (
	"errors"
	"testing"
)

func TestErrorCode_Nil(t *testing.T) {
	if code := ErrorCode(nil); code != OK {
		t.Errorf("ErrorCode(nil) = %q, want %q", code, OK)
	}
}

func TestErrorCode_Plain(t *testing.T) {
	if code := ErrorCode(errors.New("plain")); code != Unknown {
		t.Errorf("ErrorCode(plain) = %q, want %q", code, Unknown)
	}
}

func TestErrorCode_Coded(t *testing.T) {
	err := New(NotFound, nil, "not found")
	if code := ErrorCode(err); code != NotFound {
		t.Errorf("ErrorCode(coded) = %q, want %q", code, NotFound)
	}
}

func TestErrorCode_Wrapped(t *testing.T) {
	err := Wrap(New(NotFound, nil, "not found"), "wrapped")
	if code := ErrorCode(err); code != NotFound {
		t.Errorf("ErrorCode(wrapped) = %q, want %q", code, NotFound)
	}
}

// TestErrorCode_NewWithPlainCause covers the code New introduces itself. Wrap
// takes its code from the cause, so the cause carries it too and the chain
// finds it either way; a code New put on a plain error exists nowhere but the
// wrapping error, which Unwrap does not return.
func TestErrorCode_NewWithPlainCause(t *testing.T) {
	err := New(InvalidArgument, errors.New("plain"), "decoding request")
	if code := ErrorCode(err); code != InvalidArgument {
		t.Errorf("ErrorCode(new with plain cause) = %q, want %q", code, InvalidArgument)
	}
}

// TestErrorCode_NewWithCodedCause pins which of the two codes wins: ErrorCode
// reports the first it finds, and the wrapping error is the more recent
// judgement about the failure.
func TestErrorCode_NewWithCodedCause(t *testing.T) {
	err := New(FailedPrecondition, New(NotFound, nil, "no such user"), "loading account")
	if code := ErrorCode(err); code != FailedPrecondition {
		t.Errorf("ErrorCode(new with coded cause) = %q, want %q", code, FailedPrecondition)
	}
}

// TestIsCode_NewWithCodedCause is the other half of that: ErrorCode picks one,
// IsCode sees the whole tree.
func TestIsCode_NewWithCodedCause(t *testing.T) {
	err := New(FailedPrecondition, New(NotFound, nil, "no such user"), "loading account")
	for _, code := range []Code{FailedPrecondition, NotFound} {
		if !IsCode(err, code) {
			t.Errorf("IsCode(%q) = false, want true", code)
		}
	}
	if IsCode(err, Internal) {
		t.Error("IsCode(internal) = true, want false")
	}
}

func TestErrorCode_DoubleWrapped(t *testing.T) {
	err := Wrap(Wrap(New(NotFound, nil, "not found"), "first"), "second")
	if code := ErrorCode(err); code != NotFound {
		t.Errorf("ErrorCode(double wrapped) = %q, want %q", code, NotFound)
	}
}

func TestErrorCode_Annotated(t *testing.T) {
	err := Annotate(New(NotFound, nil, "not found"), "key", "value")
	if code := ErrorCode(err); code != NotFound {
		t.Errorf("ErrorCode(annotated) = %q, want %q", code, NotFound)
	}
}

func TestNew_WithoutWrapping(t *testing.T) {
	err := New(NotFound, nil, "not found")
	if err.Error() != "not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "not found")
	}
	if _, ok := err.(*codedError); !ok { //nolint:errorlint // testing concrete type
		t.Errorf("expected *codedError, got %T", err)
	}
}

func TestNew_WithWrapping(t *testing.T) {
	inner := errors.New("inner")
	err := New(NotFound, inner, "outer")
	if _, ok := err.(*wrappedError); !ok { //nolint:errorlint // testing concrete type
		t.Errorf("expected *wrappedError, got %T", err)
	}
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to find inner error")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(NotFound, nil, "item %d not found", 42)
	if err.Error() != "item 42 not found" {
		t.Errorf("Error() = %q, want %q", err.Error(), "item 42 not found")
	}
}
