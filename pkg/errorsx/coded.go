package errorsx

import (
	"errors"
	"fmt"
)

type Code string

const (
	OK                 Code = "ok"                  // 200
	InvalidArgument    Code = "invalid_argument"    // 400
	Unauthenticated    Code = "unauthenticated"     // 401
	PermissionDenied   Code = "permission_denied"   // 403
	NotFound           Code = "not_found"           // 404
	AlreadyExists      Code = "already_exists"      // 409
	FailedPrecondition Code = "failed_precondition" // 412
	ResourceExhausted  Code = "resource_exhausted"  // 429
	Internal           Code = "internal"            // 500
	Unknown            Code = "unknown"             // 500
	Unimplemented      Code = "unimplemented"       // 501
	Unavailable        Code = "unavailable"         // 503
)

func ErrorCode(err error) Code {
	if err == nil {
		return OK
	}
	var coded CodedError
	if errors.As(err, &coded) {
		return coded.ErrorCode()
	}
	return Unknown
}

// IsCode reports whether any error in err's chain has the given code.
// Unlike ErrorCode, which returns the first code found, IsCode checks
// every error including all children of aggregated errors.
func IsCode(err error, code Code) bool {
	if err == nil {
		return code == OK
	}

	if coded, ok := err.(CodedError); ok && coded.ErrorCode() == code { //nolint:errorlint // checking current node only; tree walk is manual
		return true
	}

	switch e := err.(type) { //nolint:errorlint // manually walking the error tree
	case interface{ Unwrap() error }:
		return IsCode(e.Unwrap(), code)
	case interface{ Unwrap() []error }:
		for _, child := range e.Unwrap() {
			if IsCode(child, code) {
				return true
			}
		}
	}

	return false
}

func New(code Code, err error, format string) error {
	return newError(code, err, format, captureStack())
}

func Newf(code Code, err error, format string, args ...any) error {
	return newError(code, err, fmt.Sprintf(format, args...), captureStack())
}

func newError(code Code, err error, format string, stack []uintptr) error {
	current := &codedError{
		code:  code,
		err:   errors.New(format),
		stack: stack,
	}
	if err != nil {
		return &wrappedError{
			current: current,
			wrapped: err,
		}
	}
	return current
}

type CodedError interface {
	error
	ErrorCode() Code
}

var (
	_ CodedError  = (*codedError)(nil)
	_ StackTracer = (*codedError)(nil)
)

type codedError struct {
	code  Code
	err   error
	stack []uintptr
}

func (c *codedError) Error() string {
	return c.err.Error()
}

func (c *codedError) ErrorCode() Code {
	return c.code
}

func (c *codedError) StackTrace() []uintptr {
	return c.stack
}

func (c *codedError) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fmt.Fprintf(f, "[%s] %s", c.code, c.err.Error())
		fmt.Fprint(f, FormatStack(c))
		return
	}
	fmt.Fprint(f, c.Error())
}
