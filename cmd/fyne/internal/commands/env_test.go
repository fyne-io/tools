package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lookupDirWithGoModRelativeDir(t *testing.T) {
	wd, err := lookupDirWithGoMod(".")
	assert.Equal(t, "", wd)
	assert.Equal(t, os.ErrNotExist, err)
}

func Test_lookupDirWithGoModInvalidDir(t *testing.T) {
	wd, err := lookupDirWithGoMod("./any/thing/to/scan")
	assert.Equal(t, "", wd)
	assert.Equal(t, os.ErrNotExist, err)
}

func Test_lookupDirWithGoModInToolsRepo(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	wd, err := lookupDirWithGoMod(cwd)
	assert.NoError(t, err)
	assert.NotEqual(t, cwd, wd)
}

func Test_lookupDirWithGoModHitRoot(t *testing.T) {
	wd, err := lookupDirWithGoMod("/shouldnotexist123")
	assert.Equal(t, os.ErrNotExist, err)
	assert.Equal(t, "", wd)
}
