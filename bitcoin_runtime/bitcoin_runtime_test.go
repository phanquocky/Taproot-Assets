package bitcoin_runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitcoinRuntime(t *testing.T) {
	br := New()

	t.Run("Should start bitcoin runtime", func(t *testing.T) {
		err := br.SetUpRuntime()

		assert.Nil(t, err)
	})

	// t.Run("Should stop bitcoin runtime", func(t *testing.T) {
	// 	br.stopBtcd()
	// 	time.Sleep(3 * time.Second)
	// 	br.stopBtcwallet()
	// })
}
