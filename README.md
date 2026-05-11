# go-errorsx

Package `errorsx` provides structured error types that compose together: error codes, wrapping, annotations, and aggregation.

All types integrate with the standard `errors.Is` and `errors.As` functions.

## Installation

```sh
go get github.com/pasataleo/go-errorsx
```

## Error Codes

Attach a semantic code to an error:

```go
err := errorsx.New(errorsx.NotFound, nil, "user not found")
errorsx.ErrorCode(err) // errorsx.NotFound
```

Built-in codes, ordered by HTTP status:

| Code | HTTP |
|------|------|
| `OK` | 200 |
| `InvalidArgument` | 400 |
| `Unauthenticated` | 401 |
| `PermissionDenied` | 403 |
| `NotFound` | 404 |
| `AlreadyExists` | 409 |
| `FailedPrecondition` | 412 |
| `ResourceExhausted` | 429 |
| `Internal` | 500 |
| `Unknown` | 500 |
| `Unimplemented` | 501 |
| `Unavailable` | 503 |

`Code` is a `string` type, so you can define your own:

```go
const RateLimited errorsx.Code = "rate_limited"
```

`ErrorCode` returns the first code found by walking the error chain. Use `IsCode` to search the entire chain, including all branches of aggregated errors:

```go
errorsx.ErrorCode(err) == errorsx.NotFound // first code in chain
errorsx.IsCode(err, errorsx.NotFound)      // any error in chain
```

## Wrapping

Add context to an error while preserving its code:

```go
err := errorsx.Wrap(err, "loading config")
err := errorsx.Wrapf(err, "loading config for %s", name)
```

Error messages format as `"outer (inner)"`. Wrapping `nil` returns `nil`.

## Annotations

Attach key-value metadata to an error:

```go
err = errorsx.Annotate(err, "user_id", 42)

val, ok := errorsx.GetAnnotation(err, "user_id")
all := errorsx.GetAnnotations(err)
```

Annotating the same error multiple times reuses the existing annotation map. Annotating `nil` returns `nil`.

## Aggregation

Collect multiple errors into one:

```go
err = errorsx.Append(err, validateName(), validateEmail())

for _, e := range errorsx.Errors(err) {
    // handle each error
}
```

`Append` merges into an existing aggregated error rather than nesting. `errors.Is` and `errors.As` search through all aggregated errors.

## Stack Traces

`New` and `Newf` capture a stack trace at the call site. `Wrap` and `Wrapf` capture a stack trace only when the error being wrapped does not already carry one, so wrapping a wrapped error does not overwrite the original call site.

The `StackTracer` interface is public — implement it in your own error types to integrate with `FormatStack`:

```go
type StackTracer interface {
    StackTrace() []uintptr
}
```

Use `FormatStack` to render a stack trace to a string. It returns an empty string if the error does not implement `StackTracer` or carries no frames:

```go
fmt.Println(errorsx.FormatStack(err))
```

Stack traces appear automatically in `%+v` output.

## Formatting

All error types implement `fmt.Formatter`. Use `%v` for the standard message or `%+v` for a verbose representation that includes codes, annotations, the full error chain, and stack traces:

```go
err := errorsx.Annotate(
    errorsx.Wrap(
        errorsx.New(errorsx.NotFound, nil, "user not found"),
        "loading profile",
    ),
    "user_id", 42,
)

fmt.Sprintf("%v", err)  // "loading profile (user not found)"
fmt.Sprintf("%+v", err) // "[not_found] loading profile\n    goroutine ...\n  - [not_found] user not found user_id=42"
```
