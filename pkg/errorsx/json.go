package errorsx

import (
	"errors"
)

// Node types used by Encoded. One per error type this package produces, plus
// TypeError for any error it does not recognise.
const (
	TypeCoded      = "coded"
	TypeWrapped    = "wrapped"
	TypeAnnotated  = "annotated"
	TypeAggregated = "aggregated"
	TypeError      = "error"
)

// Encoded is a serialisable representation of an error tree. It is a plain
// struct with JSON tags, so it composes into a larger document with
// encoding/json as normal and neither direction needs a custom marshaller.
//
// Note that this package does not import encoding/json itself — struct tags
// are inert, and Encode and Decode are pure tree transforms. Callers targeting
// TinyGo or WASM, where reflection is expensive or unavailable, pay nothing
// for this file unless they marshal an Encoded.
//
// Every node kind uses this struct; unused fields are omitted. Type identifies
// the kind:
//
//	coded       code, message, stack
//	wrapped     current (the wrapping error), cause (the wrapped error)
//	annotated   annotations, cause
//	aggregated  errors
//	error       message
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

// EncodeOpt is a functional option for Encode.
type EncodeOpt func(*encodeOpts)

type encodeOpts struct {
	stacks bool
}

// WithoutStacks omits stack frames from the encoded tree. Worth passing when
// errors travel over a wire on every call, where a resolved stack per error is
// a lot of bytes for something the consumer usually cannot act on.
func WithoutStacks() EncodeOpt {
	return func(o *encodeOpts) {
		o.stacks = false
	}
}

// Encode converts err into a serialisable representation, preserving the whole
// tree: codes, wrapping, annotations, aggregation, and stack traces. Encoding
// nil returns nil.
//
// Errors this package did not produce are encoded as opaque TypeError nodes
// carrying their message, so their sentinel identity does not survive a round
// trip. Codes do.
//
// Annotation values are copied as-is. It is the caller's codec that must be
// able to marshal them; a value it cannot handle fails that marshal rather
// than this call.
func Encode(err error, options ...EncodeOpt) *Encoded {
	opts := encodeOpts{stacks: true}
	for _, option := range options {
		option(&opts)
	}
	return encode(err, &opts)
}

func encode(err error, opts *encodeOpts) *Encoded {
	if err == nil {
		return nil
	}

	switch e := err.(type) { //nolint:errorlint // encoding the concrete tree, not searching it
	case *codedError:
		encoded := &Encoded{
			Type:    TypeCoded,
			Code:    e.code,
			Message: e.err.Error(),
		}
		if opts.stacks {
			encoded.Stack = stackFrames(e)
		}
		return encoded
	case *wrappedError:
		return &Encoded{
			Type:    TypeWrapped,
			Current: encode(e.current, opts),
			Cause:   encode(e.wrapped, opts),
		}
	case *annotatedError:
		return &Encoded{
			Type:        TypeAnnotated,
			Annotations: copyAnnotations(e.annotations),
			Cause:       encode(e.err, opts),
		}
	case *aggregatedError:
		errs := make([]*Encoded, 0, len(e.errs))
		for _, child := range e.errs {
			if encoded := encode(child, opts); encoded != nil {
				errs = append(errs, encoded)
			}
		}
		return &Encoded{
			Type:   TypeAggregated,
			Errors: errs,
		}
	default:
		return &Encoded{
			Type:    TypeError,
			Message: err.Error(),
		}
	}
}

// Decode rebuilds an error from its encoded representation. Decoding nil
// returns a nil error.
//
// Annotation values round-trip as whatever the caller's codec produced, so
// after a JSON round trip an int arrives as a float64 and a struct as a map.
// An unrecognised Type decodes as an opaque error carrying its message, so a
// newer producer does not break an older consumer.
func (e *Encoded) Decode() error {
	if e == nil {
		return nil
	}

	switch e.Type {
	case TypeCoded:
		code := e.Code
		if code == "" {
			code = Unknown
		}
		return &codedError{
			code:   code,
			err:    errors.New(e.Message),
			frames: e.Stack,
		}
	case TypeWrapped:
		current := e.Current.Decode()
		if current == nil {
			// A wrapped node with no wrapping error is malformed, but its
			// message is still worth keeping.
			current = errors.New(e.Message)
		}
		return &wrappedError{
			current: current,
			wrapped: e.Cause.Decode(),
		}
	case TypeAnnotated:
		err := e.Cause.Decode()
		if err == nil {
			err = errors.New(e.Message)
		}
		return &annotatedError{
			err: err,
			// Annotate mutates this map in place, so it must never be nil.
			annotations: copyAnnotationsNonNil(e.Annotations),
		}
	case TypeAggregated:
		errs := make([]error, 0, len(e.Errors))
		for _, child := range e.Errors {
			if err := child.Decode(); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return &aggregatedError{errs: errs}
	default:
		return errors.New(e.Message)
	}
}

// copyAnnotations snapshots an annotation map so that later calls to Annotate
// on the original error do not mutate an already-encoded tree.
func copyAnnotations(annotations map[string]interface{}) map[string]interface{} {
	if len(annotations) == 0 {
		return nil
	}
	return copyAnnotationsNonNil(annotations)
}

func copyAnnotationsNonNil(annotations map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{}, len(annotations))
	for key, value := range annotations {
		copied[key] = value
	}
	return copied
}
