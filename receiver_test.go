package call_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func TestReceiver_Swap(t *testing.T) {
	chk := assert.New(t)
	var bob, sally *examples.Person
	bob = &examples.Person{
		Name: "Bob",
		Age:  40,
	}
	sally = &examples.Person{
		Name: "Sally",
		Age:  30,
	}
	{
		// Pointers
		r := &call.Receiver{
			Value:        bob,
			ReflectValue: reflect.ValueOf(bob),
		}
		chk.Equal(r.Value, bob)
		chk.Equal(r.ReflectValue.Interface(), bob)
		//
		r.Rebind(sally)
		chk.Equal(r.Value, sally)
		chk.Equal(r.ReflectValue.Interface(), sally)
	}
	{
		// Non-pointers
		r := &call.Receiver{
			Value:        *bob,
			ReflectValue: reflect.ValueOf(*bob),
		}
		chk.Equal(r.Value, *bob)
		chk.Equal(r.ReflectValue.Interface(), *bob)
		//
		r.Rebind(*sally)
		chk.Equal(r.Value, *sally)
		chk.Equal(r.ReflectValue.Interface(), *sally)
	}
}

func TestReceiver_SwapPanics(t *testing.T) {
	chk := assert.New(t)
	var bob *examples.Person
	var talk examples.Talker
	bob = &examples.Person{
		Name: "Bob",
		Age:  40,
	}
	r := &call.Receiver{
		Value:        bob,
		ReflectValue: reflect.ValueOf(bob),
	}
	var panicked bool
	defer func() {
		chk.True(panicked)
	}()
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	r.Rebind(talk)
}
