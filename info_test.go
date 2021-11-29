package call_test

import (
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func Test_MethodInfo_Call(t *testing.T) {
	t.Run("HTTP.Handler", func(t *testing.T) {
		var talk examples.HTTP
		methods := call.Stat(talk)
		var m call.MethodInfo
		for _, m = range methods.Methods {
			if m.Name == "Handler" {
				break
			}
		}
		//
		args := m.Args()
		m.Call(args)
	})
	t.Run("ManyArgs.Many", func(t *testing.T) {
		var many examples.ManyArgs
		methods := call.Stat(many)
		var m call.MethodInfo
		for _, m = range methods.Methods {
			if m.Name == "Many" {
				break
			}
		}
		//
		args := m.Args()
		m.Call(args)
	})
}

func Benchmark_MethodInfo_Call_StandardBaseline(b *testing.B) {
	var w *http.ResponseWriter
	var req **http.Request
	var sess *examples.Session
	var talk examples.HTTP
	for k := 0; k < b.N; k++ {
		// Calling Handler(nil, nil, nil, struct{...}) is a ridiculously unfair comparison.
		// We create these values to indicate some small cost in instantiating and providing useful
		// values.
		w = new(http.ResponseWriter)
		req = new(*http.Request)
		sess = new(examples.Session)
		talk.Handler(*w, *req, *sess, struct {
			Username string "form:\"username\""
			Password string "form:\"password\""
		}{
			Username: "hi",
			Password: "bye",
		})
	}
}
func Benchmark_MethodInfo_Call_Baseline(b *testing.B) {
	var talk examples.HTTP
	methods := call.Stat(talk)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Handler" {
			break
		}
	}
	//
	// An implementation of MethodInfo.Args that is pure dumb and uses none of the performance
	// struct members.
	// Performance hit: Interface values are always recreated and not obtained from InCache.
	Args := func() *call.Args {
		var value, ptr reflect.Value
		rv := &call.Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		for k, typ := range m.InTypes {
			if k == 0 {
				rv.Values[k] = methods.Receiver.ReflectValue
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
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = Args()
		m.Call(args)
	}
}

func Benchmark_MethodInfo_Call_WithInCache(b *testing.B) {
	var talk examples.HTTP
	methods := call.Stat(talk)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Handler" {
			break
		}
	}
	//
	// An implementation of MethodInfo.Args used an InCache (type map[int]reflect.Value) of reused
	// reflect.Value when calling Args().
	// Performance gain: Interface values are not needlessly recreated.
	// Performance hit:	 Accessing m.InCache can be needlessly slow.
	//
	// NB: InCache is removed from the MethodInfo struct.  To retain this benchmark we'll recreate
	// an equivalent map here.
	InCache := map[int]reflect.Value{}
	for _, arg := range m.InCacheArgs {
		InCache[arg.N] = arg.V
	}
	Args := func() *call.Args {
		var value, ptr reflect.Value
		rv := &call.Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		for k, typ := range m.InTypes {
			if k == 0 {
				rv.Values[k] = methods.Receiver.ReflectValue
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
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = Args()
		m.Call(args)
	}
}

func Benchmark_MethodInfo_Call_UsingArgs(b *testing.B) {
	var talk examples.HTTP
	methods := call.Stat(talk)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Handler" {
			break
		}
	}
	//
	// An implementation of MethodInfo.Args that uses InCreateArgs and InCacheArgs
	// as slices for argument creation.
	//
	// Performance gain: No map lookups for argument indexes or type.
	// Performance hit: Lots of allocation hits the garbage collector.
	Args := func() *call.Args {
		var V reflect.Value
		rv := &call.Args{
			Values:   make([]reflect.Value, m.NumIn),
			Pointers: make([]interface{}, m.NumIn),
		}
		rv.Values[0], rv.Pointers[0] = methods.Receiver.ReflectValue, nil
		for _, arg := range m.InCreateArgs {
			V = reflect.New(arg.T)
			rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
		}
		for _, arg := range m.InCacheArgs {
			rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = Args()
		m.Call(args)
	}
}

func Benchmark_MethodInfo_Call_UsingPool(b *testing.B) {
	var talk examples.HTTP
	methods := call.Stat(talk)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Handler" {
			break
		}
	}
	//
	pool := &sync.Pool{
		New: func() interface{} {
			return &call.Args{
				Values:   []reflect.Value{},
				Pointers: []interface{}{},
			}
		},
	}
	//
	// An implementation of MethodInfo.Args that uses InCreateArgs and InCacheArgs alongside
	// sync.Pool to alleviate burden on garbage collector.
	Args := func() *call.Args {
		var V reflect.Value
		rv := pool.Get().(*call.Args)
		if m.NumIn > cap(rv.Values) {
			rv.Values, rv.Pointers = make([]reflect.Value, m.NumIn), make([]interface{}, m.NumIn)
		}
		rv.Values, rv.Pointers = rv.Values[:m.NumIn], rv.Pointers[:m.NumIn]
		rv.Values[0], rv.Pointers[0] = methods.Receiver.ReflectValue, nil
		for _, arg := range m.InCreateArgs {
			V = reflect.New(arg.T)
			rv.Values[arg.N], rv.Pointers[arg.N] = V.Elem(), V.Interface()
		}
		for _, arg := range m.InCacheArgs {
			rv.Values[arg.N], rv.Pointers[arg.N] = arg.V, nil
		}
		return rv
	}
	//
	b.ResetTimer()
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = Args()
		m.Call(args)
		pool.Put(args)
	}
}

func Benchmark_MethodInfo_Call_Current(b *testing.B) {
	var talk examples.HTTP
	methods := call.Stat(talk)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Handler" {
			break
		}
	}
	//
	b.ResetTimer()
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = m.Args()
		m.Call(args)
	}
}

func Benchmark_MethodInfo_Call_ManyArgs(b *testing.B) {
	var many examples.ManyArgs
	methods := call.Stat(many)
	var m call.MethodInfo
	for _, m = range methods.Methods {
		if m.Name == "Many" {
			break
		}
	}
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		args := m.Args()
		m.Call(args)

	}
}
