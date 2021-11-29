package call

import (
	"fmt"
	"reflect"
)

// Receiver holds information about the original type passed to Stat().
type Receiver struct {
	// The original argument passed to Stat().
	Value interface{}
	// The Value field as a reflect.Value type.
	ReflectValue reflect.Value
}

// Rebind sets the receiver to the new value.
//
// If the incoming value does not have the same type as the original receiver then a panic will occur.
func (r *Receiver) Rebind(in interface{}) {
	if r == nil {
		return
	}
	v := reflect.ValueOf(in)
	t, T := v.Type(), r.ReflectValue.Type()
	if t != T {
		panic(fmt.Sprintf("%T.Rebind expects same underlying type: original %T not compatible with incoming %T", r, T, t))
	}
	r.Value = in
	r.ReflectValue = v
}
