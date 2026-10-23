package diagnostics

import (
	"maps"
	"slices"
)

// Diagnostics are the diagnostics collected from one operation, in the order
// they were added. Two sets merge with the built-in append.
type Diagnostics []Diagnostic

func (diags Diagnostics) HasErrors() bool {
	for _, d := range diags {
		if d.Severity() == SeverityError {
			return true
		}
	}
	return false
}

// Filter returns a new set holding only the diagnostics with one of the given
// severities.
func (diags Diagnostics) Filter(severities ...Severity) Diagnostics {
	var result Diagnostics
	for _, d := range diags {
		if slices.Contains(severities, d.Severity()) {
			result = append(result, d)
		}
	}
	return result
}

// WithAnnotations returns a new set in which every diagnostic also has the
// given annotations. Keys a diagnostic already has are not overwritten, and
// the original diagnostics are not modified.
func (diags Diagnostics) WithAnnotations(annotations map[string]any) Diagnostics {
	if diags == nil {
		return nil
	}

	result := make(Diagnostics, len(diags))
	for i, d := range diags {
		result[i] = &annotatedDiagnostic{
			Diagnostic:  d,
			annotations: merge(annotations, d.Annotations()),
		}
	}
	return result
}

// merge returns a new map holding base overlaid with top, or nil if both are
// empty.
func merge(base, top map[string]any) map[string]any {
	if len(base) == 0 && len(top) == 0 {
		return nil
	}
	result := make(map[string]any, len(base)+len(top))
	maps.Copy(result, base)
	maps.Copy(result, top)
	return result
}

var (
	_ Diagnostic = (*annotatedDiagnostic)(nil)
)

// annotatedDiagnostic adds annotations to a diagnostic, keeping the original
// rather than copying it. Its annotations are merged once, when it is created,
// with the wrapped diagnostic's own taking precedence.
type annotatedDiagnostic struct {
	Diagnostic
	annotations map[string]any
}

func (a *annotatedDiagnostic) Annotations() map[string]any {
	return a.annotations
}

func (a *annotatedDiagnostic) String() string {
	return String(a)
}
