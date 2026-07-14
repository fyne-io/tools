package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hardeningCFlagsLookup(t *testing.T) {
	// no compiler, no flags
	assert.Equal(t, "", hardeningCFlagsLookup("", "ubuntu", "amd64"))

	// compiler version lower or equal
	assert.Equal(t, "-fcf-protection -fstack-protector-strong", hardeningCFlagsLookup("cc (Ubuntu) 11.3.0", "ubuntu", "amd64"))
	assert.Equal(t, "-fcf-protection -fstack-protector-strong", hardeningCFlagsLookup("cc (Ubuntu) 11.4.0", "ubuntu", "amd64"))

	// compiler version higher
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("cc (Ubuntu) 11.4.1", "ubuntu", "amd64"))

	// different compiler
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("clang version 11.3.0", "ubuntu", "amd64"))
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("clang version 11.4.0", "ubuntu", "amd64"))

	// different arch
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("cc (Ubuntu) 11.4.0", "ubuntu", "arm64"))

	// no specific flags
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("clang version 1.2.3", "darwin", "arm"))
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("cc (Whatever) 1.2.3", "linux", "i386"))

	// windows/mingw lacks -fcf-protection support
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningCFlagsLookup("cc (GCC) 2.3.4", "windows", "i386"))
}
