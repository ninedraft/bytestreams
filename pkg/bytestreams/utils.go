package bytestreams

import "io"

type Middleware func(io.Writer) io.Writer

func noopMiddleware(wr io.Writer) io.Writer { return wr }

// Flusher describes type, which interanl buffer can be flushed.
type Flusher interface {
	Flush() error
}

func noopCallback() error {
	return nil
}
