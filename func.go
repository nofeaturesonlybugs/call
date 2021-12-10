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

// Func represents a single function call and facilitates creating arguments
// for the func as well as invoking it.
type Func struct {
	// Func is the reflect.Value representing the function.
	Func reflect.Value

	// NumIn, InKinds, and InTypes describe the function's arguments.
	//
	// NumIn is the number of arguments.
	//
	// InKinds and InTypes are slices of reflect.Kind and reflect.Type representing the
	// argument list.
	NumIn   int
	InKinds []reflect.Kind
	InTypes []reflect.Type

	// InCreate is a deterministic list of arguments to create during Args().
	//
	// Args() must return *all* arguments required to successfully invoke Call(); however
	// not all arguments need to be created during Args().
	//
	// For example interface arguments are taken from cache and not created (see InCache).
	//
	// PruneIn() removes entries from InCreate in cases where the caller of Args() and
	// Call() knows it will be providing arguments of certain types.
	InCreate []Arg

	// InCache is a cache of arguments that are reused during calls to Args(); in other words
	// multiple calls to Args() will receive the same reflect.Value in the argument position.
	//
	// Consider an argument that is an interface I.  Args() needs to create a reflect.ValueOf(I).
	// Args() also has no way of knowing how to create the concrete type to satisfy I.  Therefore
	// Args() will return a reflect.Value representing I(nil) -- the nil value of the interface.
	// I(nil) is effectively useless to the caller and the work creating such a value is wasted.
	//
	// Therefore such values are stored in InCache and returned in calls to Args as appropriate.
	InCache []Arg

	// NumOut is the length of the OutTypes slice.
	NumOut int
	// OutTypes is the type-list of values returned by calling the function.
	OutTypes []reflect.Type
}

// StatFunc accepts an arbitrary function and returns an associated Func.
func StatFunc(f interface{}) *Func {
	T := reflect.TypeOf(f)
	F := reflect.ValueOf(f)
	return newFunc(F, T)
}

// newFunc creates a Func struct from the given reflect type which must represent a function
// or a panic occurs.
func newFunc(F reflect.Value, T reflect.Type) *Func {
	if T.Kind() != reflect.Func {
		panic("function argument expected")
	}
	numIn, numOut := T.NumIn(), T.NumOut()
	inKinds := make([]reflect.Kind, numIn)
	inTypes, outTypes := make([]reflect.Type, numIn), make([]reflect.Type, numOut)
	inCache, inCreate := []Arg{}, []Arg{}
	for k := 0; k < numIn; k++ {
		in := T.In(k)
		inKinds[k] = in.Kind()
		inTypes[k] = in
		//
		// Certain types+kinds are stored in the InCache member of Func.
		if inKinds[k] == reflect.Interface {
			inCache = append(inCache, Arg{N: k, T: in, V: reflect.Indirect(reflect.New(in))})
		} else {
			inCreate = append(inCreate, Arg{N: k, T: in})
		}
	}
	for k := 0; k < numOut; k++ {
		out := T.Out(k)
		outTypes[k] = out
	}
	//
	return &Func{
		Func:     F,
		NumIn:    numIn,
		InCache:  inCache,
		InCreate: inCreate,
		InKinds:  inKinds,
		InTypes:  inTypes,
		NumOut:   numOut,
		OutTypes: outTypes,
	}
}

// Args returns an *Args type where its Values and Pointers members are populated with
// the necessary values to call the function via Call().
//
// The return value and its members are pooled resources that will be reclaimed during Call().
//	args := f.Args()
//		// + args (*Args) is pooled
//		// + args.Values slice is pooled
//		// + args.Pointers slice is pooled
//		// NB: Slice elements are **not** pooled; you can retain handles to
//		//     individual elements; i.e. args.Values[1] or args.Pointers[3].
//	f.Call(args)	// args, args.Values, & args.Pointers returned to pool.
//
// Some of the Values may be retrieved from InCache and shared across calls to Args() (see InCache member).
// When a Value from InCache is used its associated Pointers entry is set to nil.
//
// If you wish to populate or unmarshal into an argument you can use its index K in Pointers:
//	json.Unmarshal(data, args.Pointers[K])
//	// NB: If you experience panics while accessing args.Pointers[K] chances are:
//	//     1. Your index K exceeds the length of the Pointers slice or
//	//     2. You are unmarshaling into a type whose Pointers entry is nil, such as
//	//        an interface `type I interface {...}`
func (f *Func) Args() *Args {
	var V reflect.Value
	rv := argPool.Get().(*Args)
	rv.Reset(f.NumIn)
	rv.Values, rv.Pointers = rv.Values[:f.NumIn], rv.Pointers[:f.NumIn]
	for _, arg := range f.InCreate {
		V = reflect.New(arg.T)
		rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
	}
	for _, arg := range f.InCache {
		rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
	}
	return rv
}

// Call invokes the function described by Func; call Args() to obtain the arguments.
//	f := Stat(SomeFunc)
//	args := f.Args()
//	// NB: Args returns zero-value arguments; you probably want to mutate
//	//     them by populating them with data before the next line.
//	f.Call(args)
//
// During Call() the args are returned to the argument pool (see Args()).
func (f *Func) Call(args *Args) Result {
	var iface interface{}
	var result Result
	//
	defer func() {
		for k, max := 0, len(args.Values); k < max; k++ {
			args.Values[k], args.Pointers[k] = zeroReflectValue, nil
		}
		argPool.Put(args)
	}()
	//
	returns := f.Func.Call(args.Values)
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

// Pretty returns a string representing the func( args... ) return-value(s).
func (f *Func) Pretty() string {
	var args, returns []string
	for _, arg := range f.InTypes {
		args = append(args, arg.String())
	}
	for _, rv := range f.OutTypes {
		returns = append(returns, rv.String())
	}
	argstr, rvstr := strings.Join(args, ", "), strings.Join(returns, ", ")
	ro, rc := "", ""
	if f.NumOut == 1 {
		ro = " "
	} else if f.NumOut > 1 {
		ro, rc = " (", ")"
	}
	return fmt.Sprintf("func (%v)%v%v%v", argstr, ro, rvstr, rc)
}

// PruneIn searches both InCache and InCreate for the given types.  When a type is found
// in either InCache or InCreate it is removed from the slice and added to the return
// value.
//
// The purpose of PruneIn is to remove elements from InCache and InCreate slices that the
// caller or client code knows it will be providing when setting up arguments before
// Call().  There is no sense in Args() creating or providing arguments that the caller
// will replace.
//
// Correct usage of PruneIn will provide performance increases for code using this package.
func (f *Func) PruneIn(types ...reflect.Type) []Arg {
	var rv []Arg
	//
	prune := func(slice []Arg) []Arg {
		for _, T := range types {
			for k, size := 0, len(slice); k < size; size = len(slice) {
				arg := slice[k]
				if arg.T == T {
					rv = append(rv, arg)
					slice = append(slice[:k], slice[k+1:]...)
				} else {
					k++
				}
			}
		}
		return slice
	}
	f.InCache = prune(f.InCache)
	f.InCreate = prune(f.InCreate)
	return rv
}
