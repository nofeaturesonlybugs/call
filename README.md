[![Go Reference](https://pkg.go.dev/badge/github.com/nofeaturesonlybugs/call.svg)](https://pkg.go.dev/github.com/nofeaturesonlybugs/call)
[![Go Report Card](https://goreportcard.com/badge/github.com/nofeaturesonlybugs/call)](https://goreportcard.com/report/github.com/nofeaturesonlybugs/call)
[![Build Status](https://app.travis-ci.com/nofeaturesonlybugs/call.svg?branch=master)](https://app.travis-ci.com/nofeaturesonlybugs/call)
[![codecov](https://codecov.io/gh/nofeaturesonlybugs/call/branch/master/graph/badge.svg)](https://codecov.io/gh/nofeaturesonlybugs/call)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Package `call` is a small wrapper around the official reflect package that eases dynamic function or method calls.

`call` can be useful for rigging routes to handlers on Go types in a dynamic fashion. An example of that will follow but let's first see some examples of `call` in action.

## A Useless Case

```go
fn := func(str string, num int) {
    fmt.Printf("str=%v num=%v\n", str, num)
}

f := call.StatFunc(fn)
f.Call(f.Args())

// prints:
// str= num=0
```

The call to `StatFunc` returns a `Func` type that can be used to create function arguments and then invoke the function as seen by `f.Call(f.Args())`.

When `Args()` creates arguments it creates zero values for the argument types.

## Setting Argument Values

The previous example is somewhat useless because the function is called with a zero-value for each argument. Now we'll set the argument values using the `Pointers` field of `Args`:

```go
// Same function as before.
fn := func(str string, num int) {
    fmt.Printf("str=%v num=%v\n", str, num)
}

f := call.StatFunc(fn)
// Args contains two slices that give us access to the created arguments.
args := f.Args()
for k := range args.Values {
    // args.Pointers are pointers to the arguments.
    pointer := args.Pointers[k]
    switch p := pointer.(type) {
    case *string:
        *p = "Hi!"
    case *int:
        *p = 42
    }
    // args.Values are reflect.Value of the argument.
    value := args.Values[k].Interface()
    fmt.Printf("%T %v %T\n", value, value, pointer)
}
f.Call(args)

// prints:
// string Hi! *string
// int 42 *int
// str=Hi! num=42
```

## Struct Arguments

The `Pointers` field is also useful for unmarshaling data into function arguments:

```go
type Request struct {
    Str string `json:"str"`
    Num int    `json:"num"`
}
fn := func(req Request) {
    fmt.Printf("str=%v num=%v\n", req.Str, req.Num)
}

data := []byte(`{"str" : "Hi!", "num" : 42}`)
f := call.StatFunc(fn)
args := f.Args()
// For brevity we unmarshal straight into args.Pointers[0]
if err := json.Unmarshal(data, args.Pointers[0]); err != nil {
    fmt.Println(err)
    return
}
f.Call(args)

// prints:
// str=Hi! num=42
```

## Interface Arguments

When an argument is an interface I its value is `I(nil)` and its pointer is also `nil`:

```go
// Interfaces are always passed as nil.
fn := func(w http.ResponseWriter) {
    fmt.Println(w)
}

f := call.StatFunc(fn)
args := f.Args()
// When an argument represents an interface I its Values is an I(nil)
// and its Pointers is nil.
fmt.Println(args.Values[0].Interface(), args.Pointers[0])
f.Call(args)

// prints:
// <nil> <nil>
// <nil>
```

## Interface Arguments & Pruning <sup>The Beginnings of an http.Handler</sup>

Since interface types are provided as nil values by `Args()` you may wish to configure the `*Func` to stop managing such types. You do this by calling `PruneIn()`, which accepts a variadic list of `reflect.Type`:

```go
// In order to prune a type we need its reflect.Type.  Let's pretend we're writing
// a more general purpose http.Handler and want to prune http.ResponseWriter
// and *http.Request from types created via `Args()`:
TypeRequest := reflect.TypeOf((*http.Request)(nil))
// This is the idomatic way to get the type of a nil interface.
TypeResponseWriter := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

fn := func( w http.ResponseWriter, req *http.Request, req SomeStructType ) {
}

// We're going to take the standard http.Handler signature and dynamically invoke fn
handler := func( w http.ResponseWriter, req *http.Request ) {
    f := call.StatFunc(fn)
    pruned := f.PruneIn(TypeRequest, TypeResponseWriter)
    args := f.Args()
    //
    // Before invoking f we should see if we can provide any pruned arguments:
    for _, arg := range pruned {
        switch arg.T {
            case TypeRequest:
                args.Values[arg.N] = reflect.ValueOf(req)
            case TypeResponseWriter:
                args.Values[arg.N] = reflect.ValueOf(w)
        }
    }
    //
    // NB   A more useful handler would potentially unmarshal req.Body
    //      into args.Pointers that could accept it.
    //
    f.Call(args)
}
```

## A Better http.Handler

Let's take some ideas from the previous snippet and create a `http.Handler` factory:

```go
TypeRequest := reflect.TypeOf((*http.Request)(nil))
TypeResponseWriter := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

// Factory accepts a function and turns it into an http.Handler.
Factory := func(opaque interface{}) http.Handler {
	f := call.StatFunc(opaque)
	pruned := f.PruneIn(TypeRequest, TypeResponseWriter)
	//
	// The created handler does not represent a "production-ready" http.Handler but it does
	// demonstrate how "package call" can be used to:
	//	+ invoke end-of-chain handlers with adhoc or variable signatures
	//	+ how to unmarshal and provide data to the handler arguments
	//		(by unmarshaling application/json requests, for example)
	fn := func(w http.ResponseWriter, req *http.Request) {
		args := f.Args()
		// Before invoking f we should see if we can provide any pruned arguments:
		for _, arg := range pruned {
			switch arg.T {
			case TypeRequest:
				args.Values[arg.N] = reflect.ValueOf(req)
			case TypeResponseWriter:
				args.Values[arg.N] = reflect.ValueOf(w)
			}
		}
		//
		// If the request is application/json we will unmarshal into any arguments
		// that are struct.
		// NB:  An intelligent handler factory would have examined f.InCreate and possibly set a
		//		hasJSON=true|false flag and could theoretically skip this logic block if the
		//		end-of-chain handler doesn't have targets for JSON data.
		if req.Header.Get("Content-Type") == "application/json" {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, arg := range f.InCreate {
				if arg.T.Kind() != reflect.Struct {
					continue
				}
				if err = json.Unmarshal(body, args.Pointers[arg.N]); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		//
		// NB:  This handler doesn't do anything with any return values.  A better handler
		// 		factory would probably make use of any error returned or possibly accept
		//		some type of result and then write to the response appropriately.
		f.Call(args)
	}
	return http.HandlerFunc(fn)
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
Login := func(w http.ResponseWriter, post LoginRequest) {
	fmt.Fprintf(w, "%v", post)
}
Logout := func(w http.ResponseWriter) {
	fmt.Fprint(w, "Logged out!")
}

mux := http.NewServeMux()
mux.Handle("/login", Factory(Login))
mux.Handle("/logout", Factory(Logout))

// /login
w := httptest.NewRecorder()
w.Body = &bytes.Buffer{}
req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"test","password":"s3cr3t"}`))
req.Header.Set("Content-Type", "application/json")
mux.ServeHTTP(w, req)
fmt.Println(w.Body.String())

// /logout
w = httptest.NewRecorder()
w.Body = &bytes.Buffer{}
req = httptest.NewRequest(http.MethodPost, "/logout", nil)
mux.ServeHTTP(w, req)
fmt.Println(w.Body.String())

// prints:
// {test s3cr3t}
// Logged out!
```

## API Consistency and Breaking Changes

I am making a very concerted effort to break the API as little as possible while adding features or fixing bugs. However this software is currently in a pre-1.0.0 version and breaking changes _are_ allowed under standard semver. As the API approaches a stable 1.0.0 release I will list any such breaking changes here and they will always be signaled by a bump in _minor_ version.

-   0.1.x ⭢ 0.2.0
    -   Several types have been renamed to be more ergonomic:
        -   `Methods` renamed to `Instance`
        -   `MethodInfo` renamed to `Method`
        -   `MethodResult` renamed to `Result`
        -   Fields `InCacheArgs` and `InCreateArgs` have had the `Args` suffix dropped and are now simply `InCache` and `InCreate`.
        -   `Receiver` type dropped entirely; the `Rebind()` function now exists on `Instance`.
    -   `call` now supports invoking methods on types or regular functions. To support this a new type `Func` has been introduced. `Func` was created by pulling several fields out of `Method` (previously `MethodInfo`). `Method` retains access to this extracted information by embedding `*Func`; in other words `Func` is for calling regular functions, `Method` is for calling functions that have receivers, and `Method` is a superset of `Func`.
    -   Added a new `type Methods []Method` which has a helper function for finding a method by name; note that this `Methods` type is not the same nor is it compatible with `Methods` type in the previous release.
