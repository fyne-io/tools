package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want int
	}{
		{"33.0.2", "36.0.0", -1},
		{"36.0.0", "33.0.2", 1},
		{"9.0.0", "36.0.0", -1},
		{"36.0.0", "36.0.0", 0},
		{"36", "36.0.0", 0},
		{"36.1", "36.0.0", 1},
		{"36.0.0-rc1", "36.0.0", 0},
		{"36.0.0-rc1", "35.0.1", 1},
		{"android-9", "android-36", -1},
		{"android-36.1", "android-36", 1},
		{"android-36", "android-Baklava", 1},
		{"android-Baklava", "", 0},
		{"", "30.0.3", -1},
	} {
		assert.Equal(t, tt.want, CompareVersions(tt.a, tt.b), "%q vs %q", tt.a, tt.b)
	}
}
