package call

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nofeaturesonlybugs/call/examples"
)

func TestStat_SwapReceiver(t *testing.T) {
	chk := assert.New(t)
	//
	var talk *examples.Talker
	var instance *Instance
	for k := 0; k < 100; k++ {
		talk = new(examples.Talker)
		instance = TypeCache.Stat(talk)
		chk.Equal(talk, instance.receiver)
		chk.Equal(talk, instance.receiverValue.Interface())
	}
}
