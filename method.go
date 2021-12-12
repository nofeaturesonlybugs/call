package call

import (
	"reflect"
)

// Methods is a slice of Method.
type Methods []Method

// Named returns the Method with the following name or ErrNotFound.
func (m Methods) Named(name string) (Method, error) {
	for _, elem := range m {
		if elem.Name == name {
			return elem, nil
		}
	}
	return Method{}, ErrNotFound
}

// Method contains information about a single method on a Go type.
//
// Each instance of Method has an internal *Instance pointer that ties it
// to its receiver.  Calling Instance.Rebind() updates these internal pointers
// to the new receiver.
type Method struct {
	// Name is the method name.
	Name string

	// Method is the reflect.Method value.
	Method reflect.Method

	// A Method is a superset of a Func.
	*Func

	// The Instance containing the receiver we are tied to.
	instance *Instance
}

// Args returns an *Args type where its Values and Pointers members are populated with
// the necessary values to call the method via Method.Call().
//
// Args calls down to Func.Args() but then sets the 0 index of Values and Pointers to
// the correct receiver and nil respectively.
func (m Method) Args() *Args {
	args := m.Func.Args()
	args.Values[0], args.Pointers[0] = m.instance.receiverValue, nil
	return args
}

// Pretty returns a string representing the method-name( args... ) return-value(s).
func (m Method) Pretty() string {
	// Get Pretty from Func but replace leading 4 (func) with our method name.
	return m.Name + m.Func.Pretty()[4:]
}
