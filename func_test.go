package call_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
	"github.com/stretchr/testify/assert"
)

func ExampleFunc() {
	fn := func(str string, num int) {
		fmt.Printf("str=%v num=%v\n", str, num)
	}

	f := call.StatFunc(fn)
	f.Call(f.Args())

	// Output: str= num=0
}

func ExampleFunc_Args() {
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

	// Output: string Hi! *string
	// int 42 *int
	// str=Hi! num=42
}

func ExampleFunc_Args_struct() {
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

	// Output: str=Hi! num=42
}

func ExampleFunc_Args_interface() {
	// Interfaces are always passed as nil.
	//
	// The reason is there is no way for package call to know which concrete type is the appropriate
	// type to satisfy an interface present in an argument list.
	fn := func(w http.ResponseWriter) {
		fmt.Println(w)
	}

	f := call.StatFunc(fn)
	args := f.Args()
	// When an argument represents an interface I its Values is an I(nil)
	// and its Pointers is nil.
	fmt.Println(args.Values[0].Interface(), args.Pointers[0])
	f.Call(args)

	// Output: <nil> <nil>
	// <nil>
}

func ExampleStatFunc() {
	fn := func(req examples.Request, res examples.Response) {
		fmt.Printf("%T %v\n", req, req)
		if res != nil {
			fmt.Println("Second argument should have been nil.")
		} else {
			fmt.Println("Second argument was nil but that is expected because it is an interface.")
		}
	}

	f := call.StatFunc(fn)
	fmt.Println(f.Pretty())

	args := f.Args()
	f.Call(args)

	// Output: func (examples.Request, examples.Response)
	// examples.Request { }
	// Second argument was nil but that is expected because it is an interface.
}

func ExampleFunc_hTTPHandlerFactory() {
	TypeRequest := reflect.TypeOf((*http.Request)(nil))
	TypeResponseWriter := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()

	// Factory accepts a function and turns it into an http.Handler.
	Factory := func(opaque interface{}) http.Handler {
		// Stat the handler.
		f := call.StatFunc(opaque)
		//
		// If the handler has *http.Request or http.ResponseWriter arguments we want to "prune"
		// them from arguments created when calling f.Args().
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
				// NB: Originally used io.ReadAll() but removed to support Go <= 1.15
				//     It's just an example anyways.
				b := make([]byte, 2048)
				read, err := req.Body.Read(b)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				body := b[0:read]
				//
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

	// Output: {test s3cr3t}
	// Logged out!
}

func ExampleFunc_PruneIn() {
	fn := func(req examples.Request, store examples.Session) {
		fmt.Printf("%T %v\n", req, req)
		if store == nil {
			fmt.Println("nil store")
		} else if store != nil {
			store.Set("message", "Hello, World!")
		}
	}

	//
	// Create and call our function.
	f := call.StatFunc(fn)
	fmt.Println(f.Pretty())
	args := f.Args()
	f.Call(args)

	// Since examples.Session is an interface Args() will always create a
	// nil instance.  This is not very useful to us so we'll remove examples.Session
	// type from the list of arguments returned or created.
	T := reflect.TypeOf((*examples.Session)(nil)).Elem() // The idomatic way to get reflect.Type of an interface.
	pruned := f.PruneIn(T)

	// Now when we create args the examples.Session value is not provided.
	// If we call Call() we expect a panic:
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Panic!")
			}
		}()
		args = f.Args()
		f.Call(args)
	}()

	// Since Args() no longer provides the session we must provide it ourselves:
	//
	// NB: The for...range is the generic way to identify the positioning of the pruned arg.
	// For this example we could have also done:
	//	args.Values[1] = reflect.ValueOf(sess)
	sess := examples.MapSession{}
	args = f.Args()
	for _, arg := range pruned {
		if arg.T == T {
			args.Values[arg.N] = reflect.ValueOf(sess)
		}
	}
	f.Call(args)
	fmt.Println(sess.Get("message").(string))

	// Output: func (examples.Request, examples.Session)
	// examples.Request { }
	// nil store
	// Panic!
	// examples.Request { }
	// Hello, World!
}

func BenchmarkStatFunc(b *testing.B) {
	fn := func(req examples.Request, res examples.Response) {}
	for k := 0; k < b.N; k++ {
		call.StatFunc(fn)
	}
}

func TestStatFunc_NonFuncPanics(t *testing.T) {
	chk := assert.New(t)
	panicked := false
	defer func() {
		chk.True(panicked)
	}()
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	call.StatFunc(chk)
}
