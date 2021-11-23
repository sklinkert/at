/*
Package ring provides a simple implementation of a ring buffer.
*/
package ring

import "math"

/*
The DefaultCapacity of an uninitialized Ring buffer.

Changing this value only affects ring buffers created after it is changed.
*/
var DefaultCapacity int = 10

/*
Type Ring implements a Circular Buffer.
The default value of the Ring struct is a valid (empty) Ring buffer with capacity DefaultCapacify.
*/
type Ring struct {
	head int // the most recent value written
	tail int // the least recent value written
	buff []float64
}

/*
Set the maximum size of the ring buffer.
*/
func (r *Ring) SetCapacity(size int) {
	r.checkInit()
	r.extend(size)
}

/*
Capacity returns the current capacity of the ring buffer.
*/
func (r Ring) Capacity() int {
	return len(r.buff)
}

/*
Enqueue a value into the Ring buffer.
*/
func (r *Ring) Enqueue(i float64) {
	r.checkInit()
	r.set(r.head+1, i)
	old := r.head
	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
}

/*
Dequeue a value from the Ring buffer.

Returns NaN if the ring buffer is empty.
*/
func (r *Ring) Dequeue() float64 {
	r.checkInit()
	if r.head == -1 {
		return math.NaN()
	}
	v := r.get(r.tail)
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

/*
Read the value that Dequeue would have dequeued without actually dequeuing it.

Returns NaN if the ring buffer is empty.
*/
func (r *Ring) Peek() float64 {
	r.checkInit()
	if r.head == -1 {
		return math.NaN()
	}
	return r.get(r.tail)
}

/*
Values returns a slice of all the values in the circular buffer without modifying them at all.
The returned slice can be modified independently of the circular buffer. However, the values inside the slice
are shared between the slice and circular buffer.
*/
func (r *Ring) Values() []float64 {
	if r.head == -1 {
		return []float64{}
	}
	arr := make([]float64, 0, r.Capacity())
	for i := 0; i < r.Capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

/**
*** Unexported methods beyond this point.
**/

// sets a value at the given unmodified index and returns the modified index of the value
func (r *Ring) set(p int, v float64) {
	r.buff[r.mod(p)] = v
}

// gets a value based at a given unmodified index
func (r *Ring) get(p int) float64 {
	return r.buff[r.mod(p)]
}

// returns the modified index of an unmodified index
func (r *Ring) mod(p int) int {
	return p % len(r.buff)
}

func (r *Ring) checkInit() {
	if r.buff == nil {
		r.buff = make([]float64, DefaultCapacity)
		for i := range r.buff {
			r.buff[i] = math.NaN()
		}
		r.head, r.tail = -1, 0
	}
}

func (r *Ring) extend(size int) {
	if size == len(r.buff) {
		return
	} else if size < len(r.buff) {
		r.buff = r.buff[0:size]
	}
	newb := make([]float64, size-len(r.buff))
	for i := range newb {
		newb[i] = math.NaN()
	}
	r.buff = append(r.buff, newb...)
}
