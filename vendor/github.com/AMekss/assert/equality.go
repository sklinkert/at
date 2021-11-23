// Package assert - makes sure that 2 things are the same
package assert

import (
	"time"
)

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// EqualErrors - asserts that specific error was produced
func EqualErrors(reporter interface{}, want, got error) {
	if want == nil || got == nil || want.Error() != got.Error() {
		reportError(reporter, &failedErrorCompMsg{want, got})
	}
}

// EqualInt - asserts that two integers are the same
func EqualInt(reporter interface{}, want, got int) {
	if want != got {
		reportError(reporter, &failedIntCompMsg{want, got})
	}
}

// EqualFloat32 - asserts that two floats are the same
func EqualFloat32(reporter interface{}, want, got float32) {
	if want != got {
		reportError(reporter, &failedFloatCompMsg{float64(want), float64(got)})
	}
}

// EqualFloat32Tol - asserts that two floats are the same,
// allowing for (relative) tolerance given as a parameter
func EqualFloat32Tol(reporter interface{}, want, got, relTol float32) {
	if want == 0.0 && got != 0.0 {
		reportError(reporter, &failedFloatCompMsg{float64(want), float64(got)})
	} else if abs(float64((want-got)/want)) > float64(relTol) {
		reportError(reporter, &failedFloatCompMsg{float64(want), float64(got)})
	}
}

// EqualFloat64 - asserts that two floats are the same
func EqualFloat64(reporter interface{}, want, got float64) {
	if want != got {
		reportError(reporter, &failedFloatCompMsg{want, got})
	}
}

// EqualFloat64Tol - asserts that two floats are the same,
// allowing for (relative) tolerance given as a parameter
func EqualFloat64Tol(reporter interface{}, want, got, relTol float64) {
	if want == 0.0 && got != 0.0 {
		reportError(reporter, &failedFloatCompMsg{want, got})
	} else if abs((want-got)/want) > relTol {
		reportError(reporter, &failedFloatCompMsg{want, got})
	}
}

// EqualStrings - asserts that two strings are equal
func EqualStrings(reporter interface{}, want, got string) {
	if want != got {
		reportError(reporter, &failedStrCompMsg{want, got})
	}
}

// EqualTime - asserts that two time.Time are the same
func EqualTime(reporter interface{}, want, got time.Time) {
	if !got.Equal(want) {
		reportError(reporter, &failedTimeCompMsg{want, got})
	}
}
