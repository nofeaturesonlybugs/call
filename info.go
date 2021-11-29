package call

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	// zeroReflectValue is a global re-usable instance of a zero reflect.Value
	zeroReflectValue reflect.Value
)

// MethodInfo contains information about a single method on a Go type.
//
// Each instance of a MethodInfo has an internal value representing the
// method receiver.  This receiver is the original value passed to Stat()
// when creating a Methods type.
type MethodInfo struct {
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

	// Receiver links to the original value used to create the MethodInfo structure.
	//
	// When Stat() is called and multiple MethodInfo structs are created they all point
	// to the same Receiver struct.  This allows for swapping of the receiver across
	// all methods by assigning new values to Receiver.Value and Receiver.ReflectValue.
	receiver *Receiver
}

// Args returns an *Args type where its Values and Pointers members are populated with
// the necessary values to call the method via MethodInfo.Call().
//
// The returned *Args is a pooled resource that is reclaimed during MethodInfo.Call().  Namely the
// struct Args, its Values slice, and its Pointers slice are pooled.  The elements inside Values and
// Pointers are not pooled.  During MethodInfo.Call() the *Args object is reclaimed to the Pool and
// the caller should no longer access the Values or Pointers slices.
//
// Values[0] is the Receiver.Value field.  Pointers[0] is set to nil.  You should not manipulate, swap
// or mutate this value without great care.
//
// Some of the Values may be retrieved from the InCacheArgs struct member and shared across calls to Args();
// see the documentation of InCacheArgs for reasoning and explanation.  When a Value from InCacheArgs is used
// its associated Pointers entry is set to nil.
func (m MethodInfo) Args() *Args {
	var V reflect.Value
	rv := argPool.Get().(*Args)
	rv.Reset(m.NumIn)
	rv.Values, rv.Pointers = rv.Values[:m.NumIn], rv.Pointers[:m.NumIn]
	rv.Values[0], rv.Pointers[0] = m.receiver.ReflectValue, nil
	for _, arg := range m.InCreateArgs {
		V = reflect.New(arg.T)
		rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
	}
	for _, arg := range m.InCacheArgs {
		rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
	}
	return rv
}

// Calls calls the method described by MethodInfo.
//
// An appropriate *Args type can be obtained via MethodInfo.Args().
//
// During Call() the args are returned to the argument pool.  After Call() returns the caller
// should no longer access the *Args object or its fields.  See MethodInfo.Args() for more
// explanation.
func (m MethodInfo) Call(args *Args) MethodResult {
	var iface interface{}
	var result MethodResult
	//
	defer func() {
		for k, max := 0, len(args.Values); k < max; k++ {
			args.Values[k], args.Pointers[k] = zeroReflectValue, nil
		}
		argPool.Put(args)
	}()
	//
	returns := m.Method.Func.Call(args.Values)
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
