package call_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func Test_Method_Call(t *testing.T) {
	t.Run("HTTP.Handler", func(t *testing.T) {
		var talk examples.HTTP
		instance := call.Stat(talk)
		m, _ := instance.Methods.Named("Handler")
		//
		args := m.Args()
		m.Call(args)
	})
	t.Run("ManyArgs.Many", func(t *testing.T) {
		var many examples.ManyArgs
		instance := call.Stat(many)
		m, _ := instance.Methods.Named("Many")
		//
		args := m.Args()
		m.Call(args)
	})
}

func ExampleMethods_Named() {
	var A examples.MapSession
	var am call.Method
	var err error

	a := call.Stat(A)
	if am, err = a.Methods.Named("Get"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(am.Pretty())

	if am, err = a.Methods.Named("MethodDoesNotExist"); err != nil {
		fmt.Println(err)
		return
	}

	// Output: Get (examples.MapSession, string) interface {}
	// not found
}

func Benchmark_Method_Call_StandardBaseline(b *testing.B) {
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

func Benchmark_Method_Call_Current(b *testing.B) {
	var talk examples.HTTP
	instance := call.Stat(talk)
	m, err := instance.Methods.Named("Handler")
	if err != nil {
		b.Fatal(err)
	}
	//
	b.ResetTimer()
	var args *call.Args
	for k := 0; k < b.N; k++ {
		args = m.Args()
		m.Call(args)
	}
}

func Benchmark_Method_Call_ManyArgs(b *testing.B) {
	var many examples.ManyArgs
	instance := call.Stat(many)
	m, err := instance.Methods.Named("Many")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		args := m.Args()
		m.Call(args)
	}
}
