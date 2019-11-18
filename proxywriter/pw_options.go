package proxywriter

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// WithCloser put io.Closer.Close method as callback on stack of callbacks, called by .Close method.
func WithCloser(closer io.Closer) Option {
	return func(pw *ProxyWriter) {
		var next = pw.onClose
		pw.onClose = func() error {
			return pw.reduceError(closer.Close(), next())
		}
	}
}

// WithClose put close callback on stack of callbacks, called by .Close method.
func WithClose(cl func() error) Option {
	return func(pw *ProxyWriter) {
		var next = pw.onClose
		pw.onClose = func() error {
			return pw.reduceError(cl(), next())
		}
	}
}

// WithFlusher put Flusher.Flush method as callback on stack of callbacks, called by .Flush method.
func WithFlusher(flusher Flusher) Option {
	return func(pw *ProxyWriter) {
		var next = pw.onFlush
		pw.onFlush = func() error {
			return pw.reduceError(flusher.Flush(), next())
		}
	}
}

// WithFlush put flush callback on stack of callbacks, called by .Flush method.
func WithFlush(flush func() error) Option {
	return func(pw *ProxyWriter) {
		var next = pw.onFlush
		pw.onFlush = func() error {
			return pw.reduceError(flush(), next())
		}
	}
}

// ErrorChain stores multiple indepndent errors.
// It supports standart error matching mechanisms a.k.a errors.Is and errors.As.
type ErrorChain []error

func (errs ErrorChain) Error() string {
	var builder = &strings.Builder{}
	for i, err := range errs {
		_, _ = fmt.Fprintf(builder, "%s", err)
		if i < len(errs)-1 {
			_, _ = builder.WriteString("; ")
		}
	}
	return builder.String()
}

// As method searches for first matching error in chain by errors.As function.
func (errs ErrorChain) As(asErr interface{}) bool {
	for _, err := range errs {
		if errors.As(err, asErr) {
			return true
		}
	}
	return false
}

// Is method searches for first matching error in chain by errors.Is function.
func (errs ErrorChain) Is(isErr error) bool {
	for _, err := range errs {
		if errors.Is(err, isErr) {
			return true
		}
	}
	return false
}

// KeepLastError forces ProxyWriter too keep only last catched errors
// while .Close and .Flush methods execution.
func KeepLastError(pw *ProxyWriter) {
	pw.reduceError = func(next, _ error) error {
		return next
	}
}

// KeepFirstError forces ProxyWriter too keep only first catched error
// while .Close and .Flush methods execution.
func KeepFirstError(pw *ProxyWriter) {
	pw.reduceError = func(_, prev error) error {
		return prev
	}
}

// WithErrorReducer forces ProxyWriter too use user-defined function to
// solve multiple error conflicts while .Close and .Flush methods execution.
func WithErrorReducer(reduce func(a, b error) error) Option {
	return func(pw *ProxyWriter) {
		pw.reduceError = reduce
	}
}

// AccumulateErrors stores all errors, occured while .Close and .Flush methods execution
// in ErrorChain.
func AccumulateErrors(pw *ProxyWriter) {
	pw.reduceError = func(next, prev error) error {
		var prevChain, isPrevChain = prev.(ErrorChain)
		var nextChain, isNextChain = next.(ErrorChain)
		switch {
		case isPrevChain && !isNextChain:
			return append(prevChain, nextChain)
		case isNextChain && !isPrevChain:
			return append(ErrorChain{prev}, nextChain...)
		case isPrevChain && isNextChain:
			return append(prevChain, nextChain...)
		default:
			return ErrorChain{prev, next}
		}
	}
}
