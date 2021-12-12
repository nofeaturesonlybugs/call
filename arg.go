package call

import (
	"reflect"
	"sync"
)

const (
	// argPoolAlloc specifies the allocation size of *Args returned from argPool.
	argPoolAlloc = 5
)

// argPool is a sync.Pool for *Args values.
var argPool = sync.Pool{
	New: func() interface{} {
		return &Args{
			Values:   make([]reflect.Value, argPoolAlloc),
			Pointers: make([]interface{}, argPoolAlloc),
		}
	},
}

// Arg describes a function or method argument by its type T, its index N, and if it can be
// known or calculated in advance its value V.
type Arg struct {
	// Argument index.
	N int
	// Argument type.
	T reflect.Type
	// If reusable then the reusable reflect.Value
	V reflect.Value
}

// Args is created by calling Args() on a Func or a Method.
//
// Args contains arguments as a pair of slices.  The Values slice represents
// the arguments as []reflect.Value while the Pointers slice represents
// addressable values that can be passed to decoders or unmarshalers to
// populate the arguments dynamically.
//
// Not all Values have an associated pointer; when a value with index K does not
// have a pointer Pointers[k] is nil.
type Args struct {
	Values   []reflect.Value
	Pointers []interface{}
}

// Reset ensures the Values and Pointers slices have enough capacity for N elements.
func (args *Args) Reset(N int) {
	if N > cap(args.Values) {
		args.Values, args.Pointers = make([]reflect.Value, N), make([]interface{}, N)
	}
}
