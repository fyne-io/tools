package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_keyValueFlag_String(t *testing.T) {
	kvflag := &keyValueFlag{}
	assert.NoError(t, kvflag.Set("foo=bar"))
	assert.NoError(t, kvflag.Set("a=b=c"))
	assert.Error(t, kvflag.Set("meep"))

	assert.Equal(t, `"a=b=c", "foo=bar"`, kvflag.String())
}
