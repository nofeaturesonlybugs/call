package call_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func ExampleStat() {
	var talk examples.Talker
	methods := call.Stat(talk)
	for _, method := range methods.Methods {
		// Second return value are pointers-to-args and aren't needed here.
		args, _ := method.Args()
		fmt.Println(method.Pretty())
		result := method.Call(args)
		if result.Error != nil {
			fmt.Println(result.Error)
		}
	}

	// Output: Error (examples.Talker, examples.Response, *examples.Request) error
	// examples.Talker made an error
	// Goodbye (examples.Talker, *examples.Request, struct { StringField string; NumField int })
	// Hello (examples.Talker, examples.Response, *examples.Request) (bool, error)
}

func TestStat_TypeHasNoMethods(t *testing.T) {
	chk := assert.New(t)
	var s string
	methods := call.Stat(s)
	chk.Empty(methods.Methods)
}

func TestStat_SwapReceiver(t *testing.T) {
	chk := assert.New(t)
	//
	var talk *examples.Talker
	var methods call.Methods
	for k := 0; k < 100; k++ {
		talk = new(examples.Talker)
		methods = call.TypeCache.Stat(talk)
		chk.Equal(talk, methods.Receiver.Value)
		chk.Equal(talk, methods.Receiver.ReflectValue.Interface())
	}
}

func BenchmarkStat(b *testing.B) {
	var talk examples.Talker
	for k := 0; k < b.N; k++ {
		call.Stat(talk)
	}
}

func ExampleMethodInfo_Call_swapReceiver() {
	var bob, sally *examples.Person
	bob = &examples.Person{
		Name: "Bob",
		Age:  40,
	}
	sally = &examples.Person{
		Name: "Sally",
		Age:  30,
	}
	methods := call.Stat(bob)
	for _, m := range methods.Methods {
		values, _ := m.Args()
		rv := m.Call(values)
		for _, v := range rv.Values {
			fmt.Println(v)
		}
	}
	methods.Receiver.Rebind(sally)
	for _, m := range methods.Methods {
		values, _ := m.Args()
		rv := m.Call(values)
		for _, v := range rv.Values {
			fmt.Println(v)
		}
	}
	// Output: Hello!  My name is Bob and I am 40 year(s) old.
	// Hello!  My name is Sally and I am 30 year(s) old.
}
