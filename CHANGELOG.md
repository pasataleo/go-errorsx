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

<!--
## IMPROVEMENTS
Enhancements to existing functionality.
-->

<!--
## BUG FIXES
Issues that have been resolved.
-->

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
