package errorsx

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// roundTrip encodes err, marshals it, unmarshals it, and decodes it again, so
// that the tests exercise the same path a value takes across a process
// boundary rather than just Encode followed by Decode.
func roundTrip(t *testing.T, err error) error {
	t.Helper()

	data, marshalErr := json.Marshal(Encode(err))
	if marshalErr != nil {
		t.Fatalf("marshal: %v", marshalErr)
	}

	var encoded Encoded
	if unmarshalErr := json.Unmarshal(data, &encoded); unmarshalErr != nil {
		t.Fatalf("unmarshal: %v", unmarshalErr)
	}
	return encoded.Decode()
}

func TestEncode_Nil(t *testing.T) {
	if got := Encode(nil); got != nil {
		t.Errorf("Encode(nil) = %v, want nil", got)
	}
}

func TestDecode_Nil(t *testing.T) {
	var encoded *Encoded
	if got := encoded.Decode(); got != nil {
		t.Errorf("(*Encoded)(nil).Decode() = %v, want nil", got)
	}
}

func TestRoundTrip_Coded(t *testing.T) {
	err := New(NotFound, nil, "user not found")
	got := roundTrip(t, err)

	if got.Error() != "user not found" {
		t.Errorf("Error() = %q, want %q", got.Error(), "user not found")
	}
	if code := ErrorCode(got); code != NotFound {
		t.Errorf("ErrorCode() = %q, want %q", code, NotFound)
	}
}

func TestRoundTrip_CustomCode(t *testing.T) {
	const rateLimited Code = "rate_limited"

	got := roundTrip(t, New(rateLimited, nil, "slow down"))
	if code := ErrorCode(got); code != rateLimited {
		t.Errorf("ErrorCode() = %q, want %q", code, rateLimited)
	}
}

// TestRoundTrip_NewWithPlainCause pins the code across a hop for the shape
// that keeps it on the wrapping error alone. Every other round trip here uses
// Wrap, which copies the cause's code, so they would pass with it lost.
func TestRoundTrip_NewWithPlainCause(t *testing.T) {
	err := New(InvalidArgument, errors.New("unexpected end of JSON input"), "decoding request")

	got := roundTrip(t, err)
	if code := ErrorCode(got); code != InvalidArgument {
		t.Errorf("ErrorCode() = %q, want %q", code, InvalidArgument)
	}
	if got.Error() != err.Error() {
		t.Errorf("Error() = %q, want %q", got.Error(), err.Error())
	}
}

// The tree from the README: an annotation over a wrap over a coded error.
func TestRoundTrip_AnnotatedWrappedCoded(t *testing.T) {
	err := Annotate(
		Wrap(
			New(NotFound, nil, "user not found"),
			"loading profile",
		),
		"user_id", 42,
	)

	got := roundTrip(t, err)

	if got.Error() != err.Error() {
		t.Errorf("Error() = %q, want %q", got.Error(), err.Error())
	}
	if !IsCode(got, NotFound) {
		t.Errorf("IsCode(NotFound) = false, want true")
	}
	// Annotation values round-trip as JSON, so 42 comes back as a float64.
	if v, ok := GetAnnotation(got, "user_id"); !ok || v != float64(42) {
		t.Errorf("GetAnnotation(user_id) = %v (%T), %v; want 42 (float64), true", v, v, ok)
	}
	// The wrapped cause must still be reachable, not flattened into the message.
	if inner := Unwrap(got); inner == nil {
		t.Error("expected the decoded tree to preserve its wrapped cause")
	}
}

func TestRoundTrip_Aggregated(t *testing.T) {
	err := Append(nil,
		New(NotFound, nil, "missing name"),
		New(InvalidArgument, nil, "bad email"),
	)

	got := roundTrip(t, err)

	errs := Errors(got)
	if len(errs) != 2 {
		t.Fatalf("Errors() returned %d errors, want 2", len(errs))
	}
	// IsCode searches every branch, so both codes must survive.
	if !IsCode(got, NotFound) {
		t.Error("IsCode(NotFound) = false, want true")
	}
	if !IsCode(got, InvalidArgument) {
		t.Error("IsCode(InvalidArgument) = false, want true")
	}
	if got.Error() != err.Error() {
		t.Errorf("Error() = %q, want %q", got.Error(), err.Error())
	}
}

func TestRoundTrip_ForeignError(t *testing.T) {
	sentinel := errors.New("plain")

	got := roundTrip(t, sentinel)

	if got.Error() != "plain" {
		t.Errorf("Error() = %q, want %q", got.Error(), "plain")
	}
	// Documented caveat: sentinel identity does not survive the round trip.
	if errors.Is(got, sentinel) {
		t.Error("errors.Is matched the original sentinel; identity is not expected to survive")
	}
}

