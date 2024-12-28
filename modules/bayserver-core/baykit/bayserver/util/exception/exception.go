package exception

import (
	"fmt"
	"runtime"
)

type Exception interface {
	error
	Frames() *runtime.Frames
}

type ExceptionImpl struct {
	err     error
	Message string
	frames  *runtime.Frames
}

func (b *ExceptionImpl) Error() string {
	if b.err == nil {
		return b.Message

	} else {
		return b.Message + "(" + b.err.Error() + ")"
	}
}

func (b *ExceptionImpl) Frames() *runtime.Frames {
	return b.frames
}

func (b *ExceptionImpl) ConstructException(skip int, err error, format string, args ...interface{}) {
	b.err = err
	b.Message = fmt.Sprintf(format, args...)
	b.frames = MakeStackTrace(skip)
}

func MakeStackTrace(skip int) *runtime.Frames {
	stack := make([]uintptr, 1024)
	for {
		n := runtime.Callers(0, stack[:])
		if n < len(stack) {
			frames := runtime.CallersFrames(stack[:n])
			for i := 0; i < skip; i++ {
				frames.Next()
			}
			return frames
		}
		stack = make([]uintptr, len(stack)*2)
	}
}

func NewException(format string, args ...interface{}) Exception {
	e := &ExceptionImpl{}
	e.ConstructException(4, nil, format, args...)
	return e
}

func NewExceptionFromError(err error) Exception {
	e := &ExceptionImpl{}
	e.ConstructException(4, err, err.Error())
	return e
}
