package call_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func ExampleStat() {
	var talk examples.Talker
	instance := call.Stat(talk)
	for _, method := range instance.Methods {
		args := method.Args()
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
	instance := call.Stat(s)
	chk.Empty(instance.Methods)
}

func ExampleInstance_Copy() {
	// The point of this example is to demonstrate that a copy of an instance can have its methods
	// mutated without affecting the original.

	var A examples.MapSession
	var am, cpm call.Method

	a := call.Stat(A)
	am, _ = a.Methods.Named("Get") // error ignored for brevity
	fmt.Println(am.Pretty())

	cp := a.Copy()
	cpm, _ = cp.Methods.Named("Get") // error ignored for brevity

	// Now we'll prune the key from our copy.
	cpm.PruneIn(reflect.TypeOf(""))

	// Calling am is fine because all arguments are provided by am.Args():
	args := am.Args()
	am.Call(args)

	// Calling cpm will panic because cpm.Args() is no longer creating the string argument:
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("cpm panics!")
			}
		}()
		args := cpm.Args()
		cpm.Call(args)
	}()

	// But if we provide the argument there's no problem.
	args = cpm.Args()
	args.Values[1] = reflect.ValueOf("key")
	cpm.Call(args)

	// Output: Get (examples.MapSession, string) interface {}
	// cpm panics!
}

func ExampleInstance_Rebind() {
	var bob, sally *examples.Person
	bob = &examples.Person{
		Name: "Bob",
		Age:  40,
	}
	sally = &examples.Person{
		Name: "Sally",
		Age:  30,
	}
	instance := call.Stat(bob)
	for _, m := range instance.Methods {
		args := m.Args()
		rv := m.Call(args)
		for _, v := range rv.Values {
			fmt.Println(v)
		}
	}
	instance.Rebind(sally)
	for _, m := range instance.Methods {
		args := m.Args()
		rv := m.Call(args)
		for _, v := range rv.Values {
			fmt.Println(v)
		}
	}
	// Output: Hello!  My name is Bob and I am 40 year(s) old.
	// Hello!  My name is Sally and I am 30 year(s) old.
}

func ExampleInstance_Rebind_panic() {
	var bob *examples.Person
	var i int
	bob = &examples.Person{
		Name: "Bob",
		Age:  40,
	}
	instance := call.Stat(bob)
	for _, m := range instance.Methods {
		args := m.Args()
		rv := m.Call(args)
		for _, v := range rv.Values {
			fmt.Println(v)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Rebind panics because types are not the same.")
		}
	}()
	instance.Rebind(i)
	// Output: Hello!  My name is Bob and I am 40 year(s) old.
	// Rebind panics because types are not the same.
}
