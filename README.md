# Bytestreams

[![](https://godoc.org/github.com/ninedraft/bytestreams/pkg/bytestreams?status.svg)](https://godoc.org/github.com/ninedraft/bytestreams/pkg/bytestreams) [![](https://goreportcard.com/badge/github.com/ninedraft/bytestreams)](https://goreportcard.com/report/github.com/ninedraft/bytestreams) ![](https://img.shields.io/badge/license-Apache-blue) ![](https://img.shields.io/github/go-mod/go-version/ninedraft/bytestreams) [![](https://img.shields.io/gitter/room/ninedraft/bytestreams)](https://gitter.im/go-bytestreams/community) [![](https://img.shields.io/badge/golangci--lint-report-blueviolet)](https://golangci.com/r/github.com/ninedraft/bytestreams)

- [Bytestreams](#bytestreams)
  - [ProxyWriter](#proxywriter)
    - [ProxyWriter usage](#proxywriter-usage)
  - [StackWriter](#stackwriter)
    - [StackWriter usage](#stackwriter-usage)

## ProxyWriter

Package: [`github.com/ninedraft/bytestreams/pkg/bytestreams`](/pkg/bytestreams)

ProxyWriter is an implementation of an *error writer*, described in Dave Cheney's blogpost [Eliminate error handling by eliminating errors](https://dave.cheney.net/2019/01/27/eliminate-error-handling-by-eliminating-errors).

### ProxyWriter usage

A typical usecase is an implementation of a `.WriteTo`-like methods:

```go
import "github.com/ninedraft/bytestreams/pkg/bytestreams"

type User struct {
    Name             string
    Topics           []string
    AdditionalData   []byte
}

func (user *User) WriteTo(wr io.Writer) (int64, error) {
    var pw = proxywr.New(wr)
    pw.Printf("user: ", user.Name)
    pw.Printf("topics: ")
    pw.PrintSlice(",", user.Topics...)
    pw.Printf("\n")
    _, _ = pw.Write(user.AdditionalData)
    return pw.Result()
}
```

## StackWriter

Package: [`github.com/ninedraft/bytestreams/pkg/bytestreams`](/pkg/bytestreams)

StackWriter is a chain of bytestream converters. It can be used to build chains like `json.Encoder`->`tar.Writer`->`gzip.Writer`->`bufio.Writer`->`net.Conn` with automatic `.Flush` and `.Close` method handling.

### StackWriter usage

```go
import (
    "github.com/ninedraft/bytestreams/pkg/bytestreams"
    "net"
)

type User struct {
    Name             string
    Topics           []string
    AdditionalData   []byte
}


var conn, errDial = net.Dial("tcp", "$ADDR")
//...

var wr = bytestreams.NewStackWriter(conn).
    Push(
        // all encoders will be flushed and closed automagically
        func(wr io.Writer) io.Writer { return bufio.Writer(wr) },
        func(wr io.Writer) io.Writer { return gzip.Writer(wr) },
        func(wr io.Writer) io.Writer { return tar.Writer(wr) },
    )
defer wr.Close()

var encoder = json.Encoder()
// ...

```
