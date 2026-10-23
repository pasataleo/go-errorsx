# go-errorsx

Package `errorsx` provides structured error types that compose together: error codes, wrapping, annotations, and aggregation.

All types integrate with the standard `errors.Is` and `errors.As` functions.

Package `diagnostics` carries warnings alongside errors, for operations that find several problems and only some of
them fatal. See [Diagnostics](#diagnostics).

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

## Diagnostics

Some operations find several problems, and only some of them are fatal. Validating a configuration is one example: an
unknown reference might be an error, while an unused setting is only worth a warning. An `error` can't express that.
Anything returned as an error makes `if err != nil` fail, so warnings are carried alongside errors rather than inside
them.

Package `github.com/pasataleo/go-errorsx/pkg/diagnostics` has three public types:

| Type | Kind | Purpose |
|---|---|---|
| `Diagnostic` | Interface | A single problem. Callers can implement it themselves; the package's own come from a `Builder`. |
| `Diagnostics` | `[]Diagnostic` | The diagnostics collected from one operation. |
| `Builder` | Struct | Constructs the package's own diagnostics. |

### `Diagnostic`

```go
type Severity string

const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)

type Diagnostic interface {
    Severity() Severity          // required
    Summary() string             // required: a short, one-line description
    Detail() string              // optional: a longer explanation; "" if none
    Annotations() map[string]any // optional: nil if none; callers must not modify it
    Cause() error                // optional: the error behind the problem; nil if none
}
```

Annotations carry extra structured data, such as source location (file or component, line, column, snippet). The
package gives no meaning to any key.

`diagnostics.String(d)` formats any diagnostic on one line, as `<severity>: <summary>`, followed by `: <detail>` when
there is a detail. The package's own diagnostics also implement `fmt.Stringer` using it.

### Building diagnostics

Each constructor takes the required severity and summary, and returns a `*Builder`:

```go
diagnostics.Error("summary")
diagnostics.Warning("summary")
diagnostics.Info("summary")

diagnostics.Errorf("component %q not registered", name)
diagnostics.Warningf(format, args...)
diagnostics.Infof(format, args...)

diagnostics.FromError(err) // Error severity, the error's message as summary, err as the cause
```

Adding the optional parts:
- **`.Detail(string)` / `.Detailf(format, args...)`** set the detail.
- **`.Annotate(key string, value any)`** adds one annotation.
- **`.WithAnnotations(map[string]any)`** adds several. The map is copied, so later changes to it don't affect the
  diagnostic.
- **`.Cause(error)`** attaches the error behind the problem.
  - The original error is kept, so `errors.Is`, `errors.As`, `ErrorCode` and `IsCode` still work on it.
  - Annotations from anywhere in the error's tree are copied onto the diagnostic, as `GetAnnotations` returns them.
    They never replace annotations set directly on the builder, whichever was called first.

Finishing:
- **`.Build()`** returns the `Diagnostic`.
- **`.Append(diags Diagnostics)`** builds the diagnostic, appends it to `diags`, and returns the result, so a list can
  be built in a single expression. `diags` can be nil.

```go
var diags diagnostics.Diagnostics
diags = diagnostics.Errorf("component %q not registered", "Card").
    Detailf("<Card> is used in %q, but no component %q is registered.", "Page", "Card").
    Annotate("component", "Page").
    Annotate("line", 12).
    Append(diags)
```

A builder makes exactly one diagnostic, and misusing one panics, since each case is a programming mistake:
- building with an empty summary, or calling `FromError(nil)`
- calling `.Build()` or `.Append()` a second time on the same builder
- calling any other builder method after it has been built

### `Diagnostics`

```go
diags.HasErrors()                                          // whether any diagnostic has SeverityError
diags.Filter(diagnostics.SeverityError, ...)               // a new set with only those severities
diags.WithAnnotations(map[string]any{"component": "Page"}) // a new set with the annotations added to each
```

`WithAnnotations` is for adding context later. For example, a template parser can report problems without knowing
which component it is parsing, and its caller then adds the component name to all of them. Keys a diagnostic already
has are not overwritten. Each diagnostic is wrapped rather than copied, so the originals are kept unmodified, and ones
the caller implemented work the same as the package's own.

Diagnostics stay in the order they were added. There is no built-in sorting, so callers decide the order. Because
`Diagnostics` is a slice, merging two sets is the built-in `append`: `diags = append(diags, other...)`.

Like the rest of this module, the package depends on nothing outside the standard library and doesn't import
`encoding/json`, so TinyGo and WASM callers can use it.
