// Package assert utility for formating test failure messages
package assert

import (
	"fmt"
	"time"
)

const (
	failedEqualityFormat         = "\nEquality assertion failed:\n\t want: %s \n\t  got: %s"
	expectationsOnNil            = "\nEquality assertion failed:\n\t reason: expectations on nil is not permitted"
	failedInclusionFormat        = "\nInclusion assertion failed:\n\t substring: %s \n\t to be present in: %s"
	failedInclusionNoErrorFormat = "\nInclusion assertion failed:\n\t substring: %s \n\t to be present, but no error was received"
	expectationsOnEmptyPhrase    = "\nInclusion assertion failed:\n\t reason: expectations on empty phrase is not permitted"
)

type failedNoErrorCompMsg struct{ got error }
type failedPanicMsg struct{ want, got string }
type failedBoolCompMsg struct{ want, got bool }
type failedErrorCompMsg struct{ want, got error }
type failedIntCompMsg struct{ want, got int }
type failedFloatCompMsg struct{ want, got float64 }
type failedStrCompMsg struct{ want, got string }
type failedTimeCompMsg struct{ want, got time.Time }
type failedStrIncludeMsg struct{ expectedPhrase, got string }
type failedErrorIncludeMsg struct {
	expectedPhrase string
	got            error
}
type failedIsNilCompMsg struct{ got interface{} }

func (msg *failedPanicMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatPanic(msg.want), formatPanic(msg.got))
}

func (msg *failedBoolCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatBool(msg.want), formatBool(msg.got))
}

func (msg *failedNoErrorCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatErr(nil), formatErr(msg.got))
}

// Equality messages
func (msg *failedErrorCompMsg) String() string {
	if msg.want == nil {
		return expectationsOnNil
	}
	return fmt.Sprintf(failedEqualityFormat, formatErr(msg.want), formatErr(msg.got))
}

func (msg *failedIntCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatInt(msg.want), formatInt(msg.got))
}

func (msg *failedFloatCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatFloat(msg.want), formatFloat(msg.got))
}

func (msg *failedStrCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatStr(msg.want), formatStr(msg.got))
}

func (msg *failedTimeCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, formatTime(msg.want), formatTime(msg.got))
}

func (msg *failedIsNilCompMsg) String() string {
	return fmt.Sprintf(failedEqualityFormat, "Nil", formatInterface(msg.got))
}

// Inclusion
func (msg *failedStrIncludeMsg) String() string {
	if msg.expectedPhrase == "" {
		return expectationsOnEmptyPhrase
	}
	return fmt.Sprintf(failedInclusionFormat, formatStr(msg.expectedPhrase), formatStr(msg.got))
}

func (msg *failedErrorIncludeMsg) String() string {
	if msg.expectedPhrase == "" {
		return expectationsOnEmptyPhrase
	}
	if msg.got == nil {
		return fmt.Sprintf(failedInclusionNoErrorFormat, formatStr(msg.expectedPhrase))
	}
	return fmt.Sprintf(failedInclusionFormat, formatStr(msg.expectedPhrase), formatErr(msg.got))
}
