package call_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call"
	"github.com/nofeaturesonlybugs/call/examples"
)

func TestCache_Stat_Nil(t *testing.T) {
	chk := assert.New(t)
	//
	instance := call.Stat(nil)
	chk.Nil(instance)
}

func BenchmarkStat(b *testing.B) {
	var talk examples.Talker
	for k := 0; k < b.N; k++ {
		call.Stat(talk)
	}
}
