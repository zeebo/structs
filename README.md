# Structs

[![GoDoc](https://godoc.org/github.com/zeebo/structs?status.svg)](https://godoc.org/github.com/zeebo/structs)
[![Sourcegraph](https://sourcegraph.com/github.com/zeebo/structs/-/badge.svg)](https://sourcegraph.com/github.com/zeebo/structs?badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/zeebo/structs)](https://goreportcard.com/report/github.com/zeebo/structs)

## Usage

#### type Option

```go
type Option interface {
	// contains filtered or unexported methods
}
```

Option controls the operation of a Decode.

#### type Result

```go
type Result struct {
	Error   error
	Used    map[string]struct{}
	Missing map[string]struct{}
	Broken  map[string]struct{}
}
```

Result contains information about the result of a Decode.

#### func  Decode

```go
func Decode(input map[string]interface{}, output interface{}, opts ...Option) Result
```
Decode takes values out of input and stores them into output, allocating as necessary.
