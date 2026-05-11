package errorsx

import (
	"fmt"
	"runtime"
	"strings"
)

type StackTracer interface {
	StackTrace() []uintptr
}

func FormatStack(err error) string {
	st, ok := err.(StackTracer)
	if !ok || len(st.StackTrace()) == 0 {
		return ""
	}
	pcs := st.StackTrace()
	frames := runtime.CallersFrames(pcs)
	var sb strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&sb, "\n    %s\n        %s:%d", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return sb.String()
}

func captureStack() []uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	return pcs[:n]
}
