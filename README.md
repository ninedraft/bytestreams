# Bytestreams

- [Bytestreams](#bytestreams)
  - [ProxyWriter](#proxywriter)
    - [Usage](#usage)

## ProxyWriter

[![godoc](https://godoc.org/github.com/ninedraft/bytestreams?status.svg)](https://godoc.org/github.com/ninedraft/bytestreams/proxywriter)

Package: [`github.com/ninedraft/bytestreams/proxywriter`](/proxywriter)

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
