package call

import (
	"reflect"
	"sync"
)

// TypeInfoCache
type TypeInfoCache interface {
	// Stat accepts an arbitrary variable and returns the associated Methods structure where
	// Methods.Receiver references V.
	Stat(V interface{}) Methods
	// StatType is similar to Stat except it accepts a reflect.Type and the returned Methods
	// has a Receiver that is the zero value for T.
	StatType(T reflect.Type) Methods
}

// TypeCache is a global TypeInfoCache.
var TypeCache = NewTypeInfoCache()

// NewTypeInfoCache creates a new TypeInfoCache.
func NewTypeInfoCache() TypeInfoCache {
	return &typeInfoCache{
		cache: &sync.Map{},
	}
}

// typeInfoCache is the implementation of a TypeInfoCache for this package.
type typeInfoCache struct {
	cache *sync.Map
}

// Stat accepts an arbitrary variable and returns the associated Methods structure.
func (me *typeInfoCache) Stat(V interface{}) Methods {
	cp := me.StatType(reflect.TypeOf(V)).Copy()
	cp.Receiver.Rebind(V)
	return cp
}

// StatType is similar to Stat except it accepts a reflect.Type and the returned Methods
// has a Receiver that is the zero value for T.
func (me *typeInfoCache) StatType(T reflect.Type) Methods {
	if T == nil {
		return Methods{}
	}
	if rv, ok := me.cache.Load(T); ok {
		return rv.(Methods)
	}
	//
	V := reflect.Zero(T)
	//
	receiver := &Receiver{
		Value:        V.Interface(),
		ReflectValue: V,
	}
	rv := Methods{
		Receiver: receiver,
		Methods:  []MethodInfo{},
	}
	//
	num := T.NumMethod()
	if num == 0 {
		return rv
	}
	rv.Methods = make([]MethodInfo, num)
	for k := 0; k < num; k++ {
		method := T.Method(k)
		numIn, numOut := method.Type.NumIn(), method.Type.NumOut()
		inKinds := make([]reflect.Kind, numIn)
		inTypes, outTypes := make([]reflect.Type, numIn), make([]reflect.Type, numOut)
		inCacheArgs, inCreateArgs := []Arg{}, []Arg{}
		for k := 0; k < numIn; k++ {
			in := method.Type.In(k)
			inKinds[k] = in.Kind()
			inTypes[k] = in
			//
			// Certain types+kinds are stored in the InCache member of MethodInfo.
			if inKinds[k] == reflect.Interface {
				inCacheArgs = append(inCacheArgs, Arg{N: k, T: in, V: reflect.Indirect(reflect.New(in))})
			} else if k != 0 {
				inCreateArgs = append(inCreateArgs, Arg{N: k, T: in})
			}
		}
		for k := 0; k < numOut; k++ {
			out := method.Type.Out(k)
			outTypes[k] = out
		}
		//
		info := MethodInfo{
			Receiver:     receiver,
			Name:         method.Name,
			Method:       method,
			NumIn:        numIn,
			InCacheArgs:  inCacheArgs,
			InCreateArgs: inCreateArgs,
			InKinds:      inKinds,
			InTypes:      inTypes,
			NumOut:       numOut,
			OutTypes:     outTypes,
		}
		rv.Methods[k] = info
	}
	//
	me.cache.Store(T, rv)
	//
	return rv
}
