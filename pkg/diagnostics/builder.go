package diagnostics

import (
	"fmt"

	"github.com/pasataleo/go-errorsx/pkg/errorsx"
)

// Builder constructs a single diagnostic. Calling Build or Append more than
// once, or any other method after either, panics.
type Builder struct {
	severity Severity
	summary  string
	detail   string
	cause    error
	built    bool

	// Annotations set on the builder are kept apart from those copied from the
	// cause, so that the builder's always win whichever was set first.
	annotations      map[string]any
	causeAnnotations map[string]any
}

func Error(summary string) *Builder {
	return newBuilder(SeverityError, summary)
}

func Warning(summary string) *Builder {
	return newBuilder(SeverityWarning, summary)
}

func Info(summary string) *Builder {
	return newBuilder(SeverityInfo, summary)
}

func Errorf(format string, args ...any) *Builder {
	return newBuilder(SeverityError, fmt.Sprintf(format, args...))
}

func Warningf(format string, args ...any) *Builder {
	return newBuilder(SeverityWarning, fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) *Builder {
	return newBuilder(SeverityInfo, fmt.Sprintf(format, args...))
}

// FromError starts an error diagnostic whose summary is the error's message
// and whose cause is the error itself.
func FromError(err error) *Builder {
	if err == nil {
		panic("diagnostics: FromError called with a nil error")
	}
	return Error(err.Error()).Cause(err)
}

func newBuilder(severity Severity, summary string) *Builder {
	return &Builder{
		severity: severity,
		summary:  summary,
	}
}

func (b *Builder) Detail(detail string) *Builder {
	b.checkNotBuilt()
	b.detail = detail
	return b
}

func (b *Builder) Detailf(format string, args ...any) *Builder {
	return b.Detail(fmt.Sprintf(format, args...))
}

func (b *Builder) Annotate(key string, value any) *Builder {
	b.checkNotBuilt()
	if b.annotations == nil {
		b.annotations = make(map[string]any)
	}
	b.annotations[key] = value
	return b
}

// WithAnnotations adds every entry in annotations. The map is copied, so later
// changes to it don't affect the diagnostic.
func (b *Builder) WithAnnotations(annotations map[string]any) *Builder {
	b.checkNotBuilt()
	for k, v := range annotations {
		b.Annotate(k, v)
	}
	return b
}

// Cause attaches the error behind the problem. The annotations of the whole
// error tree, as errorsx.GetAnnotations returns them, are copied onto the
// diagnostic, but never replace annotations set on the builder.
func (b *Builder) Cause(err error) *Builder {
	b.checkNotBuilt()
	b.cause = err
	b.causeAnnotations = errorsx.GetAnnotations(err)
	return b
}

func (b *Builder) Build() Diagnostic {
	b.checkNotBuilt()
	if b.summary == "" {
		panic("diagnostics: a diagnostic needs a summary")
	}
	b.built = true

	return &diagnostic{
		severity:    b.severity,
		summary:     b.summary,
		detail:      b.detail,
		annotations: merge(b.causeAnnotations, b.annotations),
		cause:       b.cause,
	}
}

// Append builds the diagnostic and appends it to diags, which may be nil.
func (b *Builder) Append(diags Diagnostics) Diagnostics {
	return append(diags, b.Build())
}

func (b *Builder) checkNotBuilt() {
	if b.built {
		panic("diagnostics: builder used after its diagnostic was built")
	}
}
