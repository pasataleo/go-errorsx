package errorsx

// httpStatuses maps each built-in code to its HTTP status. The statuses are
// written as numbers so this package does not import net/http.
var httpStatuses = map[Code]int{
	OK:                 200,
	InvalidArgument:    400,
	Unauthenticated:    401,
	PermissionDenied:   403,
	NotFound:           404,
	AlreadyExists:      409,
	FailedPrecondition: 412,
	ResourceExhausted:  429,
	Internal:           500,
	Unknown:            500,
	Unimplemented:      501,
	Unavailable:        503,
}

// codesForHTTP inverts httpStatuses. Internal and Unknown share 500, which
// maps back to Unknown: all a bare 500 says is that something failed.
var codesForHTTP = map[int]Code{
	200: OK,
	400: InvalidArgument,
	401: Unauthenticated,
	403: PermissionDenied,
	404: NotFound,
	409: AlreadyExists,
	412: FailedPrecondition,
	429: ResourceExhausted,
	500: Unknown,
	501: Unimplemented,
	503: Unavailable,
}

// exitCodes maps each built-in code to a process exit status, following
// sysexits.h so that scripts can tell failures apart without reading stderr.
var exitCodes = map[Code]int{
	OK:                 0,
	InvalidArgument:    64, // EX_USAGE
	FailedPrecondition: 65, // EX_DATAERR
	NotFound:           66, // EX_NOINPUT
	Unimplemented:      69, // EX_UNAVAILABLE
	Unavailable:        69, // EX_UNAVAILABLE
	Internal:           70, // EX_SOFTWARE
	AlreadyExists:      73, // EX_CANTCREAT
	ResourceExhausted:  75, // EX_TEMPFAIL
	Unauthenticated:    77, // EX_NOPERM
	PermissionDenied:   77, // EX_NOPERM
}

// codesForExit inverts exitCodes. Statuses two codes share map back to one:
// 69 to Unavailable, and 77 to Unauthenticated. An exit status says less than
// the error did, so this is the best guess the status allows.
var codesForExit = map[int]Code{
	0:  OK,
	64: InvalidArgument,
	65: FailedPrecondition,
	66: NotFound,
	69: Unavailable,
	70: Internal,
	73: AlreadyExists,
	75: ResourceExhausted,
	77: Unauthenticated,
}

// HTTPStatus is the HTTP status for the code, or 500 for a code without one.
func (c Code) HTTPStatus() int {
	if status, ok := httpStatuses[c]; ok {
		return status
	}
	return 500
}

// CodeForHTTPStatus is the code for an HTTP status, or Unknown for a status
// without one.
func CodeForHTTPStatus(status int) Code {
	if code, ok := codesForHTTP[status]; ok {
		return code
	}
	return Unknown
}

// CodeForExitCode is the code for a process exit status, or Unknown for a
// status without one.
func CodeForExitCode(status int) Code {
	if code, ok := codesForExit[status]; ok {
		return code
	}
	return Unknown
}

// ExitCode is the process exit status for the code, or 1 for a code without
// one.
func (c Code) ExitCode() int {
	if code, ok := exitCodes[c]; ok {
		return code
	}
	return 1
}
