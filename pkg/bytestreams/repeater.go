package bytestreams

// Repeater emits a one chunk of data indefinitely.
type Repeater struct {
	data []byte
	i    int
}

// NewRepeater creates a new repeater with provided chunk of data.
func NewRepeater(data []byte) *Repeater {
	var repeater = new(Repeater)
	repeater.Reset(data)
	return repeater
}

// Reset drops internal state of repeater and sets a new chunk to repeat.
// The data parameter can be nil. This methid is suitable for usage in sync.Pool.
func (repeater *Repeater) Reset(data []byte) {
	repeater.i = 0
	repeater.data = data
}

// Read next portion of data and advance internal counter.
func (repeater *Repeater) Read(dst []byte) (int, error) {
	var n = len(dst)
	var dataLen = len(repeater.data)
	switch {
	case n <= dataLen:
		repeater.smallRead(dst)
	case n > dataLen:
		repeater.longRead(dst)
	}
	repeater.i = (repeater.i + n) % dataLen
	return n, nil
}

func (repeater *Repeater) smallRead(dst []byte) {
	var n = len(dst)
	var dataLen = len(repeater.data)
	var i = repeater.i
	copy(dst, repeater.data[i:])
	if i > 0 && n > dataLen-i {
		var offset = minInt(i, n)
		copy(dst[offset:], repeater.data[:i])
	}

}

func (repeater *Repeater) longRead(dst []byte) {
	var n = len(dst)
	var dataLen = len(repeater.data)
	var nStubs = n / dataLen
	for i := 0; i < nStubs; i++ {
		var offset = i * dataLen
		var readOffset = offset % dataLen
		var head = repeater.data[readOffset:]
		var tail = repeater.data[:readOffset]
		copy(dst[offset:], head)
		copy(dst[offset+dataLen-readOffset:], tail)
	}
	var tailStub = n % dataLen
	copy(dst[n-tailStub:], repeater.data[tailStub:])
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
