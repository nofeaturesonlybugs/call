// Package call is a small wrapper around the official reflect package that facilitates type inspection for the purpose of
// invoking methods on those types.
//
// This package is primarily useful if you want to create a router of sorts that maps strings to methods on Go types
// serving as handlers.  Most such routers require:
//	1. explicit registration of handlers and
//	2. uniform handler signatures.
//
// A common example of the above requirements are seen in the net/http package with http.Handler:
//	func ServeHttp(ResponseWriter, *Request)
//
// There are two negative side effects that result from handlers adhering to a common signature.
//
// The first is that many handlers may start or end with common boilerplate.  It is undesirable to
// copy+paste the same 5 or 10 lines across many handlers.  Middlewares alleviate this pain to a degree
// but in the instances where middlewares extract data and push it further along the chain we are left
// with the key=value store of Context, which reverts our data to interface{} and forces type assertion
// at a later point in the request handling.
//
// Now consider the following types and methods:
//	type AuthLoginRequest struct {
//		Username string `post:"username"`
//		Password string `post:"password"`
//	}
//
//	type Auth struct {
//		DB *sql.DB
//	}
//
//	func(auth *Auth) HTTPLogin(post AuthLoginRequest, sess session.Session) (Result, error) {...}
//	func(auth *Auth) HTTPLogout(sess session.Session) (Result, error) {...}
//
// Those are good end-of-chain handlers.
//	1. they can be implicitly registered via a prefix=HTTP matching rule,
//	2. do not rely on unpacking interface{} values from a Context in an *http.Request,
//	3. are called only with the arguments they need in the format they need, and
//	3. are potentially easier to test.
//
// Installing those routes on a ServeMux can also be compatible with http.Handler:
//	auth := &Auth{
//		DB : db,
//	}
//	router := &Router{
//		Handlers : []interface{}{ auth, ... }, // Here ... can be any other types that have routes.
//	}
//	mux := Handle( "/", router.Handler() )
//
// Creating such a router is difficult.  It requires reflect and unfortunately reflect has caught a bad
// reputation in the Go community.  Code invoking reflect can become unwieldy and difficult to understand.
// Calls into reflect are also much slower than equivalent code not using reflect; therefore the amount
// of reflect calls needs to be carefully controlled and unnecessary calls avoided -- further complicating code
// involving reflect.
//
// This is where package call becomes useful.  It serves to decouple such sophisticated routers
// from the reflect calls required for the router to work.  It can be tested and benchmarked indepedently
// of where it is used.  Thus we can gain confidence that general purpose routers are working correctly
// and performing well.
//
// Performance
//
// This package is carefully benchmarked and some best efforts have been made to increase performance
// where possible.
//
// The TypeInfoCache is used to prevent overhead when creating Methods structures.  The Stat() function
// uses a global instace of TypeInfoCache.  In other words repeatedly calling Stat(V), Stat(V1), etc.
// where all values have the same type performs well.
//
// Invoking a method via MethodInfo.Args() and MethodInfo.Call() uses a shared memory pool.  In short the
// value returned by MethodInfo.Args() is a pooled resource and is reclaimed during MethodInfo.Call().  See
// the documentation for MethodInfo.Args() for more information.
//
package call
