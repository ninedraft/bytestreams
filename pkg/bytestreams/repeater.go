package bytestreams

// Repeater emits a one chunk of data indefinitely.
type Repeater struct {
	data   []byte
	offset int
}

// NewRepeater creates a new repeater with provided chunk of data.
func NewRepeater(data []byte) *Repeater {
	var repeater = new(Repeater)
	repeater.Reset(data)
	return repeater
}

// Reset drops internal state of repeater and sets a new chunk to repeat.
// The data parameter can be nil. This method is suitable for usage in sync.Pool.
func (repeater *Repeater) Reset(data []byte) {
	repeater.offset = 0
	repeater.data = data
}

// Read next portion of data and advance internal counter.
func (repeater *Repeater) Read(dst []byte) (int, error) {
	var n = len(dst)
	var data = repeater.data
	var dataLen = len(data)
	var written = 0
	for i := 0; i <= n/dataLen; i++ {
		var offset = (written + repeater.offset) % dataLen
		written += repeater.slab(offset, dst[written:])
	}
	repeater.offset = (n + repeater.offset) % dataLen
	return n, nil
}

func (repeater *Repeater) slab(offset int, dst []byte) int {
	var data = repeater.data
	var head = data[offset:]
	var written = copy(dst, head)
	// the if condition can be ommited without logic violation,
	// but we can avoid extra copy call
	// TODO: benchmark bounds checks VS naive method (no branching)
	if written < len(dst) {
		var tail = data[:len(data)-written]
		var dstTail = dst[written:]
		written += copy(dstTail, tail)
	}
	return written
}
