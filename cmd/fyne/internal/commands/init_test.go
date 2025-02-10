package commands

import (
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
		{"sub.example.com", "com.example.sub"},
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
		{"sub.domain.fyne.example", "sub"},
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
}
