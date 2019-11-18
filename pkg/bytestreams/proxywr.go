package bytestreams

import (
	"fmt"
	"io"
)

// ProxyWriter describes a wrapper around io.Writer, which can be used to write multiple chunks of data.
// It can be good idea to use ProxyWriter with bufio.Writer as backend to increase throughput.
type ProxyWriter struct {
	err     error
	written int64
	wr      io.Writer
	config
}

// NewProxyWriter creates a new ProxyWriter with backend writer and options.
func NewProxyWriter(wr io.Writer, options ...Option) *ProxyWriter {
	// if we just a thin layer over another proxy writer, when just
	// return the backend proxy writer. So multiple nested New calls
	// will not fill app stack with useless proxy writers.
	if pw, isProxy := wr.(*ProxyWriter); isProxy && len(options) == 0 {
		return pw
	}
	var pw = &ProxyWriter{
		wr: wr,
		config: config{
			onClose:      noopCallback,
			onFlush:      noopCallback,
			reduceErrors: func(_, old error) error { return old },
		},
	}
	if closer, isCloser := wr.(io.Closer); isCloser {
		pw.onClose = closer.Close
	}
	if flusher, isFlusher := wr.(Flusher); isFlusher {
		pw.onFlush = flusher.Flush
	}
	for _, setOption := range options {
		setOption(&pw.config)
	}
	return pw
}

// Write writes data, if ProxyWriter had encountered no errors.
// It returns stored error, if exists.
func (pw *ProxyWriter) Write(data []byte) (int, error) {
	if pw.HasErr() {
		return 0, pw.Err()
	}
	var n, err = pw.wr.Write(data)
	if err != nil {
		pw.err = err
	}
	pw.written += int64(n)
	return n, err
}

// WriteString calls io.WriteString over backend writer, which can boost writes in specific cases, because it skips
// a .Write call.
func (pw *ProxyWriter) WriteString(data string) (int, error) {
	if pw.HasErr() {
		return 0, pw.Err()
	}
	var n, err = io.WriteString(pw.wr, data)
	if err != nil {
		pw.err = err
	}
	pw.written += int64(n)
	return n, err
}

// Printf prints formatted string to the backend writer, if ProxyWriter had encountered no errors on previous calls.
func (pw *ProxyWriter) Printf(ff string, args ...interface{}) {
	_, _ = fmt.Fprintf(pw, ff, args...)
}

// Println prints string with a trailing newline(\n) to the backend writer,
// if ProxyWriter had encountered no errors on previous calls.
func (pw *ProxyWriter) Println(vv ...interface{}) {
	_, _ = fmt.Fprintln(pw, vv...)
}

// PrintSlice prints slice of strings, separated by sep parameter value to the backend writer,
// if ProxyWriter had encountered no errors on previous calls.
// PrintSlice doesn't append separator to string afer last item.
func (pw *ProxyWriter) PrintSlice(sep string, items ...string) {
	var bsep = []byte(sep)
	for i, item := range items {
		if pw.HasErr() {
			return
		}
		_, _ = pw.Write([]byte(item))
		if i < len(items)-1 {
			_, _ = pw.Write(bsep)
		}
	}
}

// ReadFrom calls io.Copy to pipe data from reader to background writer if no errors had been
// catched on previous method calls. ReadFrom doesn't shadow any errors.
func (pw *ProxyWriter) ReadFrom(re io.Reader) (int64, error) {
	if pw.HasErr() {
		return 0, pw.Err()
	}
	var n, err = io.Copy(pw.wr, re)
	if err != nil {
		pw.err = err
	}
	pw.written += n
	return n, err
}

// Err returns error if exists.
func (pw *ProxyWriter) Err() error {
	return pw.err
}

// HasErr reports if ProxyWriter had encountered any errors on previous calls.
func (pw *ProxyWriter) HasErr() bool {
	return pw.err != nil
}

// Written returns number of written bytes.
func (pw *ProxyWriter) Written() int64 {
	return pw.written
}

// WrittenInt returns number of written bytes as int number.
// Just a shortland for int(ProxyWriter.Written()).
func (pw *ProxyWriter) WrittenInt() int {
	return int(pw.Written())
}

// Result returns a number of written bytes and error value.
// Useful for implementation of io.WriteTo method.
// Just a shortland for return pw.Written(), pw.Err().
func (pw *ProxyWriter) Result() (int64, error) {
	return pw.Written(), pw.Err()
}

// Close calls close callback.
// If the background writer is a io.Closer, then it's .Close method will be called.
// If multiple close callbacks returns non-nil errors, then ProxyWriter will use
// reduceErrors callback. By default only the first error survives.
func (pw *ProxyWriter) Close() error {
	return pw.onClose()
}

// Flush calls flush callback.
// If the background writer is a Flusher, then it's .Flush method will be called.
// If multiple flush callbacks returns non-nil errors, then ProxyWriter will use
// reduceErrors callback. By default only the first error survives.
func (pw *ProxyWriter) Flush() error {
	return pw.onFlush()
}
