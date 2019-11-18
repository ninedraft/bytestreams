package bytestreams

import "io"

// StackWriter describes a layered set of writers, which can be used to pipeline multiple bytestream handlers.
// It handles not only data pipelining, but .Close and .Flush calls too.
type StackWriter struct {
	wr io.Writer
	config
}

// NewStackWriter creates a new stacked writer with provided target writer and optional configuration.
func NewStackWriter(wr io.Writer, options ...Option) *StackWriter {
	var stack = &StackWriter{
		wr: wr,
		config: config{
			onClose:      noopCallback,
			onFlush:      noopCallback,
			reduceErrors: func(_, old error) error { return old },
		},
	}
	for _, setOption := range options {
		setOption(&stack.config)
	}
	return stack
}

// Push puts middlewares on top of the writer stack.
// If wrapped writers impleement io.Closer or FLusher,
// then corresponding methods will be pushed too callback stack too.
func (stack *StackWriter) Push(wrappers ...Middleware) *StackWriter {
	for _, wrap := range wrappers {
		var wr = wrap(stack.wr)
		if closer, isCloser := wr.(io.Closer); isCloser {
			WithCloser(closer)(&stack.config)
		}
		if flusher, isFlusher := wr.(Flusher); isFlusher {
			WithFlusher(flusher)(&stack.config)
		}
		stack.wr = wr
	}
	return stack
}

// Flush calls stacked .Flush methods, in reverse order of pushing of middlewares.
// Error handling policy depends on special optional callback, .Flush method will return
// a first occured error by default.
func (stack *StackWriter) Flush() error {
	return stack.onFlush()
}

// Close calls stacked .Close methods in reverse order of pushing of middlewares.
// Error handling policy depends on special optional callback, .Close method will return
// a first occured error by default.
func (stack *StackWriter) Close() error {
	return stack.onClose()
}
