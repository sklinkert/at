package assert

// True - asserts that passed expression is evaluated to true
func True(reporter interface{}, expression bool) {
	if !expression {
		reportError(reporter, &failedBoolCompMsg{want: true, got: false})
	}
}

// False - asserts that passed expression is evaluated to false
func False(reporter interface{}, expression bool) {
	if expression {
		reportError(reporter, &failedBoolCompMsg{want: false, got: true})
	}
}
