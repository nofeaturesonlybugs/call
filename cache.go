package call

import (
	"reflect"
	"sync"
)

// TypeInfoCache inspects a value or a reflect.Type and returns an appropriate *Instance type.
type TypeInfoCache interface {
	// Stat accepts an arbitrary variable and returns a *Instance whose receiver is V.
	Stat(V interface{}) *Instance

	// StatType is similar to Stat except it accepts a reflect.Type and the returned *Instance
	// has a Receiver that is the zero value for T.
	StatType(T reflect.Type) *Instance
}

// TypeCache is a global TypeInfoCache.
var TypeCache = NewTypeInfoCache()

// Stat calls TypeCache.Stat() on the global TypeInfoCache.  It is provided as a convenience
// if you do not wish to maintain your own TypeInfoCache instance.
func Stat(value interface{}) *Instance {
	return TypeCache.Stat(value)
}

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

// Stat accepts an arbitrary variable and returns a *Instance whose receiver is V.
func (me *typeInfoCache) Stat(V interface{}) *Instance {
	if V == nil {
		return nil
	}
	cp := me.StatType(reflect.TypeOf(V)).Copy()
	cp.Rebind(V)
	return cp
}

// StatType is similar to Stat except it accepts a reflect.Type and the returned *Instance
// has a Receiver that is the zero value for T.
func (me *typeInfoCache) StatType(T reflect.Type) *Instance {
	if rv, ok := me.cache.Load(T); ok {
		return rv.(*Instance)
	}
	//
	V := reflect.Zero(T)
	//
	rv := &Instance{
		Methods:       []Method{},
		receiver:      V.Interface(),
		receiverType:  T,
		receiverValue: V,
	}
	//
	num := T.NumMethod()
	if num == 0 {
		return rv
	}
	rv.Methods = make([]Method, num)
	for k := 0; k < num; k++ {
		method := T.Method(k)
		//
		info := Method{
			instance: rv,
			Name:     method.Name,
			Method:   method,
			Func:     newFunc(method.Func, method.Func.Type()),
		}
		// InCreate[0] represents the receiver which we do not need to create.
		info.Func.InCreate = info.Func.InCreate[1:]
		//
		rv.Methods[k] = info
	}
	//
	me.cache.Store(T, rv)
	//
	return rv
}
