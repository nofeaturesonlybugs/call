package call

// Methods summarizes a type and its methods.
//
// Methods and each element in Methods.Methods are bound to the same receiver; see
// the Methods (SwapReceiver) example for an example of rebinding the receiver.
type Methods struct {
	Receiver *Receiver
	Methods  []MethodInfo
}

// Stat inspects a value in order to return a Methods type.  It is a shortcut to calling
// TypeCache.Stat() -- i.e. it invokes Stat() on the global TypeCache in this package.
func Stat(value interface{}) Methods {
	return TypeCache.Stat(value)
}

// Copy creates a copy of the Methods object.
//
// The *Receiver field is updated to a new struct pointer that, if modified, will cause
// method calls to occur on the new receiver without affecting the original.
func (m Methods) Copy() Methods {
	var receiver *Receiver
	//
	// If Methods was created by calling Stat(nil) then Receiver field is nil.
	if m.Receiver != nil {
		receiver = &Receiver{
			Value:        m.Receiver.Value,
			ReflectValue: m.Receiver.ReflectValue,
		}
	}
	//
	cp := Methods{
		Receiver: receiver,
		Methods:  append([]MethodInfo{}, m.Methods...),
	}
	for k := range cp.Methods {
		cp.Methods[k].receiver = cp.Receiver
	}
	return cp
}
