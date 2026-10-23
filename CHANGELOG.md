# v0.2.0

## FEATURES

- `Encode` and `Encoded.Decode` serialise an error tree and rebuild it, preserving codes, wrapping, annotations, aggregation, and stack traces rather than collapsing them to a message
- `Encoded` is a plain struct with JSON tags, so it composes into a larger document with `encoding/json` and needs no custom marshaller in either direction
- `WithoutStacks` omits stack frames from an encoded tree, for errors that travel over a wire on every call
- `Frame` and `StackFramer` for errors whose stack frames have already been resolved, as decoded errors are
- `diagnostics` package for operations that find several problems, only some of them fatal: a `Diagnostic` carries a severity (`error`, `warning` or `info`), a summary, and optionally a detail, annotations and the error behind it, so warnings travel alongside errors instead of failing `if err != nil`
- `diagnostics.Builder`, from `Error`, `Warning`, `Info`, their `f` variants and `FromError`, with `Detail`, `Annotate`, `WithAnnotations` and `Cause`, finished by `Build` or by `Append` onto a list in one expression. A cause keeps the original error, so `errors.Is`, `ErrorCode` and `IsCode` still work on it, and its tree's annotations are copied onto the diagnostic without replacing the builder's own
- `diagnostics.Diagnostics` with `HasErrors`, `Filter` by severity, and `WithAnnotations` to add context to a whole set after the fact without overwriting keys or modifying the originals. Every diagnostic is wrapped rather than copied, so the originals are kept, including ones the caller implemented
- `diagnostics.String` formats any diagnostic on one line, and the package's own diagnostics implement `fmt.Stringer` with it

## IMPROVEMENTS

- `FormatStack` renders stacks from either `StackTracer` or `StackFramer`, so a decoded error still reports the stack captured in the process that created it
- This package still imports nothing outside the standard library, and in particular does not import `encoding/json` — struct tags are inert, so consumers targeting TinyGo or WASM pay nothing for the new API unless they marshal an `Encoded`

## BUG FIXES

- `ErrorCode` now reports the code `New` was given when the error also has a cause. A wrapped error has two children and `Unwrap` returns the cause, so the wrapping error — the one holding the code — was off the chain `errors.As` walks: `ErrorCode(New(InvalidArgument, err, "decoding request"))` reported the code of `err`, and a code wrapped around an error carrying none was lost entirely. `Wrap` was unaffected, because it takes its code from the cause, where the chain finds it anyway
- Where a code is wrapped around another code, `ErrorCode` reports the outermost, which is the more recent judgement about the failure. `IsCode` still reports every code in the tree, as documented
- `FormatStack` renders the stack of the first error in the tree that carries one, rather than only the error it is given. An annotated error, or one wrapped with `fmt.Errorf`, implements no stack interface itself, so `FormatStack` returned nothing for it even though the error underneath had a stack and `%+v` showed it

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
