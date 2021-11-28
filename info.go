package call

import (
	"fmt"
	"reflect"
	"strings"
)

// Arg describes a method argument by its type T, its index N, and if it can be
// known or calculated in advance its value V.
type Arg struct {
	// Argument index.
	N int
	// Argument type.
	T reflect.Type
	// If reusable then the reusable reflect.Value
	V reflect.Value
}

// MethodInfo contains information about a single method on a Go type.
type MethodInfo struct {
	// Receiver links to the original value used to create the MethodInfo structure.
	//
	// When Stat() is called and multiple MethodInfo structs are created they all point
	// to the same Receiver struct.  This allows for swapping of the receiver across
	// all methods by assigning new values to Receiver.Value and Receiver.ReflectValue.
	Receiver *Receiver

	// Name is the method name.
	Name string

	// Method is the reflect.Method value.
	Method reflect.Method

	// NumIn, InKinds, and InTypes describe the method's arguments.
	//
	// NumIn is the number of arguments.
	//
	// InKinds and InTypes are slices of reflect.Kind and reflect.Type representing the
	// argument list.
	NumIn   int
	InKinds []reflect.Kind
	InTypes []reflect.Type
	//
	// InCreateArgs is the list of arguments that are created during Args().  Not all arguments
	// are created during Args(); for example Interface args are taken from cache (see InCacheArgs).
	// By caching the types we know we will create we can perform a faster slice iteration over
	// InCreateArgs versus iterating InTypes and testing if the type is one we create.
	InCreateArgs []Arg
	//
	// InCacheArgs is a set of reusable reflect.Values for arguments that can not be created
	// during Args() and exists for performance.
	//
	// Consider an argument that is an interface I.  Args() needs to create a reflect.ValueOf(I).
	// Args() also has no way of knowing how to create the concrete type to satisfy I.  Therefore
	// Args() will return a reflect.Value representing I(nil) -- the nil value of the interface.
	// I(nil) is effectively useless to the caller and the work creating such a value is wasted.
	//
	// Therefore such values are stored in InCacheArgs and returned in calls to Args as appropriate.
	InCacheArgs []Arg

	// NumOut is the length of the OutTypes slice.
	NumOut int
	// OutTypes is the type-list of values returned by calling the method.
	OutTypes []reflect.Type
}

// Args creates the arguments needed to call a method and returns them as a pair of slices.
//
// The first slice represents the arguments as values suitable for passing to MethodInfo.Call().
// The second slice represents pointers to the arguments and is provided as a convienience for
// populating the arguments with tools or packages that require addressable values.
//
// Values[0] is the Receiver.Value field.  Pointers[0] is set to nil.
//
// Some of the Values may be retrieved from the InCache struct member and shared across calls to Args();
// see the documentation of InCache for reasoning and explanation.  When a Value from InCache is used
// its associated Pointers entry is set to nil.
func (m MethodInfo) Args() (Values []reflect.Value, Pointers []interface{}) {
	var V reflect.Value
	Values, Pointers = make([]reflect.Value, m.NumIn), make([]interface{}, m.NumIn)
	Values[0], Pointers[0] = m.Receiver.ReflectValue, nil
	for _, arg := range m.InCreateArgs {
		V = reflect.New(arg.T)
		Values[arg.N], Pointers[arg.N] = V.Elem(), V.Interface()
	}
	for _, arg := range m.InCacheArgs {
		Values[arg.N], Pointers[arg.N] = arg.V, nil
	}
	return
}

// Calls calls the method described by MethodInfo.
func (m MethodInfo) Call(in []reflect.Value) MethodResult {
	var iface interface{}
	var result MethodResult
	//
	returns := m.Method.Func.Call(in)
	for _, rv := range returns {
		iface = rv.Interface()
		result.Values = append(result.Values, iface)
		if err, ok := iface.(error); ok {
			result.Error = err
		}
	}
	//
	return result
}

// Pretty returns a string representing the method-name( args... ) return-value(s).
func (m MethodInfo) Pretty() string {
	var args, returns []string
	for _, arg := range m.InTypes {
		args = append(args, arg.String())
	}
	for _, rv := range m.OutTypes {
		returns = append(returns, rv.String())
	}
	argstr, rvstr := strings.Join(args, ", "), strings.Join(returns, ", ")
	ro, rc := "", ""
	if m.NumOut == 1 {
		ro = " "
	} else if m.NumOut > 1 {
		ro, rc = " (", ")"
	}
	return fmt.Sprintf("%v (%v)%v%v%v", m.Name, argstr, ro, rvstr, rc)
}
