package bytestreams_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/ninedraft/bytestreams/pkg/bytestreams"
)

// TODO: add a variadic buf reader
func TestRepeater_Read(test *testing.T) {
	const testdata = "doot "

	var testRead = func(name string, toRead int, modifiers ...func(io.Reader) io.Reader) {
		test.Run(name, func(test *testing.T) {
			var repeater = bytestreams.NewRepeater([]byte(testdata))
			var buf = &bytes.Buffer{}
			var re io.Reader = io.LimitReader(repeater, int64(toRead))
			for _, modify := range modifiers {
				re = modify(re)
			}
			var _, errCopy = io.Copy(buf, re)
			if errCopy != nil {
				test.Fatalf("unexpected error %v", errCopy)
			}
			var expected string
			switch {
			case toRead < len(testdata):
				expected = testdata[:toRead]
			default:
				var n = len(testdata)
				expected = strings.Repeat(testdata, toRead/n) + testdata[:toRead%n]
			}
			if expected != buf.String() {
				test.Fatalf("expected %q, got %q", expected, buf)
			}
		})
	}

	testRead("short read 1/2", len(testdata)/2)
	testRead("short read 1/2 single byte reader", len(testdata)/2,
		iotest.OneByteReader)

	testRead("short read 2/3", len(testdata)*2/3)
	testRead("short read 2/3", len(testdata)*2/3,
		iotest.OneByteReader)

	testRead("long read 10", 10*len(testdata))
	testRead("long read 10 single byte reader", 10*len(testdata),
		iotest.OneByteReader)
}