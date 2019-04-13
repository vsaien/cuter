package threading

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/vsaien/cuter/lib/lang"

	"github.com/stretchr/testify/assert"
)

func TestRunSafe(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	i := 0

	defer func() {
		assert.Equal(t, 1, i)
	}()

	ch := make(chan lang.PlaceholderType)
	go RunSafe(func() {
		defer func() {
			ch <- lang.Placeholder
		}()

		panic("panic")
	})

	<-ch
	i++
}
