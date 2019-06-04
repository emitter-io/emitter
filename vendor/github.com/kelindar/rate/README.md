# RateLimit

[![GoDoc](https://godoc.org/github.com/kelindar/rate?status.png)](http://godoc.org/github.com/kelindar/rate)
[![Go Report Card](https://goreportcard.com/badge/github.com/kelindar/rate)](https://goreportcard.com/report/github.com/bsm/ratelimit)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

This package contains a simple, thread-safe Go rate-limiter. This is a fork of [bsm/ratelimit](https://github.com/bsm/ratelimit) which was originally inspired by [Antti Huima's algorithm](http://stackoverflow.com/a/668327).

### Example

```go
package main

import (
  "github.com/kelindar/rate"
  "log"
)

func main() {
  // Create a new rate-limiter, allowing up-to 10 calls per second
  rl := rate.New(10, time.Second)

  for i:=0; i<20; i++ {
    if rl.Limit() {
      fmt.Println("DOH! Over limit!")
    } else {
      fmt.Println("OK")
    }
  }
}
```

### Documentation

Full documentation is available on [GoDoc](http://godoc.org/github.com/kelindar/rate)
