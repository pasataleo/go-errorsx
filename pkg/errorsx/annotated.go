package errorsx

import "fmt"

type AnnotatedError interface {
	error
	Annotate(key string, value interface{}) (interface{}, bool)
	GetAnnotation(key string) (interface{}, bool)
	GetAnnotations() map[string]interface{}
}

func Annotate(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	if annotated, ok := err.(AnnotatedError); ok { //nolint:errorlint // intentionally checking immediate type to avoid mutating a deeply wrapped annotation
		annotated.Annotate(key, value)
		return annotated
	}
	return &annotatedError{
		err: err,
		annotations: map[string]interface{}{
			key: value,
		},
	}
}

func GetAnnotation(err error, key string) (interface{}, bool) {
	if err == nil {
		return nil, false
	}

	if annotated, ok := err.(AnnotatedError); ok { //nolint:errorlint // manually walking the error tree
		if v, found := annotated.GetAnnotation(key); found {
			return v, true
		}
	}

	switch e := err.(type) { //nolint:errorlint // manually walking the error tree
	case interface{ Unwrap() error }:
		return GetAnnotation(e.Unwrap(), key)
	case interface{ Unwrap() []error }:
		for _, child := range e.Unwrap() {
			if v, found := GetAnnotation(child, key); found {
				return v, true
			}
		}
	}

	return nil, false
}

func GetAnnotations(err error) map[string]interface{} {
	result := make(map[string]interface{})
	getAnnotations(err, result)
	if len(result) == 0 {
		return nil
	}
	return result
}

func getAnnotations(err error, result map[string]interface{}) {
	if err == nil {
		return
	}

	// Walk deeper first so that closer annotations take precedence.
	switch e := err.(type) { //nolint:errorlint // manually walking the error tree
	case interface{ Unwrap() error }:
		getAnnotations(e.Unwrap(), result)
	case interface{ Unwrap() []error }:
		for _, child := range e.Unwrap() {
			getAnnotations(child, result)
		}
	}

	if annotated, ok := err.(AnnotatedError); ok { //nolint:errorlint // manually walking the error tree
		for k, v := range annotated.GetAnnotations() {
			result[k] = v
		}
	}
}

var (
	_ AnnotatedError = (*annotatedError)(nil)
)

type annotatedError struct {
	err         error
	annotations map[string]interface{}
}

func (a *annotatedError) Error() string {
	return a.err.Error()
}

func (a *annotatedError) Unwrap() error {
	return a.err
}

func (a *annotatedError) Annotate(key string, value interface{}) (interface{}, bool) {
	existing, ok := a.annotations[key]
	a.annotations[key] = value
	return existing, ok
}

func (a *annotatedError) GetAnnotation(key string) (interface{}, bool) {
	annotation, ok := a.annotations[key]
	return annotation, ok
}

func (a *annotatedError) GetAnnotations() map[string]interface{} {
	return a.annotations
}

func (a *annotatedError) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fmt.Fprintf(f, "%+v", a.err)
		for k, v := range a.annotations {
			fmt.Fprintf(f, " %s=%v", k, v)
		}
		return
	}
	fmt.Fprint(f, a.Error())
}
