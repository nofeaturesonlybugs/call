package call

// MethodResult is the result of invoking a method via MethodInfo.Call().
type MethodResult struct {
	// If the method returns an error then Error is set to the returned error.
	//
	// If the method returns multiple error values then Error is set to the last error.
	//
	// Any error value will also exist in the Values member; it is provided
	// here as a convenience for checking a method's error without inspecting
	// Values.
	Error error

	// Values holds the returned values.
	Values []interface{}
}
