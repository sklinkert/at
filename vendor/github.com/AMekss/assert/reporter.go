// Package assert private utility functions for error reporting
package assert

import (
	"fmt"
	"reflect"
	"runtime"
)

type testingT interface {
	Errorf(format string, args ...interface{})
}

func reportError(reporterIface interface{}, msg fmt.Stringer) {
	source := ""
	_, file, line, ok := runtime.Caller(2)
	if ok {
		source = fmt.Sprintf("\n\t line: %s:%d", file, line)
	}

	reporterFrom(reporterIface)("%s %s", msg, source)
}

func reporterFrom(reporterIface interface{}) func(format string, args ...interface{}) {
	kind := reflect.ValueOf(reporterIface).Kind()
	switch kind {
	case reflect.Func:
		if reporter, ok := reporterIface.(func(format string, args ...interface{})); ok {
			return reporter
		}
		panic("provided reporter function doesn't implement `func(format string, args ...interface{})`")
	case reflect.Interface, reflect.Ptr:
		if reporter, ok := reporterIface.(testingT); ok {
			return reporter.Errorf
		}
		panic("provided interface doesn't implement `Fatalf(format string, args ...interface{})`")
	default:
		panic(fmt.Sprintf("don't know how to handle %s", kind))
	}
}
