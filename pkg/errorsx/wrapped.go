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
	_ CodedError   = (*wrappedError)(nil)
	_ StackTracer  = (*wrappedError)(nil)
	_ StackFramer  = (*wrappedError)(nil)
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

// ErrorCode reports the code of the error that wraps, falling back to the code
// of the cause when the wrapping error carries none.
//
// A wrapped error has two children, and Unwrap can only return one of them: it
// returns the cause, because that is the chain errors.Is and errors.As have to
// walk to find a sentinel. That leaves the error that wraps — the one holding
// the code New was given — off the chain, so without this method
// ErrorCode(New(NotFound, err, "…")) would report the code of err, and a code
// wrapped around a plain error would be lost entirely.
func (w *wrappedError) ErrorCode() Code {
	if coded, ok := w.current.(CodedError); ok { //nolint:errorlint // the wrapping error itself, not its chain
		return coded.ErrorCode()
	}
	return ErrorCode(w.wrapped)
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

// StackFrames mirrors StackTrace for errors carrying already-resolved frames,
// so that a decoded wrapped error still reports the stack captured in the
// process that created it.
func (w *wrappedError) StackFrames() []Frame {
	if frames := stackFrames(w.current); len(frames) > 0 {
		return frames
	}
	if w.wrapped != nil {
		return stackFrames(w.wrapped)
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
