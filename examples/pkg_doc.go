// Package examples exports a few types for testing and examples.
package examples

import (
	"fmt"
	"net/http"
)

// Request is used to demonstrate how the call package handles struct or ptr-to-struct
// types when calling MethodInfo.Args()
type Request struct {
	Origin string
	Token  string
}

// Response is used to demonstrate how the call package handles interface types
// when calling MethodInfo.Args().
type Response interface {
	A()
	B()
}

// Session represents a user session.
type Session interface {
	Get(string) interface{}
	Set(string, interface{})
}

// Talker has a few methods to demonstrate the parent package call.
type Talker struct{}

// Error always returns an error.
func (t Talker) Error(r Response, req *Request) error {
	t.hidden() // Calling t.hidden() here quiets the linter.
	return fmt.Errorf("%T made an error", t)
}

// Goodbye says goodbye via .
func (t Talker) Goodbye(req *Request, inlineStruct struct {
	StringField string
	NumField    int
}) {
}

// Hello says hello via .
func (t Talker) Hello(r Response, req *Request) (bool, error) {
	return false, nil
}

// hidden is an unexported method and only useful as part of tests.
func (t Talker) hidden() {
}

// Person implements some simple methods.
type Person struct {
	Name string
	Age  int
}

// Greet returns a string with a greeting from the Person.
func (p Person) Greet() string {
	return fmt.Sprintf("Hello!  My name is %v and I am %v year(s) old.", p.Name, p.Age)
}

// HTTP has methods using types from `net/http` and exists for performance benchmarking.
type HTTP struct{}

// Handler has the signature required for http.Handler.
func (h HTTP) Handler(w http.ResponseWriter, req *http.Request, sess Session, form struct {
	Username string `form:"username"`
	Password string `form:"password"`
}) {
	if w != nil {
		// For testing purposes some of our tests can pass this in as nil.
		w.WriteHeader(http.StatusOK)
	}
}

// ManyArgs has a method with many arguments.
type ManyArgs struct{}

func (m ManyArgs) Many(r Response, req *Request, sess Session, a, b, c *Request) {
}
