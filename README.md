
# Snappy compression for Go HTTP clients and servers

[![Documentation](https://godoc.org/github.com/bakins/snappy?status.svg)](http://godoc.org/github.com/bakins/snappy)
[![license](https://img.shields.io/github/license/bakins/snappy?maxAge=2592000)](hhttps://github.com/bakins/snappy/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/bakins/snappy)](https://goreportcard.com/report/github.com/bakins/snappy)

Add [snappy stream compression](https://godoc.org/github.com/golang/snappy) to Go HTTP clients and Servers.

## Example

### Client

```go
import (
    "net/http"

    "github.com/bakins/snappy"
)

func main() {
    client := &http.Client{
        Transport: snappy.Transport(),
    }

    // use client as normal
}
```

### Server 

```go
import (
    "net/http"

    "github.com/bakins/snappy"
)

func main() {

    server := http.Server{
        Handler: snappy.Handler(http.DefaultServeMux)
    }
    // use server as normal
}
```



