package diagnostics

import (
	"errors"
	"testing"

	"github.com/pasataleo/go-errorsx/pkg/errorsx"
)

func TestConstructors_Severity(t *testing.T) {
	tests := map[string]struct {
		builder *Builder
		want    Severity
	}{
		"Error":    {Error("summary"), SeverityError},
		"Warning":  {Warning("summary"), SeverityWarning},
		"Info":     {Info("summary"), SeverityInfo},
		"Errorf":   {Errorf("%s", "summary"), SeverityError},
		"Warningf": {Warningf("%s", "summary"), SeverityWarning},
		"Infof":    {Infof("%s", "summary"), SeverityInfo},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := tt.builder.Build()
			if d.Severity() != tt.want {
				t.Errorf("Severity() = %q, want %q", d.Severity(), tt.want)
			}
			if d.Summary() != "summary" {
				t.Errorf("Summary() = %q, want %q", d.Summary(), "summary")
			}
		})
	}
}

func TestBuild_OptionalPartsDefaultToZero(t *testing.T) {
	d := Warning("summary").Build()
	if d.Detail() != "" {
		t.Errorf("Detail() = %q, want empty", d.Detail())
	}
	if d.Annotations() != nil {
		t.Errorf("Annotations() = %v, want nil", d.Annotations())
	}
	if d.Cause() != nil {
		t.Errorf("Cause() = %v, want nil", d.Cause())
	}
}

func TestDetailf(t *testing.T) {
	d := Error("summary").Detailf("used in %q", "Page").Build()
	if d.Detail() != `used in "Page"` {
		t.Errorf("Detail() = %q, want %q", d.Detail(), `used in "Page"`)
	}
}

func TestAnnotate(t *testing.T) {
	d := Error("summary").Annotate("component", "Page").Annotate("line", 12).Build()
	annotations := d.Annotations()
	if len(annotations) != 2 || annotations["component"] != "Page" || annotations["line"] != 12 {
		t.Errorf("Annotations() = %v, want {component:Page, line:12}", annotations)
	}
}

func TestWithAnnotations_CopiesTheMap(t *testing.T) {
	annotations := map[string]any{"component": "Page"}
	d := Error("summary").WithAnnotations(annotations).Build()
	annotations["component"] = "Card"
	annotations["line"] = 12

	got := d.Annotations()
	if len(got) != 1 || got["component"] != "Page" {
		t.Errorf("Annotations() = %v, want {component:Page}", got)
	}
}

func TestCause_KeepsTheOriginalError(t *testing.T) {
	sentinel := errors.New("sentinel")
	err := errorsx.Wrap(errorsx.New(errorsx.NotFound, sentinel, "not found"), "loading")
	d := Error("summary").Cause(err).Build()

	if d.Cause() != err { //nolint:errorlint // the original error must be kept, not an equivalent one
		t.Errorf("Cause() = %v, want %v", d.Cause(), err)
	}
	if !errors.Is(d.Cause(), sentinel) {
		t.Error("errors.Is(Cause(), sentinel) = false, want true")
	}
	if !errorsx.IsCode(d.Cause(), errorsx.NotFound) {
		t.Error("IsCode(Cause(), NotFound) = false, want true")
	}
}

func TestCause_CopiesAnnotationsFromTheWholeTree(t *testing.T) {
	inner := errorsx.Annotate(errors.New("inner"), "line", 12)
	err := errorsx.Annotate(errorsx.Wrap(inner, "outer"), "component", "Page")
	d := Error("summary").Cause(err).Build()

	annotations := d.Annotations()
	if len(annotations) != 2 || annotations["component"] != "Page" || annotations["line"] != 12 {
		t.Errorf("Annotations() = %v, want {component:Page, line:12}", annotations)
	}
}

func TestCause_NeverReplacesBuilderAnnotations(t *testing.T) {
	err := errorsx.Annotate(errors.New("cause"), "component", "from cause")

	before := Error("summary").Cause(err).Annotate("component", "from builder").Build()
	after := Error("summary").Annotate("component", "from builder").Cause(err).Build()

	for name, d := range map[string]Diagnostic{"cause first": before, "cause last": after} {
		if got := d.Annotations()["component"]; got != "from builder" {
			t.Errorf("%s: component = %v, want %q", name, got, "from builder")
		}
	}
}

func TestCause_DoesNotModifyTheErrorsAnnotations(t *testing.T) {
	err := errorsx.Annotate(errors.New("cause"), "component", "Page")
	Error("summary").Cause(err).Annotate("component", "Card").Build()

	if got, _ := errorsx.GetAnnotation(err, "component"); got != "Page" {
		t.Errorf("error's component = %v, want %q", got, "Page")
	}
}

func TestFromError(t *testing.T) {
	err := errorsx.Annotate(errorsx.New(errorsx.InvalidArgument, nil, "bad input"), "field", "name")
	d := FromError(err).Build()

	if d.Severity() != SeverityError {
		t.Errorf("Severity() = %q, want %q", d.Severity(), SeverityError)
	}
	if d.Summary() != "bad input" {
		t.Errorf("Summary() = %q, want %q", d.Summary(), "bad input")
	}
	if d.Cause() != err { //nolint:errorlint // the original error must be kept, not an equivalent one
		t.Errorf("Cause() = %v, want %v", d.Cause(), err)
	}
	if d.Annotations()["field"] != "name" {
		t.Errorf("Annotations() = %v, want {field:name}", d.Annotations())
	}
}

func TestFromError_NilPanics(t *testing.T) {
	expectPanic(t, func() { FromError(nil) })
}

func TestBuild_EmptySummaryPanics(t *testing.T) {
	expectPanic(t, func() { Error("").Build() })
}

func TestBuild_TwicePanics(t *testing.T) {
	b := Error("summary")
	b.Build()
	expectPanic(t, func() { b.Build() })
}

func TestAppend_AfterBuildPanics(t *testing.T) {
	b := Error("summary")
	b.Append(nil)
	expectPanic(t, func() { b.Append(nil) })
}

func TestMethods_AfterBuildPanic(t *testing.T) {
	tests := map[string]func(*Builder){
		"Detail":          func(b *Builder) { b.Detail("detail") },
		"Detailf":         func(b *Builder) { b.Detailf("%s", "detail") },
		"Annotate":        func(b *Builder) { b.Annotate("key", "value") },
		"WithAnnotations": func(b *Builder) { b.WithAnnotations(nil) },
		"Cause":           func(b *Builder) { b.Cause(errors.New("cause")) },
	}
	for name, call := range tests {
		t.Run(name, func(t *testing.T) {
			b := Error("summary")
			b.Build()
			expectPanic(t, func() { call(b) })
		})
	}
}

func TestAppend(t *testing.T) {
	var diags Diagnostics
	diags = Error("first").Append(diags)
	diags = Warning("second").Append(diags)

	if len(diags) != 2 || diags[0].Summary() != "first" || diags[1].Summary() != "second" {
		t.Errorf("diags = %v, want [first second]", diags)
	}
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("expected a panic")
		}
	}()
	fn()
}
