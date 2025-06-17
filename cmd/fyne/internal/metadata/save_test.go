package metadata

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveAppMetadata(t *testing.T) {
	r, err := os.Open("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	data, err := Load(r)
	_ = r.Close()
	assert.Nil(t, err)
	assert.Equal(t, 1, data.Details.Build)

	data.Details.Build++

	versionPath := "./testdata/Version.toml"
	w, err := os.Create(versionPath)
	assert.Nil(t, err)
	err = Save(data, w)
	assert.Nil(t, err)
	defer func() {
		os.Remove(versionPath)
	}()
	_ = w.Close()

	r, err = os.Open(versionPath)
	assert.Nil(t, err)
	defer r.Close()

	data2, err := Load(r)
	assert.Nil(t, err)
	assert.Equal(t, 2, data2.Details.Build)
}

func TestSaveIndentation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Don't run indentation tests on Windows, seems to fail for no apparent reason")
	}

	r, err := os.Open("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	data, err := Load(r)
	assert.Nil(t, err)
	r.Close()

	w, err := os.Create("./testdata/new.toml")
	assert.Nil(t, err)
	t.Cleanup(func() { os.Remove("./testdata/new.toml") })
	err = Save(data, w)
	assert.Nil(t, err)
	w.Close()

	expected, err := os.ReadFile("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	actual, err := os.ReadFile("./testdata/new.toml")
	assert.Nil(t, err)
	assert.Equal(t, string(expected), string(actual))
}
