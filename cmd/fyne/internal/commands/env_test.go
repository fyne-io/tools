package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lookupDirWithGoModRelativeDir(t *testing.T) {
	wd, err := lookupDirWithGoMod(".")
	assert.NoError(t, err)
	assert.NotEqual(t, "", wd)
}

func Test_lookupDirWithGoModInvalidDir(t *testing.T) {
	wd, err := lookupDirWithGoMod("./any/thing/to/scan")
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)

	wd, err = lookupDirWithGoMod("../../any/thing/to/scan")
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)
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
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)
}
