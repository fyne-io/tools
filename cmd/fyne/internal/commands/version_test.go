package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFyneGoModVersion(t *testing.T) {
	tests := []struct {
		dir string
		err bool
	}{
		{"/no/ne/ex/is/te/nt/dir", true},
		{".", false},
	}

	for _, c := range tests {
		v, err := getFyneGoModVersion(c.dir)
		if c.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotEqual(t, "", v)
		}
	}
}
