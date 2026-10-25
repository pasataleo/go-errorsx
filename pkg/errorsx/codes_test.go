package errorsx

import "testing"

func TestCodeMappings(t *testing.T) {
	tcs := []struct {
		code Code
		http int
		exit int
	}{
		{OK, 200, 0},
		{InvalidArgument, 400, 64},
		{Unauthenticated, 401, 77},
		{PermissionDenied, 403, 77},
		{NotFound, 404, 66},
		{AlreadyExists, 409, 73},
		{FailedPrecondition, 412, 65},
		{ResourceExhausted, 429, 75},
		{Internal, 500, 70},
		{Unknown, 500, 1},
		{Unimplemented, 501, 69},
		{Unavailable, 503, 69},
		{Code("custom"), 500, 1},
	}
	for _, tc := range tcs {
		t.Run(string(tc.code), func(t *testing.T) {
			if got := tc.code.HTTPStatus(); got != tc.http {
				t.Errorf("HTTPStatus() = %d, want %d", got, tc.http)
			}
			if got := tc.code.ExitCode(); got != tc.exit {
				t.Errorf("ExitCode() = %d, want %d", got, tc.exit)
			}
		})
	}
}

func TestCodeForExitCode(t *testing.T) {
	// Every code with an exit status of its own comes back from it. Shared
	// statuses come back as one code: 69 as Unavailable, 77 as
	// Unauthenticated.
	for _, code := range []Code{OK, InvalidArgument, FailedPrecondition, NotFound, Unavailable, Internal, AlreadyExists, ResourceExhausted, Unauthenticated} {
		if got := CodeForExitCode(code.ExitCode()); got != code {
			t.Errorf("CodeForExitCode(%d) = %q, want %q", code.ExitCode(), got, code)
		}
	}
	for status, want := range map[int]Code{69: Unavailable, 77: Unauthenticated, 1: Unknown, 2: Unknown} {
		if got := CodeForExitCode(status); got != want {
			t.Errorf("CodeForExitCode(%d) = %q, want %q", status, got, want)
		}
	}
}

func TestCodeForHTTPStatus(t *testing.T) {
	// Every built-in code but Internal comes back from its own status; 500 is
	// shared with Unknown, and means Unknown.
	for _, code := range []Code{OK, InvalidArgument, Unauthenticated, PermissionDenied, NotFound, AlreadyExists, FailedPrecondition, ResourceExhausted, Unknown, Unimplemented, Unavailable} {
		if got := CodeForHTTPStatus(code.HTTPStatus()); got != code {
			t.Errorf("CodeForHTTPStatus(%d) = %q, want %q", code.HTTPStatus(), got, code)
		}
	}
	if got := CodeForHTTPStatus(500); got != Unknown {
		t.Errorf("CodeForHTTPStatus(500) = %q, want %q", got, Unknown)
	}
	if got := CodeForHTTPStatus(418); got != Unknown {
		t.Errorf("CodeForHTTPStatus(418) = %q, want %q", got, Unknown)
	}
}
