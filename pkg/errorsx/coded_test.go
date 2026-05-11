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
