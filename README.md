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

## JSON

Errors serialise to JSON and back, preserving the whole tree — codes, wrapping, annotations, aggregation, and stack
traces — rather than collapsing it to a message and a code. This is what lets an error cross a process boundary (see
[go-plugin](https://github.com/pasataleo/go-plugin)) and still answer `IsCode`, `GetAnnotation`, and `%+v` on the far
side.

```go
type Encoded struct {
    Type        string                 `json:"type"`
    Code        Code                   `json:"code,omitempty"`
    Message     string                 `json:"message,omitempty"`
    Annotations map[string]interface{} `json:"annotations,omitempty"`
    Stack       []Frame                `json:"stack,omitempty"`
    Current     *Encoded               `json:"current,omitempty"`
    Cause       *Encoded               `json:"cause,omitempty"`
    Errors      []*Encoded             `json:"errors,omitempty"`
}

func Encode(err error, opts ...EncodeOpt) *Encoded
func (e *Encoded) Decode() error
```

`Encoded` is a plain struct with JSON tags, so it composes into a larger document with `encoding/json` as normal, and
neither direction needs a custom marshaller:

```go
data, err := json.Marshal(errorsx.Encode(err))

var encoded errorsx.Encoded
if err := json.Unmarshal(data, &encoded); err != nil {
    return err
}
original := encoded.Decode()
```

`Encode(nil)` returns `nil`, and a `nil` `*Encoded` decodes to a `nil` error.

**This package does not import `encoding/json`.** Struct tags are inert, and `Encode` and `Decode` are pure tree
transforms, so the dependency belongs to whoever marshals an `Encoded` rather than to everyone who handles an error.
That matters for consumers targeting TinyGo or WASM — [go-web](https://github.com/pasataweb/go-web) depends on this
library and deliberately uses neither `encoding/json` nor `reflect` — and it means a caller is free to marshal an
`Encoded` with a codec of its own.

### Node types

One node per error type, tagged by `type`. Every node kind uses the same struct; unused fields are omitted.

| `type` | Produced by | Fields |
|--------|-------------|--------|
| `coded` | `New` / `Newf` | `code`, `message`, `stack` |
| `wrapped` | `Wrap` / `Wrapf` | `current` (the wrapping error), `cause` (the wrapped error) |
| `annotated` | `Annotate` | `annotations`, `cause` |
| `aggregated` | `Append` | `errors` |
| `error` | any other error | `message` |

The canonical example from [Formatting](#formatting) encodes as:

```json
{
  "type": "annotated",
  "annotations": {"user_id": 42},
  "cause": {
    "type": "wrapped",
    "current": {
      "type": "coded",
      "code": "not_found",
      "message": "loading profile",
      "stack": [{"function": "main.loadProfile", "file": "/src/main.go", "line": 42}]
    },
    "cause": {
      "type": "coded",
      "code": "not_found",
      "message": "user not found"
    }
  }
}
```

An unrecognised `type` decodes as an opaque `error` node, so a newer producer does not break an older consumer.

### Stack traces

`StackTrace` returns `[]uintptr`, and program counters from one process mean nothing in another. `Encode` therefore
resolves frames before writing them:

```go
type Frame struct {
    Function string `json:"function"`
    File     string `json:"file"`
    Line     int    `json:"line"`
}

type StackFramer interface {
    StackFrames() []Frame
}
```

A decoded error reports its frames through `StackFramer`; its `StackTrace` returns nothing, since the program counters
did not survive. `FormatStack` reads either interface — resolving counters on demand, or taking already-resolved frames
as they are — so `%+v` on a decoded error still renders the stack captured in the original process.

Pass `errorsx.WithoutStacks()` to `Encode` to drop frames entirely, which matters when errors travel over a wire on
every call.

### Caveats

- **Annotation values round-trip as JSON**, so an `int` comes back as a `float64` and a struct as a `map`. `Encode`
  copies values as-is rather than inspecting them, so a value the caller's codec cannot marshal fails that marshal —
  guard annotations at the point they are attached if they have to survive a wire.
- **Sentinel identity is lost.** A foreign error encodes as an opaque `error` node carrying its message, so after a
  round trip `errors.Is(err, os.ErrNotExist)` is false. Codes survive; pointer identity cannot.
- **Annotations are attached to the node that carried them**, not flattened, so `GetAnnotations` returns the same
  merged map before and after a round trip — including the precedence rule that closer annotations win.
