package assert

import (
	"fmt"
	"reflect"
)

// NoError - asserts that no error was produced
func NoError(reporter interface{}, got error) {
	if got != nil {
		reportError(reporter, &failedNoErrorCompMsg{got})
	}
}

// Panic - asserts that code caused panic with specific message
func Panic(reporter interface{}, withMessage string) {
	r := recover()
	msg := &failedPanicMsg{want: withMessage}
	if r == nil {
		reportError(reporter, msg)
	} else {
		msg.got = fmt.Sprint(r)
		if msg.want != msg.got {
			reportError(reporter, msg)
		}
	}
}

// IsNil - asserts that provided interface has nil value
func IsNil(reporter interface{}, got interface{}) {
	if !isNil(got) {
		reportError(reporter, &failedIsNilCompMsg{got})
	}
}

// https://github.com/goccy/go-reflect/blob/b725637422e43c3e5556f930b7f5c0d592d0dbaa/reflect.go#L886-L892
// returns true if value is nil or false otherwise
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	switch reflect.ValueOf(v).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return reflect.ValueOf(v).IsNil()
	}

	return false
}
