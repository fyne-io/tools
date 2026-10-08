package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/tools/cmd/fyne/internal/util"
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
		{"dir3", "", "", "dir3"},
		{"dir4", "example.com", "", "dir4"},
		{"dir5", "", "example.com", "dir5"},
	} {
		assert.Equal(t, test.want, getInstallBaseDir(test.path, test.pkg, test.root))
	}
}

func Test_InstallerEnsurePackage(t *testing.T) {
	target := filepath.Join(t.TempDir(), "myapp.apk")
	buildErr := errors.New("build failed")

	// an existing package is reused without a build
	build := func() error {
		return buildErr
	}
	require.NoError(t, os.WriteFile(target, nil, util.FilePermDefault))
	assert.NoError(t, (&Installer{}).ensurePackage(target, build))

	// a missing package is built
	require.NoError(t, os.Remove(target))
	built := false
	assert.NoError(t, (&Installer{}).ensurePackage(target, func() error {
		built = true
		return os.WriteFile(target, nil, util.FilePermDefault)
	}))
	assert.True(t, built)

	// a failed build is reported, so that no outdated package is installed
	require.NoError(t, os.Remove(target))
	assert.ErrorIs(t, (&Installer{}).ensurePackage(target, build), buildErr)
}