func TestRoundTrip_WrappedForeignError(t *testing.T) {
	err := Wrap(errors.New("disk full"), "writing cache")

	got := roundTrip(t, err)

	if got.Error() != "writing cache (disk full)" {
		t.Errorf("Error() = %q, want %q", got.Error(), "writing cache (disk full)")
	}
}

func TestEncode_WireFormat(t *testing.T) {
	err := Annotate(
		Wrap(
			New(NotFound, nil, "user not found"),
			"loading profile",
		),
		"user_id", 42,
	)

	data, marshalErr := json.Marshal(Encode(err, WithoutStacks()))
	if marshalErr != nil {
		t.Fatalf("marshal: %v", marshalErr)
	}

	want := `{"type":"annotated","annotations":{"user_id":42},` +
		`"cause":{"type":"wrapped",` +
		`"current":{"type":"coded","code":"not_found","message":"loading profile"},` +
		`"cause":{"type":"coded","code":"not_found","message":"user not found"}}}`

	if string(data) != want {
		t.Errorf("wire format changed:\n got: %s\nwant: %s", data, want)
	}
}

func TestEncode_StacksIncludedByDefault(t *testing.T) {
	encoded := Encode(New(NotFound, nil, "user not found"))

	if len(encoded.Stack) == 0 {
		t.Fatal("expected encoded stack frames")
	}
	if !strings.Contains(encoded.Stack[0].File, "json_test.go") {
		t.Errorf("expected top frame in json_test.go, got %s", encoded.Stack[0].File)
	}
}

func TestEncode_WithoutStacks(t *testing.T) {
	encoded := Encode(New(NotFound, nil, "user not found"), WithoutStacks())

	if len(encoded.Stack) != 0 {
		t.Errorf("expected no stack frames, got %d", len(encoded.Stack))
	}
}

func TestRoundTrip_StackSurvives(t *testing.T) {
	got := roundTrip(t, New(NotFound, nil, "user not found"))

	stack := FormatStack(got)
	if stack == "" {
		t.Fatal("expected FormatStack to render the decoded stack")
	}
	if !strings.Contains(stack, "json_test.go") {
		t.Errorf("expected decoded stack to reference json_test.go, got: %s", stack)
	}
}

func TestRoundTrip_StackSurvivesThroughWrap(t *testing.T) {
	inner := New(NotFound, nil, "user not found")
	got := roundTrip(t, Wrap(inner, "loading profile"))

	if stack := FormatStack(got); !strings.Contains(stack, "json_test.go") {
		t.Errorf("expected decoded wrapped stack to reference json_test.go, got: %s", stack)
	}
}

func TestDecode_UnknownType(t *testing.T) {
	encoded := &Encoded{Type: "something_new", Message: "from a newer producer"}

	got := encoded.Decode()
	if got == nil {
		t.Fatal("expected an opaque error, got nil")
	}
	if got.Error() != "from a newer producer" {
		t.Errorf("Error() = %q, want %q", got.Error(), "from a newer producer")
	}
}

func TestDecode_MissingCodeFallsBackToUnknown(t *testing.T) {
	encoded := &Encoded{Type: TypeCoded, Message: "no code"}

	if code := ErrorCode(encoded.Decode()); code != Unknown {
		t.Errorf("ErrorCode() = %q, want %q", code, Unknown)
	}
}

func TestDecode_AnnotatedIsStillAnnotatable(t *testing.T) {
	got := roundTrip(t, Annotate(New(NotFound, nil, "user not found"), "user_id", 42))

	// Decode must not leave a nil annotation map behind; Annotate writes to it
	// in place.
	got = Annotate(got, "request_id", "abc")

	if v, ok := GetAnnotation(got, "request_id"); !ok || v != "abc" {
		t.Errorf("GetAnnotation(request_id) = %v, %v; want abc, true", v, ok)
	}
	if v, ok := GetAnnotation(got, "user_id"); !ok || v != float64(42) {
		t.Errorf("GetAnnotation(user_id) = %v, %v; want 42, true", v, ok)
	}
}

func TestEncode_AnnotationsSnapshot(t *testing.T) {
	err := Annotate(New(NotFound, nil, "user not found"), "user_id", 42)

	encoded := Encode(err, WithoutStacks())
	// Mutating the original error afterwards must not reach into the snapshot.
	err = Annotate(err, "user_id", 99)

	if got := encoded.Annotations["user_id"]; got != 42 {
		t.Errorf("encoded annotation = %v, want 42", got)
	}
	if got, _ := GetAnnotation(err, "user_id"); got != 99 {
		t.Errorf("original annotation = %v, want 99", got)
	}
}

func TestDecode_EmptyAggregated(t *testing.T) {
	encoded := &Encoded{Type: TypeAggregated}

	if got := encoded.Decode(); got != nil {
		t.Errorf("Decode() = %v, want nil for an aggregate with no children", got)
	}
}
