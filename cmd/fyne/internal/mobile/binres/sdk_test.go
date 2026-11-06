package binres

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatestAPIResourcesPath(t *testing.T) {
	home := t.TempDir()
	for _, v := range []string{"android-33", "android-9", "android-36.1", "android-36"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(home, "platforms", v), 0o755))
	}
	t.Setenv("ANDROID_HOME", home)

	got, err := LatestAPIResourcesPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "platforms", "android-36.1", "android.jar"), got)
}
