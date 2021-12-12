// Package call is a small wrapper around the official reflect package that facilitates invocation of
// functions or methods on types.
//
// Consider using this package if you are building some type of router that has one of
// the following requirements:
//	1. Implicit registration of routes-to-handlers.
//	2. Non-uniform handler signatures.
//
// Performance Caching and Pooling
//
// The following strategies are employed by this package to help offset the performance penalties
// of using reflection.
//
// The TypeInfoCache uses a goroutine-safe cache for returning information about types T that represent
// an object or type with methods.  You may create your own instance of a TypeInfoCache or use the global
// instance; note that Stat() uses the global instance.
//
// To trim the number of allocations during Func.Args() and Func.Call() a memory pool is used.  The *Args
// instance returned by Func.Args() is a pooled resource along with its Values and Pointers slices.  These
// resources are returned to the pool during Func.Call() -- after which the *Args type and its
// Values and Pointers fields should no longer be accessed by the caller.  Note however that the slice elements
// are not pooled and the caller of Args+Call can maintain handles to those individual elements if need be.
//
// The Method type also has methods Args() and Call() that are implemented by an embedded Func.  Therefore
// the notes about pooling also apply to Method.Args() and Method.Call().
//
package call
