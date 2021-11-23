# circularbuffer
Circular Buffer in Go/Golang with some math methods.

Returns an error until `minSize` elements are reached. Overrides old values when `maxSize` is reached.

- Dynamic size (min, max)
- Average
- Median
- Quantile
- Min/Max

### Example

```go
package main

import "github.com/sklinkert/circularbuffer"

func main {
  minSize := 5
  maxSize := 10
  cb := circularbuffer(minSize, maxSize)
  for i := 1; i < 12; i++ {
		cb.Insert(float64(i))
  }

  value, err := cb.Median()
  value, err = cb.Min()
  value, err = cb.Max()
  value, err = cb.Percentile(0.8)
}
```

Feel free to add further methods if you think they could be helpful.
