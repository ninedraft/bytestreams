package bytestreams_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"testing/iotest"

	"github.com/ninedraft/bytestreams/pkg/bytestreams"
)

// TODO: add a variadic buf reader
func TestRepeater_Read(test *testing.T) {
	// "Harry Potter and the Methods of Rationality"
	// http://www.hpmor.com/chapter/1
	var testdata = []byte(`Every inch of wall space is covered by a bookcase.
	Each bookcase has six shelves, going almost to the ceiling.
	Some bookshelves are stacked to the brim with hardback books: 
	science, maths, history, and everything else. 
	Other shelves have two layers of paperback science fiction, 
	with the back layer of books propped up on old tissue boxes or lengths of wood, 
	so that you can see the back layer of books above the books in front. 
	And it still isn't enough. Books are overflowing onto the tables and the sofas 
	and making little heaps under the windows.`)

	var suite = func(name string, makeRepeater func(data []byte) io.Reader) {
		test.Run(name, func(test *testing.T) {

			test.Run("no op reads, then full read", func(test *testing.T) {
				var repeater = makeRepeater(testdata)
				var noops = 2*len(testdata) + 1
				for i := 0; i < noops; i++ {
					repeater.Read([]byte{})
				}
				readCheck(test, repeater, testdata)
			})

			test.Run("full read", func(test *testing.T) {
				var repeater = makeRepeater(testdata)
				readCheck(test, repeater, testdata)
			})

			test.Run("half read", func(test *testing.T) {
				var repeater = makeRepeater(testdata)
				var toRead = len(testdata) / 2
				readCheck(test, repeater, testdata[:toRead])
			})

			test.Run("3/2 read", func(test *testing.T) {
				var n = len(testdata)
				var half = n / 2
				var expected = append(testdata[:n:n], testdata[:half]...)
				var repeater = makeRepeater(testdata)
				readCheck(test, repeater, expected)
			})
		})
	}

	suite("plain repeater", func(data []byte) io.Reader {
		return bytestreams.NewRepeater(data)
	})

	suite("one byte reader", func(data []byte) io.Reader {
		var repeater = bytestreams.NewRepeater(data)
		return iotest.OneByteReader(repeater)
	})

	var chunks []int
	const minChunk = 2
	var maxChunk = 5 * len(testdata)

	for i := 1; i <= maxChunk/minChunk; i++ {
		var chunk = i * len(testdata) / minChunk
		chunks = append(chunks, chunk)
	}

	suite("variable chunk reader", func(data []byte) io.Reader {
		var repeater = bytestreams.NewRepeater(data)
		return newVCR(repeater, chunks...)
	})

	// no more then N failed test calls
	// to avoid avalanche of logs
	var toFail = 5
	for _, chunk := range chunks {
		if test.Failed() && toFail > 0 {
			toFail--
		}
		suite(fmt.Sprintf("%d bytes chunk reader", chunk),
			func(data []byte) io.Reader {
				var repeater = bytestreams.NewRepeater(data)
				return &chunkReader{
					Re:    repeater,
					Chunk: chunk,
				}
			})
	}
}

func readCheck(test *testing.T, src io.Reader, expected []byte) {
	var expectedN = len(expected)
	var buf = &bytes.Buffer{}
	buf.Grow(expectedN)
	var protected = io.LimitReader(src, int64(expectedN))
	var _, err = io.Copy(buf, protected)
	if err != nil {
		test.Errorf("reading bytes from src: %v", err)
		return
	}
	var got = buf.Bytes()
	var report, failed = foundDiscord(expected, got)
	if failed {
		test.Errorf("%s", report)
		return
	}
}

func foundDiscord(expected, got []byte) (string, bool) {
	if bytes.Equal(expected, got) {
		return "", false
	}
	var boundary = minInt(len(expected), len(got))
	for i := 0; i < boundary; i++ {
		var exp = expected[i]
		var g = got[i]
		if exp != g {
			var cutAt = minInt(16, boundary)
			var expectedChunk = "..." + string(cut(expected[i:], cutAt)) + "..."
			var gotChunk = "..." + string(cut(got[i:], cutAt)) + "..."
			return fmt.Sprintf("offset %d: expected %q, got %q", i, expectedChunk, gotChunk), true
		}
	}
	return "", false
}

type chunkReader struct {
	Chunk int
	Re    io.Reader

	buf bytes.Buffer
}

func (cr *chunkReader) Read(dst []byte) (int, error) {
	var bufLen = cr.buf.Len()
	if bufLen == 0 {
		cr.buf.Reset()
		cr.buf.Grow(cr.Chunk)
		var buf = cr.buf.Bytes()[:cr.Chunk]
		var n, err = cr.Re.Read(buf)
		cr.buf.Write(buf[:n])
		if err != nil && err != io.EOF {
			return copy(dst, buf[:n]), err
		}
	}
	var n, err = cr.buf.Read(dst)
	switch err {
	case io.EOF:
		cr.buf.Reset()
		return cr.Read(dst[n:])
	default:
		return n, err
	}
}

type variableChunkReader struct {
	Chunks []int
	Re     io.Reader

	buf     bytes.Buffer
	current int
}

func newVCR(re io.Reader, chunks ...int) *variableChunkReader {
	return &variableChunkReader{
		Chunks:  chunks,
		Re:      re,
		current: -1,
	}
}

func (vcr *variableChunkReader) Read(dst []byte) (int, error) {
	var bufLen = vcr.buf.Len()
	if bufLen == 0 {
		vcr.buf.Reset()
		var nextChunk = (vcr.current + 1) % len(vcr.Chunks)
		var chunk = vcr.Chunks[nextChunk]
		vcr.buf.Grow(chunk)
		var buf = vcr.buf.Bytes()[:chunk]
		var n, err = vcr.Re.Read(buf)
		vcr.buf.Write(buf[:n])
		if err != nil && err != io.EOF {
			return copy(dst, buf[:n]), err
		}
	}
	var n, err = vcr.buf.Read(dst)
	switch err {
	case io.EOF:
		return vcr.Read(dst[n:])
	default:
		return n, err
	}
}

func cut(src []byte, n int) []byte {
	if n > len(src) {
		return src
	}
	return src[:n]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
