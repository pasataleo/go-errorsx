package errorsx

import (
	"errors"
	"testing"
)

func TestAppend_NoErrors(t *testing.T) {
	err := errors.New("base")
	if result := Append(err); result != err { //nolint:errorlint // testing pointer identity
		t.Errorf("Append(err) = %v, want %v", result, err)
	}
}

func TestAppend_NilBase(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	err := Append(nil, e1, e2)
	errs := Errors(err)
	if len(errs) != 2 {
		t.Fatalf("len(errs) = %d, want 2", len(errs))
	}
	if errs[0] != e1 || errs[1] != e2 { //nolint:errorlint // testing pointer identity
		t.Errorf("errs = %v, want [e1, e2]", errs)
	}
}

func TestAppend_NonAggregatedBase(t *testing.T) {
	base := errors.New("base")
	e1 := errors.New("e1")
	err := Append(base, e1)
	errs := Errors(err)
	if len(errs) != 2 {
		t.Fatalf("len(errs) = %d, want 2", len(errs))
	}
	if errs[0] != base || errs[1] != e1 { //nolint:errorlint // testing pointer identity
		t.Errorf("errs = %v, want [base, e1]", errs)
	}
}

func TestAppend_AggregatedBase(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	e3 := errors.New("e3")
	err := Append(nil, e1, e2)
	err = Append(err, e3)
	errs := Errors(err)
	if len(errs) != 3 {
		t.Fatalf("len(errs) = %d, want 3", len(errs))
	}
	if errs[0] != e1 || errs[1] != e2 || errs[2] != e3 { //nolint:errorlint // testing pointer identity
		t.Errorf("errs = %v, want [e1, e2, e3]", errs)
	}
}

func TestAppend_AggregatedInArgs(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	e3 := errors.New("e3")
	aggregated := Append(nil, e2, e3)
	err := Append(e1, aggregated)
	errs := Errors(err)
	if len(errs) != 3 {
		t.Fatalf("len(errs) = %d, want 3", len(errs))
	}
	if errs[0] != e1 || errs[1] != e2 || errs[2] != e3 { //nolint:errorlint // testing pointer identity
		t.Errorf("errs = %v, want [e1, e2, e3]", errs)
	}
}

func TestErrors_NonAggregated(t *testing.T) {
	err := errors.New("single")
	errs := Errors(err)
	if len(errs) != 1 || errs[0] != err { //nolint:errorlint // testing pointer identity
		t.Errorf("Errors(single) = %v, want [single]", errs)
	}
}

func TestAggregated_ErrorMessage(t *testing.T) {
	err := Append(nil, errors.New("a"), errors.New("b"), errors.New("c"))
	want := "a; b; c"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestAggregated_Unwrap(t *testing.T) {
	e1 := errors.New("e1")
	e2 := New(NotFound, nil, "not found")
	err := Append(nil, e1, e2)

	if !errors.Is(err, e1) {
		t.Error("expected errors.Is to find e1 in aggregated error")
	}
	if !errors.Is(err, e2) {
		t.Error("expected errors.Is to find e2 in aggregated error")
	}
}

func TestAggregated_ErrorsAs(t *testing.T) {
	inner := New(NotFound, nil, "not found")
	err := Append(nil, errors.New("plain"), inner)

	var coded CodedError
	if !errors.As(err, &coded) {
		t.Fatal("expected errors.As to find CodedError in aggregated error")
	}
	if coded.ErrorCode() != NotFound {
		t.Errorf("ErrorCode = %q, want %q", coded.ErrorCode(), NotFound)
	}
}
