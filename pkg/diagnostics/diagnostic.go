package diagnostics

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Diagnostic is a single problem found by an operation. Severity and Summary
// are required; Detail, Annotations and Cause return their zero values when
// a diagnostic has none. Callers must not modify the map Annotations returns.
type Diagnostic interface {
	Severity() Severity
	Summary() string
	Detail() string
	Annotations() map[string]any
	Cause() error
}

// String formats any diagnostic on one line, as "<severity>: <summary>",
// followed by ": <detail>" when there is a detail.
func String(d Diagnostic) string {
	s := string(d.Severity()) + ": " + d.Summary()
	if detail := d.Detail(); detail != "" {
		s += ": " + detail
	}
	return s
}

var (
	_ Diagnostic = (*diagnostic)(nil)
)

type diagnostic struct {
	severity    Severity
	summary     string
	detail      string
	annotations map[string]any
	cause       error
}

func (d *diagnostic) Severity() Severity {
	return d.severity
}

func (d *diagnostic) Summary() string {
	return d.summary
}

func (d *diagnostic) Detail() string {
	return d.detail
}

func (d *diagnostic) Annotations() map[string]any {
	return d.annotations
}

func (d *diagnostic) Cause() error {
	return d.cause
}

func (d *diagnostic) String() string {
	return String(d)
}
