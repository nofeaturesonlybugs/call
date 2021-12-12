package call

import (
	"fmt"
	"reflect"
)

// Instance summarizes a type and its methods.
//
// Instance and each Method in Instance.Methods are bound to the same receiver; see
// the Rebind example for demonstration of rebinding the receiver.
type Instance struct {
	Methods Methods

	receiver      interface{}
	receiverType  reflect.Type
	receiverValue reflect.Value
}

// Copy creates a copy of the Instance object.
//
// Copy() followed by Rebind() will create a new *Instance that has a different receiver
// than the original.
//
// Further each method in Methods will have its *Func shallow copied to a new *Func instance.
// Mutating a Method's *Func in the copy does not affect the original.
func (m *Instance) Copy() *Instance {
	cp := &Instance{
		Methods:       append([]Method(nil), m.Methods...),
		receiver:      m.receiver,
		receiverType:  m.receiverType,
		receiverValue: m.receiverValue,
	}
	for k := range cp.Methods {
		cp.Methods[k].instance = cp
		//
		// Each method gets a copy of the embedded *Func
		f, fnew := cp.Methods[k].Func, &Func{}
		*fnew = *f
		cp.Methods[k].Func = fnew

	}
	return cp
}

// Rebind sets the receiver to the new value.
//
// If the incoming value does not have the same type as the original receiver then a panic will occur.
func (m *Instance) Rebind(in interface{}) {
	v, t := reflect.ValueOf(in), reflect.TypeOf(in)
	if t != m.receiverType {
		panic(fmt.Sprintf("%T.Rebind expects same underlying type: original %T not compatible with incoming %T", m, m.receiver, in))
	}
	m.receiver = in
	m.receiverValue = v
}
