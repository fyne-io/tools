package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageWindowsFixedVersionInfo(t *testing.T) {
	v1 := fixedVersionInfo("1.2.3")
	assert.Equal(t, 1, v1.Major)
	assert.Equal(t, 2, v1.Minor)
	assert.Equal(t, 3, v1.Patch)

	v2 := fixedVersionInfo("1.2.4-dev")
	assert.Equal(t, 1, v2.Major)
	assert.Equal(t, 2, v2.Minor)
	assert.Equal(t, 0, v2.Patch)
}
