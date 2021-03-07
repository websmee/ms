package errors

import (
	pkgErrors "github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() pkgErrors.StackTrace
}

func GetStackTrace(err error) pkgErrors.StackTrace {
	e, ok := err.(stackTracer)
	if !ok {
		return nil
	}

	return e.StackTrace()
}
