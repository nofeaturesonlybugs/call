package call

// Result is the result of invoking a function or method.
type Result struct {
	// If the function returns an error then Error is set to the returned error.
	//
	// If the function returns multiple error values then Error is set to the last error.
	//
	// Any error value will also exist in the Values member; it is provided
	// here as a convenience for checking for errors without having to inspect Values.
	Error error

	// Values holds the returned values.
	Values []interface{}
}
