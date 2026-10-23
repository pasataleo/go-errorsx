package errorsx

import (
	"fmt"
	"runtime"
	"strings"
)

type StackTracer interface {
	StackTrace() []uintptr
}

// Frame is a single stack frame resolved to a source location. Program
// counters are meaningless outside the process that captured them, so an error
// that has crossed a process boundary carries resolved frames instead.
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// StackFramer is implemented by errors whose stack frames have already been
// resolved, as decoded errors are. FormatStack accepts either this or
// StackTracer.
type StackFramer interface {
	StackFrames() []Frame
}

// FormatStack renders an error's stack trace to a string: the stack of the
// first error in its tree that carries one, outermost first, so an error that
// was annotated or wrapped after it was created still shows where it was
// created. It returns an empty string if nothing in the tree carries a stack,
// whether as program counters or as already-resolved frames.
func FormatStack(err error) string {
	frames := firstStack(err)
	if len(frames) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, frame := range frames {
		fmt.Fprintf(&sb, "\n    %s\n        %s:%d", frame.Function, frame.File, frame.Line)
	}
	return sb.String()
}

// firstStack returns the frames of the first error in err's tree that carries
// a stack, walking from err inwards and through aggregated errors in order.
func firstStack(err error) []Frame {
	if err == nil {
		return nil
	}
	if frames := stackFrames(err); len(frames) > 0 {
		return frames
	}

	switch e := err.(type) { //nolint:errorlint // manually walking the error tree
	case interface{ Unwrap() error }:
		return firstStack(e.Unwrap())
	case interface{ Unwrap() []error }:
		for _, child := range e.Unwrap() {
			if frames := firstStack(child); len(frames) > 0 {
				return frames
			}
		}
	}
	return nil
}

// stackFrames returns resolved frames for err from either stack interface.
// Program counters are resolved on demand; an error that already holds
// resolved frames returns them directly.
func stackFrames(err error) []Frame {
	if tracer, ok := err.(StackTracer); ok {
		if frames := resolveFrames(tracer.StackTrace()); len(frames) > 0 {
			return frames
		}
	}
	if framer, ok := err.(StackFramer); ok {
		return framer.StackFrames()
	}
	return nil
}

func resolveFrames(pcs []uintptr) []Frame {
	if len(pcs) == 0 {
		return nil
	}
	var resolved []Frame
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		resolved = append(resolved, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			return resolved
		}
	}
}

func captureStack() []uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	return pcs[:n]
}
