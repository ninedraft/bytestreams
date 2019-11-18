package bytestreams

import "io"

// Middleware describes a bytestream processor with push semantics.
type Middleware func(io.Writer) io.Writer

// Flusher describes type, which interanl buffer can be flushed.
type Flusher interface {
	Flush() error
}

func noopCallback() error { return nil }
