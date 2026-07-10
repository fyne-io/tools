package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hardeningFlagsLookup(t *testing.T) {
	assert.Equal(t, "", hardeningFlagsLookup("", "ubuntu", "amd64"))
	assert.Equal(t, "-fstack-protector-strong", hardeningFlagsLookup("clang version 11.3.0", "ubuntu", "amd64"))
	assert.Equal(t, "-fstack-protector-strong", hardeningFlagsLookup("clang version 11.4.0", "ubuntu", "amd64"))
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningFlagsLookup("clang version 11.4.1", "ubuntu", "amd64"))
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningFlagsLookup("clang version 11.4.0", "ubuntu", "arm64"))
	assert.Equal(t, hardeningCFLAGS, hardeningFlagsLookup("clang version 1.2.3", "darwin", "arm"))
	assert.Equal(t, hardeningCFLAGS, hardeningFlagsLookup("cc (Whatever) 1.2.3", "linux", "i386"))
}
