package cross

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineDebug_Basic(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	debug := engine.Debug()

	// Should contain target, image, and environment
	assert.Contains(t, debug, "Target:")
	assert.Contains(t, debug, "linux/amd64")
	assert.Contains(t, debug, "Image:")
	assert.Contains(t, debug, "fyneio/fyne-cross-images:linux")
	assert.Contains(t, debug, "Environment:")
}

func TestEngineDebug_WithMounts(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	tmpDir := t.TempDir()
	engine.WithMount(tmpDir, "/mnt/src")
	engine.WithMount(filepath.Join(tmpDir, "cache"), "/mnt/cache")

	debug := engine.Debug()

	assert.Contains(t, debug, "Mounts:")
	assert.Contains(t, debug, tmpDir)
	assert.Contains(t, debug, "/mnt/src")
	assert.Contains(t, debug, "/mnt/cache")
}

func TestEngineDebug_WithCustomEnv(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	engine.WithEnv(map[string]string{
		"CUSTOM_VAR": "custom_value",
	})

	debug := engine.Debug()

	assert.Contains(t, debug, "CUSTOM_VAR=custom_value")
}

func TestEngineDebug_EnvVarsSorted(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	// Add some env vars
	engine.WithEnv(map[string]string{
		"ZEBRA": "z_value",
		"ALPHA": "a_value",
		"BRAVO": "b_value",
	})

	debug := engine.Debug()
	lines := strings.Split(debug, "\n")

	// Find the indices of our env vars in the output
	var alphaIdx, bravoIdx, zebraIdx int
	for i, line := range lines {
		if strings.Contains(line, "ALPHA=") {
			alphaIdx = i
		}
		if strings.Contains(line, "BRAVO=") {
			bravoIdx = i
		}
		if strings.Contains(line, "ZEBRA=") {
			zebraIdx = i
		}
	}

	// Verify they appear in sorted order
	if alphaIdx > 0 && bravoIdx > 0 && zebraIdx > 0 {
		assert.Less(t, alphaIdx, bravoIdx, "ALPHA should come before BRAVO")
		assert.Less(t, bravoIdx, zebraIdx, "BRAVO should come before ZEBRA")
	}
}

func TestEngineDebug_WithCache(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	cacheDir := t.TempDir()
	engine.WithCache(cacheDir)

	debug := engine.Debug()

	// Should show cache-related settings
	assert.Contains(t, debug, "GOCACHE=/go/go-build")
	assert.Contains(t, debug, "ZIG_GLOBAL_CACHE_DIR=/go/zig-cache")
	assert.Contains(t, debug, "Mounts:")
	assert.Contains(t, debug, cacheDir)
}

func TestEngineDebug_InvalidTarget(t *testing.T) {
	engine, err := NewEngine("darwin", "arm64")
	assert.Error(t, err)
	assert.Nil(t, engine)
}

func TestEngineDebug_MultipleConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		goos   string
		goarch string
	}{
		{"linux amd64", "linux", "amd64"},
		{"linux arm64", "linux", "arm64"},
		{"windows amd64", "windows", "amd64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewEngine(tt.goos, tt.goarch)
			require.NoError(t, err)
			defer engine.Close()

			debug := engine.Debug()

			// Verify basic structure for each target
			assert.Contains(t, debug, "Target:")
			assert.Contains(t, debug, tt.goos+"/"+tt.goarch)
			assert.Contains(t, debug, "Image:")
			assert.NotEmpty(t, debug)
		})
	}
}

func TestEngineDebug_NoMountsWhenEmpty(t *testing.T) {
	engine, err := NewEngine("linux", "amd64")
	require.NoError(t, err)
	defer engine.Close()

	// Don't add any mounts
	debug := engine.Debug()

	// Should not have "Mounts:" if no mounts are configured
	assert.NotContains(t, debug, "Mounts:")
}
