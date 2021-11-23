// Package assert - private utility functions to convert stuff to string
package assert

import (
	"fmt"
	"time"
)

func formatBool(in bool) string {
	return fmt.Sprintf("%v", in)
}

func formatPanic(in string) string {
	if in == "" {
		return "no panic"
	}
	return fmt.Sprintf("panic with message '%s'", in)
}

func formatErr(in error) string {
	if in == nil {
		return "no error"
	}
	return fmt.Sprintf("error '%s'", in)
}

func formatInt(in int) string {
	return fmt.Sprint(in)
}

func formatFloat(in float64) string {
	return fmt.Sprint(in)
}

func formatStr(in string) string {
	return fmt.Sprintf("'%s'", in)
}

func formatTime(in time.Time) string {
	return fmt.Sprint(in)
}

func formatInterface(in interface{}) string {
	return fmt.Sprintf("%#v", in)
}
