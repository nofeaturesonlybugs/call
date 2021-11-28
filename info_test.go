package call_test

import (
	"net/http"
	"reflect"
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
		var values []reflect.Value
		values, _ = m.Args()
		m.Call(values)
	})
}

func Benchmark_MethodInfo_Call_PerformanceBaseline(b *testing.B) {
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
func Benchmark_MethodInfo_Call_PerformanceZero(b *testing.B) {
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
	Args := func() (Values []reflect.Value, Pointers []interface{}) {
		var value, ptr reflect.Value
		Values, Pointers = make([]reflect.Value, m.NumIn), make([]interface{}, m.NumIn)
		for k, typ := range m.InTypes {
			if k == 0 {
				Values[k] = m.Receiver.ReflectValue
				Pointers[k] = nil
				continue
			}
			ptr = reflect.New(typ)
			value = reflect.Indirect(ptr)
			Values[k] = value
			Pointers[k] = ptr.Interface()
		}
		return
	}
	//
	b.ResetTimer()
	var values []reflect.Value
	for k := 0; k < b.N; k++ {
		values, _ = Args()
		m.Call(values)
	}
}

func Benchmark_MethodInfo_Call_PerformanceWithInCache(b *testing.B) {
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
	Args := func() (Values []reflect.Value, Pointers []interface{}) {
		var value, ptr reflect.Value
		Values, Pointers = make([]reflect.Value, m.NumIn), make([]interface{}, m.NumIn)
		for k, typ := range m.InTypes {
			if k == 0 {
				Values[k] = m.Receiver.ReflectValue
				Pointers[k] = nil
				continue
			} else if v, ok := InCache[k]; ok {
				Values[k] = v
				Pointers[k] = nil
				continue
			}
			ptr = reflect.New(typ)
			value = reflect.Indirect(ptr)
			Values[k] = value
			Pointers[k] = ptr.Interface()
		}
		return
	}
	//
	b.ResetTimer()
	var values []reflect.Value
	for k := 0; k < b.N; k++ {
		values, _ = Args()
		m.Call(values)
	}
}

func Benchmark_MethodInfo_Call_PerformanceCurrent(b *testing.B) {
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
	var values []reflect.Value
	for k := 0; k < b.N; k++ {
		values, _ = m.Args()
		m.Call(values)
	}
}
