package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppID(t *testing.T) {
	for _, test := range []struct {
		modpath string
		appid   string
	}{
		{"github.com/example/fyne-example", "com.github.example.fyne-example"},
		{"example/fyne-example", "example.fyne-example"},
		{"fyne-example", "fyne-example"},
		{"fyne.example", "example.fyne"},
		{"fyne.example/app", "example.fyne.app"},
	} {
		assert.Equal(t, test.appid, getAppID(test.modpath))
	}
}

func TestAppName(t *testing.T) {
	for _, test := range []struct {
		modpath string
		appname string
	}{
		{"github.com/example/fyne-example", "fyne-example"},
		{"example/fyne-example", "fyne-example"},
		{"fyne-example", "fyne-example"},
		{"example.fyne", "example"},
		{"fyne.example/app", "app"},
	} {
		assert.Equal(t, test.appname, getAppName(test.modpath))
	}
}

func TestCheckFileOrDoCreate(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "checkfile")

	done := 0
	assert.Nil(t, checkFileOrDo(file, func() error {
		done++
		return nil
	}))
	assert.Equal(t, 1, done)

	assert.Nil(t, checkFileOrCreate(file, "42"))
	b, err := os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, "42", string(b))

	assert.Nil(t, checkFileOrDo(file, func() error {
		done++
		return nil
	}))
	assert.Equal(t, 1, done)

	assert.Nil(t, checkFileOrCreate(file, "43"))
	c, err := os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, "42", string(c))
}
