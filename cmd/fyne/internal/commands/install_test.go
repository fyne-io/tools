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

func TestGetInstallBaseDir(t *testing.T) {
	for _, test := range []struct {
		path string
		pkg  string
		root string
		want string
	}{
		{"dir1", "example.com/foo", "example.com/foo", "dir1"},
		{"dir2", "example.com/foo/cmd/bar", "example.com/foo", "dir2/cmd/bar"},
	} {
		assert.Equal(t, test.want, getInstallBaseDir(test.path, test.pkg, test.root))
	}
}
