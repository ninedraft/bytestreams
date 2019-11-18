package bytestreams

import "io"

type StackWriter struct {
	wr io.Writer
	config
}

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

func (stack *StackWriter) Flush() error {
	return stack.onFlush()
}

func (stack *StackWriter) Close() error {
	return stack.onClose()
}
