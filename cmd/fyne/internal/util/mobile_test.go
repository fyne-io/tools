package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAndroidBuildToolsPath_Newest(t *testing.T) {
	home := t.TempDir()
	for _, v := range []string{"30.0.3", "36.0.0", "33.0.2", "9.0.0"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(home, "build-tools", v), 0o755))
	}
	t.Setenv("ANDROID_HOME", home)

	assert.Equal(t, filepath.Join(home, "build-tools", "36.0.0"), AndroidBuildToolsPath())
}

func TestAndroidBuildToolsPath_NoVersionDir(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "build-tools")
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "zipalign"), nil, 0o755))
	t.Setenv("ANDROID_HOME", home)

	assert.Equal(t, dir, AndroidBuildToolsPath())
}
