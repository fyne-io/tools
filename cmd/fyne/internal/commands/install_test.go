package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPackageAndBranch(t *testing.T) {
	for _, test := range []struct {
		input  string
		pkg    string
		branch string
	}{
		{"foo", "foo", ""},
		{"foo@bar", "foo", "bar"},
	} {
		pkg, branch := getPackageAndBranch(test.input)
		assert.Equal(t, test.pkg, pkg)
		assert.Equal(t, test.branch, branch)
	}
}
