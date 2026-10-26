# v0.2.0

## FEATURES

- `Encode` and `Encoded.Decode` serialise an error tree to JSON and back, keeping codes, wrapping, annotations, aggregation and stacks; `WithoutStacks` leaves the stacks out, and `Frame` and `StackFramer` carry a decoded error's resolved frames.
- `diagnostics` package, for carrying warnings and info alongside errors: `Diagnostic`, `Diagnostics` and a `Builder`.
- `Code.HTTPStatus` and `CodeForHTTPStatus` map codes to HTTP statuses and back, and `Code.ExitCode` and `CodeForExitCode` to sysexits-style exit statuses.

## IMPROVEMENTS

- `FormatStack` renders stacks from `StackFramer` as well as `StackTracer`.

## BUG FIXES

- `ErrorCode` reports the code `New` was given when the error also has a cause.
- `ErrorCode` reports the outermost code where one code is wrapped around another.
- `FormatStack` finds a stack anywhere in the error tree, not only on the error it is given.

<!--
## SECURITY
Vulnerabilities or security-related changes addressed in this release.
-->

<!--
## DEPRECATIONS
Functionality that will be removed in a future release.
-->

<!--
## BREAKING CHANGES
Changes that are not backwards compatible and require updates from consumers.
-->

<!--
## UPGRADE NOTES
Steps required when upgrading from a previous version.
-->

# v0.1.0

## FEATURES

- Error codes via `New` / `Newf` with built-in codes (`NotFound`, `Internal`, etc.) and support for custom `Code` values
- `ErrorCode` to extract the first code in the error chain; `IsCode` to search the entire chain including aggregated branches
- `Wrap` / `Wrapf` to add context while preserving the original error code
- Annotations: attach arbitrary key-value metadata with `Annotate`, retrieve with `GetAnnotation` / `GetAnnotations`
- Aggregation: collect multiple errors into one with `Append`; iterate with `Errors`; `errors.Is` / `errors.As` search all branches
- Stack traces captured automatically on `New` / `Newf`; `Wrap` captures a stack only when the wrapped error does not already carry one
- `StackTracer` interface for extending or inspecting stack traces
- `FormatStack` for rendering a stack trace to a string
- `%+v` formatting on all error types — includes codes, annotations, and the full error chain with stack traces
