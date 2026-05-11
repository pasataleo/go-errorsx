package errorsx

import (
	"fmt"
)

type WrappedError interface {
	error
	Unwrap() error
}

func Wrap(err error, format string) error {
	if err == nil {
		return nil
	}
	var stack []uintptr
	if _, ok := err.(StackTracer); !ok {
		stack = captureStack()
	}
	return newError(ErrorCode(err), err, format, stack)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	var stack []uintptr
	if _, ok := err.(StackTracer); !ok {
		stack = captureStack()
	}
	return newError(ErrorCode(err), err, fmt.Sprintf(format, args...), stack)
}

func Unwrap(err error) error {
	if wrapped, ok := err.(WrappedError); ok { //nolint:errorlint // intentionally checking immediate type
		return wrapped.Unwrap()
	}
	return nil
}

var (
	_ WrappedError = (*wrappedError)(nil)
	_ StackTracer  = (*wrappedError)(nil)
)

type wrappedError struct {
	current error
	wrapped error
}

func (w *wrappedError) Error() string {
	if w.wrapped == nil {
		return w.current.Error()
	}
	return fmt.Sprintf("%v (%v)", w.current, w.wrapped)
}

func (w *wrappedError) Unwrap() error {
	return w.wrapped
}

func (w *wrappedError) StackTrace() []uintptr {
	if st, ok := w.current.(StackTracer); ok {
		if frames := st.StackTrace(); len(frames) > 0 {
			return frames
		}
	}
	if st, ok := w.wrapped.(StackTracer); ok {
		return st.StackTrace()
	}
	return nil
}

func (w *wrappedError) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fmt.Fprintf(f, "%+v", w.current)
		if w.wrapped != nil {
			fmt.Fprintf(f, "\n  - %+v", w.wrapped)
		}
		return
	}
	fmt.Fprint(f, w.Error())
}
