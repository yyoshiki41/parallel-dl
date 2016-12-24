# parallel-dl

parallel-dl is a library for downloading files in parallel.

[![godoc](https://godoc.org/github.com/yyoshiki41/parallel-dl?status.svg)](https://godoc.org/github.com/yyoshiki41/parallel-dl)
[![build](https://travis-ci.org/yyoshiki41/parallel-dl.svg?branch=master)](https://travis-ci.org/yyoshiki41/parallel-dl)
[![codecov](https://codecov.io/gh/yyoshiki41/parallel-dl/branch/master/graph/badge.svg)](https://codecov.io/gh/yyoshiki41/parallel-dl)
[![go report](https://goreportcard.com/badge/github.com/yyoshiki41/parallel-dl)](https://goreportcard.com/report/github.com/yyoshiki41/parallel-dl)

- Go 1.7 or newer

## Features

parallel-dl allows you to control these features:

- HTTP Timeouts
- Maximum number of concurrent requests
- Maximum number of errors before giving up the whole requests
- Maximum number of retries before giving up a request

## API

- `Download()`

That's all !

## Examples

```go
opt := &paralleldl.Options{
	Output:           "/path/to/download",
	MaxConcurrents:   2,
	MaxErrorRequests: 1,
	MaxAttempts:      4,
}
client, err := paralleldl.New(opt)
if err != nil {
	log.Fatal(err)
}

lists := []string{
	"http://example.com/file1",
	"http://example.com/file2",
	"http://example.com/file3",
	"http://example.com/file4",
}
errCnt := client.Download(lists)
```

#### Minimum Options

```go
opt := &paralleldl.Options{
	Output: "/path/to/download",
}
```

## License 
The MIT License

## Author
Yoshiki Nakagawa
