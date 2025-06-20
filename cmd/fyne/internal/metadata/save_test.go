package metadata

import (
	"bytes"
	"os"
	"strings"
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
	r, err := os.Open("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	data, err := Load(r)
	assert.Nil(t, err)
	r.Close()

	w := &bytes.Buffer{}
	err = Save(data, w)
	assert.Nil(t, err)

	content, err := os.ReadFile("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	expected := strings.ReplaceAll(string(content), "\r\n", "\n")
	actual := strings.ReplaceAll(w.String(), "\r\n", "\n")
	assert.Equal(t, expected, actual)
}
