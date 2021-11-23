# ring

[![GoDoc](https://godoc.org/github.com/falzm/golang-ring?status.svg)](https://godoc.org/github.com/falzm/golang-ring)

Package ring provides a simple implementation of a ring buffer.

This is a fork from original package [zfjagann/golang-ring](https://github.com/zfjagann/golang-ring) modified to store buffer values typed as `float64` numbers instead of `interface{}`.

## Usage

```go
package main

import (
	"fmt"

	"github.com/falzm/golang-ring"
)

func main() {
	r := ring.Ring{}
	r.SetCapacity(10)

	for i := 0.0; i <= 100.0; i++ {
		r.Enqueue(i)
	}

	fmt.Println(r.Values())
}
```

Output:

```
[91 92 93 94 95 96 97 98 99 100]
```