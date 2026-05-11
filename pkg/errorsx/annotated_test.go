package errorsx

import (
	"errors"
	"testing"
)

func TestAnnotate_Nil(t *testing.T) {
	if err := Annotate(nil, "key", "value"); err != nil {
		t.Errorf("Annotate(nil) = %v, want nil", err)
	}
}

func TestAnnotate_NewAnnotation(t *testing.T) {
	err := Annotate(errors.New("base"), "key", "value")
	val, ok := GetAnnotation(err, "key")
	if !ok {
		t.Fatal("expected annotation to exist")
	}
	if val != "value" {
		t.Errorf("annotation = %v, want %v", val, "value")
	}
}

func TestAnnotate_ExistingAnnotation(t *testing.T) {
	err := Annotate(errors.New("base"), "k1", "v1")
	err = Annotate(err, "k2", "v2")

	v1, ok := GetAnnotation(err, "k1")
	if !ok || v1 != "v1" {
		t.Errorf("k1 = %v, %v; want v1, true", v1, ok)
	}
	v2, ok := GetAnnotation(err, "k2")
	if !ok || v2 != "v2" {
		t.Errorf("k2 = %v, %v; want v2, true", v2, ok)
	}
}

func TestAnnotate_OverwriteAnnotation(t *testing.T) {
	err := Annotate(errors.New("base"), "key", "v1")
	err = Annotate(err, "key", "v2")
	val, ok := GetAnnotation(err, "key")
	if !ok || val != "v2" {
		t.Errorf("annotation = %v, %v; want v2, true", val, ok)
	}
}

func TestGetAnnotation_Missing(t *testing.T) {
	err := errors.New("plain")
	val, ok := GetAnnotation(err, "key")
	if ok || val != nil {
		t.Errorf("GetAnnotation(plain) = %v, %v; want nil, false", val, ok)
	}
}

func TestGetAnnotations_Nil(t *testing.T) {
	if annotations := GetAnnotations(errors.New("plain")); annotations != nil {
		t.Errorf("GetAnnotations(plain) = %v, want nil", annotations)
	}
}

func TestGetAnnotations(t *testing.T) {
	err := Annotate(errors.New("base"), "k1", "v1")
	err = Annotate(err, "k2", "v2")
	annotations := GetAnnotations(err)
	if len(annotations) != 2 {
		t.Fatalf("len(annotations) = %d, want 2", len(annotations))
	}
	if annotations["k1"] != "v1" || annotations["k2"] != "v2" {
		t.Errorf("annotations = %v, want {k1:v1, k2:v2}", annotations)
	}
}

func TestGetAnnotation_ThroughWrappedError(t *testing.T) {
	inner := Annotate(errors.New("base"), "key", "value")
	outer := Wrap(inner, "wrapped")
	val, ok := GetAnnotation(outer, "key")
	if !ok || val != "value" {
		t.Errorf("GetAnnotation through wrap = %v, %v; want value, true", val, ok)
	}
}

func TestGetAnnotations_ThroughWrappedError(t *testing.T) {
	inner := Annotate(errors.New("base"), "k1", "v1")
	outer := Annotate(Wrap(inner, "wrapped"), "k2", "v2")
	annotations := GetAnnotations(outer)
	if len(annotations) != 2 {
		t.Fatalf("len(annotations) = %d, want 2", len(annotations))
	}
	if annotations["k1"] != "v1" || annotations["k2"] != "v2" {
		t.Errorf("annotations = %v, want {k1:v1, k2:v2}", annotations)
	}
}

func TestGetAnnotations_CloserAnnotationTakesPrecedence(t *testing.T) {
	inner := Annotate(errors.New("base"), "key", "inner")
	outer := Annotate(Wrap(inner, "wrapped"), "key", "outer")
	val, ok := GetAnnotation(outer, "key")
	if !ok || val != "outer" {
		t.Errorf("GetAnnotation precedence = %v, %v; want outer, true", val, ok)
	}
	annotations := GetAnnotations(outer)
	if annotations["key"] != "outer" {
		t.Errorf("GetAnnotations precedence = %v, want outer", annotations["key"])
	}
}

func TestGetAnnotation_ThroughAggregatedError(t *testing.T) {
	inner := Annotate(errors.New("base"), "key", "value")
	err := Append(errors.New("other"), inner)
	val, ok := GetAnnotation(err, "key")
	if !ok || val != "value" {
		t.Errorf("GetAnnotation through aggregate = %v, %v; want value, true", val, ok)
	}
}

func TestAnnotate_Unwrap(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	err := Annotate(inner, "key", "value")
	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to find inner error through annotated error")
	}
}

func TestAnnotate_PreservesCode(t *testing.T) {
	err := Annotate(New(NotFound, nil, "not found"), "key", "value")
	if code := ErrorCode(err); code != NotFound {
		t.Errorf("ErrorCode = %q, want %q", code, NotFound)
	}
}

func TestAnnotate_ErrorMessage(t *testing.T) {
	err := Annotate(errors.New("base"), "key", "value")
	if err.Error() != "base" {
		t.Errorf("Error() = %q, want %q", err.Error(), "base")
	}
}
