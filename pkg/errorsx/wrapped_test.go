package errorsx

import (
	"errors"
	"testing"
)

func TestWrap_Nil(t *testing.T) {
	if err := Wrap(nil, "msg"); err != nil {
		t.Errorf("Wrap(nil) = %v, want nil", err)
	}
}

func TestWrapf_Nil(t *testing.T) {
	if err := Wrapf(nil, "msg %d", 1); err != nil {
		t.Errorf("Wrapf(nil) = %v, want nil", err)
	}
}

func TestWrap_PreservesCode(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	outer := Wrap(inner, "wrapped")
	if code := ErrorCode(outer); code != NotFound {
		t.Errorf("ErrorCode = %q, want %q", code, NotFound)
	}
}

func TestWrap_ErrorMessage(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	outer := Wrap(inner, "wrapped")
	if outer.Error() != "wrapped (not found)" {
		t.Errorf("Error() = %q, want %q", outer.Error(), "wrapped (not found)")
	}
}

func TestWrap_PlainError(t *testing.T) {
	inner := errors.New("plain")
	outer := Wrap(inner, "wrapped")
	if code := ErrorCode(outer); code != Unknown {
		t.Errorf("ErrorCode = %q, want %q", code, Unknown)
	}
	if !errors.Is(outer, inner) {
		t.Error("expected errors.Is to find inner error")
	}
}

func TestUnwrap(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	outer := Wrap(inner, "wrapped")
	if unwrapped := Unwrap(outer); unwrapped != inner { //nolint:errorlint // testing pointer identity
		t.Errorf("Unwrap = %v, want %v", unwrapped, inner)
	}
}

func TestUnwrap_NonWrapped(t *testing.T) {
	err := New(NotFound, nil, "not found")
	if unwrapped := Unwrap(err); unwrapped != nil {
		t.Errorf("Unwrap(non-wrapped) = %v, want nil", unwrapped)
	}
}
