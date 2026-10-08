package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LookupDirWithGoModRelativeDir(t *testing.T) {
	wd, err := LookupDirWithGoMod(".")
	assert.NoError(t, err)
	assert.NotEqual(t, "", wd)
}

func Test_LookupDirWithGoModInvalidDir(t *testing.T) {
	wd, err := LookupDirWithGoMod("./any/thing/to/scan")
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)

	wd, err = LookupDirWithGoMod("../../any/thing/to/scan")
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)
}

func Test_LookupDirWithGoModInToolsRepo(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	wd, err := LookupDirWithGoMod(cwd)
	assert.NoError(t, err)
	assert.NotEqual(t, cwd, wd)
}

func Test_LookupDirWithGoModHitRoot(t *testing.T) {
	wd, err := LookupDirWithGoMod("/shouldnotexist123")
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", wd)
}
