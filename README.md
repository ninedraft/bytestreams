# Bytestreams

![[godoc](https://godoc.org/github.com/ninedraft/bytestreams/pkg/bytestreams)](https://godoc.org/github.com/ninedraft/bytestreams/pkg/bytestreams?status.svg) ![[go report card](https://goreportcard.com/report/github.com/ninedraft/bytestreams)](https://goreportcard.com/badge/github.com/ninedraft/bytestreams) ![](https://img.shields.io/badge/license-Apache-blue) ![](https://img.shields.io/github/go-mod/go-version/ninedraft/bytestreams) ![[chat](https://gitter.im/go-bytestreams/community)](https://img.shields.io/gitter/room/ninedraft/bytestreams )

- [Bytestreams](#bytestreams)
  - [ProxyWriter](#proxywriter)
    - [Usage](#usage)

## ProxyWriter

Package: [`github.com/ninedraft/bytestreams/pkg/bytestreams`](/pkg/bytestreams)

ProxyWriter is an implementation of an *error writer*, described in Dave Cheney's blogpost [Eliminate error handling by eliminating errors](https://dave.cheney.net/2019/01/27/eliminate-error-handling-by-eliminating-errors).

### Usage

A typical usecase is an implementation of a `.WriteTo`-like methods:

```go
import "git.vwgroup.ru/kodix/builder/pkg/proxywr"

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
