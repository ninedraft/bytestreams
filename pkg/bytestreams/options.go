package bytestreams

import (
	"io"
)

type config struct {
	onClose      func() error
	onFlush      func() error
	reduceErrors func(new, old error) error
}

// Option is an optional bytestream config setter.
type Option func(cfg *config)

// WithCloser put io.Closer.Close method as callback on stack of callbacks, called by .Close method.
func WithCloser(closer io.Closer) Option {
	return WithClose(closer.Close)
}

// WithClose put close callback on stack of callbacks, called by .Close method.
func WithClose(cl func() error) Option {
	return func(pw *config) {
		var oldCloser = pw.onClose
		pw.onClose = func() error {
			var err = cl()
			var errOld = oldCloser()
			return pw.reduceErrors(err, errOld)
		}
	}
}

// WithFlusher put Flusher.Flush method as callback on stack of callbacks, called by .Flush method.
func WithFlusher(flusher Flusher) Option {
	return WithFlush(flusher.Flush)
}

// WithFlush put flush callback on stack of callbacks, called by .Flush method.
func WithFlush(flush func() error) Option {
	return func(pw *config) {
		var oldFlush = pw.onFlush
		pw.onFlush = func() error {
			var err = flush()
			var errOld = oldFlush()
			return pw.reduceErrors(err, errOld)
		}
	}
}

// KeepLastError forces stream too keep only last catched errors
// while .Close and .Flush methods execution.
func KeepLastError(pw *config) {
	pw.reduceErrors = func(new, _ error) error {
		return new
	}
}

// KeepFirstError forces stream too keep only first catched error
// while .Close and .Flush methods execution.
func KeepFirstError(pw *config) {
	pw.reduceErrors = func(_, old error) error {
		return old
	}
}

// WithErrorReducer forces stream too use user-defined function to
// solve multiple error conflicts while .Close and .Flush methods execution.
func WithErrorReducer(reduce func(new, old error) error) Option {
	return func(pw *config) {
		pw.reduceErrors = reduce
	}
}

// AccumulateErrors stores all errors, occured while .Close and .Flush methods execution
// in ErrorChain.
func AccumulateErrors(pw *config) {
	pw.reduceErrors = func(new, old error) error {
		var prevChain, isPrevChain = new.(ErrorChain)
		var nextChain, isNextChain = old.(ErrorChain)
		switch {
		case isPrevChain && !isNextChain:
			return append(prevChain, nextChain)
		case isNextChain && !isPrevChain:
			return append(ErrorChain{new}, nextChain...)
		case isPrevChain && isNextChain:
			return append(prevChain, nextChain...)
		default:
			return ErrorChain{new, old}
		}
	}
}
