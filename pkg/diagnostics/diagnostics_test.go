package diagnostics

import (
	"fmt"
	"testing"
)

// custom is a diagnostic implemented outside the package.
type custom struct {
	annotations map[string]any
}

func (c *custom) Severity() Severity          { return SeverityWarning }
func (c *custom) Summary() string             { return "custom" }
func (c *custom) Detail() string              { return "" }
func (c *custom) Annotations() map[string]any { return c.annotations }
func (c *custom) Cause() error                { return nil }

func TestString(t *testing.T) {
	tests := map[string]struct {
		diagnostic Diagnostic
		want       string
	}{
		"summary only":    {Warning("unused setting").Build(), "warning: unused setting"},
		"with detail":     {Error("not registered").Detail("used in Page").Build(), "error: not registered: used in Page"},
		"caller-provided": {&custom{}, "warning: custom"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := String(tt.diagnostic); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiagnostic_Stringer(t *testing.T) {
	d := Info("summary").Detail("detail").Build()
	if got := fmt.Sprint(d); got != "info: summary: detail" {
		t.Errorf("fmt.Sprint = %q, want %q", got, "info: summary: detail")
	}
}

func TestHasErrors(t *testing.T) {
	warnings := Diagnostics{Warning("w").Build(), Info("i").Build()}
	if warnings.HasErrors() {
		t.Error("HasErrors() = true for warnings and info, want false")
	}
	if !append(warnings, Error("e").Build()).HasErrors() {
		t.Error("HasErrors() = false with an error, want true")
	}
	if Diagnostics(nil).HasErrors() {
		t.Error("HasErrors() = true for nil, want false")
	}
}

func TestFilter(t *testing.T) {
	diags := Diagnostics{
		Error("e1").Build(),
		Warning("w1").Build(),
		Info("i1").Build(),
		Error("e2").Build(),
	}

	got := diags.Filter(SeverityError, SeverityInfo)
	if summaries := summaries(got); summaries != "[e1 i1 e2]" {
		t.Errorf("Filter(error, info) = %s, want [e1 i1 e2]", summaries)
	}
	if got := diags.Filter(); got != nil {
		t.Errorf("Filter() = %v, want nil", got)
	}
	if len(diags) != 4 {
		t.Errorf("original modified: len = %d, want 4", len(diags))
	}
}

func TestWithAnnotations_AddsToEveryDiagnostic(t *testing.T) {
	diags := Diagnostics{Error("e").Build(), Warning("w").Annotate("line", 12).Build()}
	got := diags.WithAnnotations(map[string]any{"component": "Page"})

	if got[0].Annotations()["component"] != "Page" || got[1].Annotations()["component"] != "Page" {
		t.Errorf("annotations = %v, %v; want component=Page on both", got[0].Annotations(), got[1].Annotations())
	}
	if got[1].Annotations()["line"] != 12 {
		t.Errorf("annotations = %v, want line kept", got[1].Annotations())
	}
}

func TestWithAnnotations_DoesNotOverwriteExistingKeys(t *testing.T) {
	diags := Diagnostics{Error("e").Annotate("component", "Card").Build()}
	got := diags.WithAnnotations(map[string]any{"component": "Page"})

	if got[0].Annotations()["component"] != "Card" {
		t.Errorf("component = %v, want Card", got[0].Annotations()["component"])
	}
}

func TestWithAnnotations_DoesNotModifyTheOriginals(t *testing.T) {
	original := Error("e").Annotate("line", 12).Build()
	diags := Diagnostics{original}
	got := diags.WithAnnotations(map[string]any{"component": "Page"})

	if diags[0] != original {
		t.Error("original set modified")
	}
	if _, ok := original.Annotations()["component"]; ok {
		t.Errorf("original diagnostic modified: %v", original.Annotations())
	}
	if got[0].Summary() != "e" || got[0].Severity() != SeverityError {
		t.Errorf("copy = %v, want error: e", got[0])
	}
}

func TestWithAnnotations_WrapsCallerDiagnostics(t *testing.T) {
	c := &custom{annotations: map[string]any{"component": "Card"}}
	got := Diagnostics{c}.WithAnnotations(map[string]any{"component": "Page", "line": 12})

	if got[0].Summary() != "custom" || got[0].Severity() != SeverityWarning {
		t.Errorf("wrapped = %v, want warning: custom", got[0])
	}
	annotations := got[0].Annotations()
	if annotations["component"] != "Card" || annotations["line"] != 12 {
		t.Errorf("annotations = %v, want {component:Card, line:12}", annotations)
	}
	if _, ok := c.annotations["line"]; ok {
		t.Errorf("caller's diagnostic modified: %v", c.annotations)
	}
}

func TestWithAnnotations_Nil(t *testing.T) {
	if got := Diagnostics(nil).WithAnnotations(map[string]any{"k": "v"}); got != nil {
		t.Errorf("WithAnnotations on nil = %v, want nil", got)
	}
}

func summaries(diags Diagnostics) string {
	s := make([]string, len(diags))
	for i, d := range diags {
		s[i] = d.Summary()
	}
	return fmt.Sprint(s)
}
