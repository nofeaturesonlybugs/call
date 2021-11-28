package call_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call"
)

func TestCache_StatNil(t *testing.T) {
	chk := assert.New(t)
	//
	methods := call.Stat(nil)
	chk.Nil(methods.Receiver)
}
