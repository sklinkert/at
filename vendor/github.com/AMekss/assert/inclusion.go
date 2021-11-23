// Package assert - makes sure that one thing includes other
package assert

import (
	"strings"
)

// ErrorIncludesMessage - asserts that received error includes particular string in the message
func ErrorIncludesMessage(reporter interface{}, expectedPhrase string, got error) {
	if expectedPhrase == "" || got == nil || !strings.Contains(got.Error(), expectedPhrase) {
		reportError(reporter, &failedErrorIncludeMsg{expectedPhrase, got})
	}
}

// IncludesString - asserts that string includes substring
func IncludesString(reporter interface{}, expectedPhrase, got string) {
	if expectedPhrase == "" || !strings.Contains(got, expectedPhrase) {
		reportError(reporter, &failedStrIncludeMsg{expectedPhrase, got})
	}
}
