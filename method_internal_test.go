package call

import (
	"reflect"
	"sync"
	"testing"

	"github.com/nofeaturesonlybugs/call/examples"
)

func Benchmark_Method_Call_Baseline(b *testing.B) {
	var talk examples.HTTP
	instance := Stat(talk)
	m, err := instance.Methods.Named("Handler")
	if err != nil {
		b.Fatal(err)
	}

	//
	// An implementation of Method.CreateArgs that is pure dumb and uses none of the performance
	// struct members.
	// Performance hit: Interface values are always recreated and not obtained from InCache.
	CreateArgs := func() *Args {
		var value, ptr reflect.Value
		rv := &Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		for k, typ := range m.InTypes {
			if k == 0 {
				rv.Values[k] = instance.receiverValue
				rv.Pointers[k] = nil
				continue
			}
			ptr = reflect.New(typ)
			value = reflect.Indirect(ptr)
			rv.Values[k] = value
			rv.Pointers[k] = ptr.Interface()
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *Args
	for k := 0; k < b.N; k++ {
		args = CreateArgs()
		m.Call(args)
	}
}

func Benchmark_Method_Call_WithInCache(b *testing.B) {
	var talk examples.HTTP
	instance := Stat(talk)
	m, err := instance.Methods.Named("Handler")
	if err != nil {
		b.Fatal(err)
	}

	//
	// An implementation of Method.Args used an InCache (type map[int]reflect.Value) of reused
	// reflect.Value when calling Args().
	// Performance gain: Interface values are not needlessly recreated.
	// Performance hit:	 Accessing m.InCache can be needlessly slow.
	//
	// NB: InCache is removed from the Method struct.  To retain this benchmark we'll recreate
	// an equivalent map here.
	InCache := map[int]reflect.Value{}
	for _, arg := range m.InCache {
		InCache[arg.N] = arg.V
	}
	CreateArgs := func() *Args {
		var value, ptr reflect.Value
		rv := &Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		for k, typ := range m.InTypes {
			if k == 0 {
				rv.Values[k] = instance.receiverValue
				rv.Pointers[k] = nil
				continue
			} else if v, ok := InCache[k]; ok {
				rv.Values[k] = v
				rv.Pointers[k] = nil
				continue
			}
			ptr = reflect.New(typ)
			value = reflect.Indirect(ptr)
			rv.Values[k] = value
			rv.Pointers[k] = ptr.Interface()
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *Args
	for k := 0; k < b.N; k++ {
		args = CreateArgs()
		m.Call(args)
	}
}

func Benchmark_Method_Call_UsingArgs(b *testing.B) {
	var talk examples.HTTP
	instance := Stat(talk)
	m, err := instance.Methods.Named("Handler")
	if err != nil {
		b.Fatal(err)
	}

	//
	// An implementation of Method.Args that uses InCreate and InCache
	// as slices for argument creation.
	//
	// Performance gain: No map lookups for argument indexes or type.
	// Performance hit: Lots of allocation hits the garbage collector.
	CreateArgs := func() *Args {
		var V reflect.Value
		rv := &Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		rv.Values[0], rv.Pointers[0] = instance.receiverValue, nil
		for _, arg := range m.InCreate {
			V = reflect.New(arg.T)
			rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
		}
		for _, arg := range m.InCache {
			rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *Args
	for k := 0; k < b.N; k++ {
		args = CreateArgs()
		m.Call(args)
	}
}

func Benchmark_Method_Call_UsingPool(b *testing.B) {
	var talk examples.HTTP
	instance := Stat(talk)
	m, err := instance.Methods.Named("Handler")
	if err != nil {
		b.Fatal(err)
	}

	//
	pool := &sync.Pool{
		New: func() interface{} {
			return &Args{
				Values:   []reflect.Value{},
				Pointers: []interface{}{},
			}
		},
	}
	//
	// An implementation of Method.Args that uses InCreate and InCache alongside
	// sync.Pool to alleviate burden on garbage collector.
	CreateArgs := func() *Args {
		var V reflect.Value
		rv := pool.Get().(*Args)
		if m.NumIn > cap(rv.Values) {
			rv.Values, rv.Pointers = make([]reflect.Value, m.NumIn), make([]interface{}, m.NumIn)
		}
		rv.Values, rv.Pointers = rv.Values[:m.NumIn], rv.Pointers[:m.NumIn]
		rv.Values[0], rv.Pointers[0] = instance.receiverValue, nil
		for _, arg := range m.InCreate {
			V = reflect.New(arg.T)
			rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
		}
		for _, arg := range m.InCache {
			rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *Args
	for k := 0; k < b.N; k++ {
		args = CreateArgs()
		m.Call(args)
		pool.Put(args)
	}
}
