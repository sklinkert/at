package circularbuffer

import (
	"fmt"
	"sort"
)

type CircularBuffer struct {
	records       []float64
	sortedRecords []float64
	minSize       int
	maxSize       int
	index         int
	enoughData    bool
}

// New creates a new buffer and returns the reference.
// minSize defines the minimum number of elements until we want allow operating on the buffer (e.g. calling Median()).
// maxSize defines the maximum number of elements in our buffer until we start overriding the oldest values.
func New(minSize, maxSize int) *CircularBuffer {
	return &CircularBuffer{
		minSize:       minSize,
		maxSize:       maxSize,
		records:       make([]float64, minSize),
		sortedRecords: []float64{},
	}
}

// GetAll returns all elements
func (cb *CircularBuffer) GetAll() ([]float64, error) {
	if !cb.enoughData {
		return []float64{}, fmt.Errorf("not enough data, have %d, need %d", cb.index, cb.minSize)
	}
	return cb.records, nil
}

// Insert new value to buffer
func (cb *CircularBuffer) Insert(value float64) {
	if cb.enoughData && len(cb.records) < cb.maxSize {
		// between minSize and maxSize
		cb.records = append(cb.records, value)
	} else {
		cb.records[cb.index] = value
	}
	if cb.index+1 == cb.minSize {
		cb.enoughData = true
	}
	if cb.index+1 == cb.maxSize {
		cb.index = 0
	} else {
		cb.index++
	}

	// sort records
	toBeSortedRecords := make([]float64, len(cb.records))
	copy(toBeSortedRecords, cb.records)
	sort.Float64s(toBeSortedRecords)
	cb.sortedRecords = toBeSortedRecords
}

// Min returns the smallest stored value in ring buffer
func (cb *CircularBuffer) Min() (float64, error) {
	return cb.Quantile(0)
}

// Max returns the highest stored value in ring buffer
func (cb *CircularBuffer) Max() (float64, error) {
	return cb.Quantile(1)
}

// Median returns the median of stored values in ring buffer
func (cb *CircularBuffer) Median() (float64, error) {
	return cb.Quantile(0.5)
}

// Average returns the average of stored values in ring buffer
func (cb *CircularBuffer) Average() (float64, error) {
	if !cb.enoughData {
		return 0, fmt.Errorf("not enough data, have %d, need %d", cb.index, cb.minSize)
	}

	var total float64
	for _, record := range cb.sortedRecords {
		total += record
	}
	return total / float64(len(cb.sortedRecords)), nil
}

// Quantile returns the element of the ring buffer for the given quantile
func (cb *CircularBuffer) Quantile(quantile float64) (float64, error) {
	if !cb.enoughData {
		return 0, fmt.Errorf("not enough data, have %d, need %d", cb.index, cb.minSize)
	}
	if quantile > 1 || quantile < 0 {
		return 0, fmt.Errorf("quantile needs to be between 0 and 1, but is %f", quantile)
	}

	i := float64(len(cb.sortedRecords)) * quantile
	if i > 0 {
		i--
	}
	return cb.sortedRecords[int(i)], nil
}
