package commands

import (
	"testing"

	 "github.com/stretchr/testify/assert"
)

func TestAppID(t *testing.T) {
	for _, test := range []struct {
		modpath string
		appid   string
	}{
		{ "github.com/example/fyne-example", "com.github.example.fyne-example" },
		{ "example/fyne-example", "example.fyne-example" },
		{ "fyne-example", "fyne-example" },
		{ "fyne.example", "example.fyne" },
		{ "fyne.example/app", "example.fyne.app" },
	} {
		assert.Equal(t, test.appid, getAppID(test.modpath))
	}
}

func TestAppName(t *testing.T) {
	for _, test := range []struct {
		modpath string
		appname string
	}{
		{ "github.com/example/fyne-example", "fyne-example" },
		{ "example/fyne-example", "fyne-example" },
		{ "fyne-example", "fyne-example" },
		{ "example.fyne", "example" },
		{ "fyne.example/app", "app" },
	} {
		assert.Equal(t, test.appname, getAppName(test.modpath))
	}
}
