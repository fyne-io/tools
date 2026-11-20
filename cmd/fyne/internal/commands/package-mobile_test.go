package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseHexColor(t *testing.T) {
	r, g, b, err := parseHexColor("#015952")
	assert.Nil(t, err)
	assert.InDelta(t, 1.0/255, r, 0.001)
	assert.InDelta(t, 0x59/255.0, g, 0.001)
	assert.InDelta(t, 0x52/255.0, b, 0.001)

	r, g, b, err = parseHexColor("")
	assert.Nil(t, err)
	assert.Equal(t, []float64{1, 1, 1}, []float64{r, g, b})

	_, _, _, err = parseHexColor("#0159")
	assert.NotNil(t, err)
	_, _, _, err = parseHexColor("#01595z")
	assert.NotNil(t, err)
}
