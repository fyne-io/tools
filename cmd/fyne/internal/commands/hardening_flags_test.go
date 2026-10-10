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

	// arm64 doesn't support -fcf-protection
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningCFlagsLookup("cc (Ubuntu) 11.4.0", "ubuntu", "arm64"))

	// no specific flags
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("clang version 1.2.3", "darwin", "amd64"))
	assert.Equal(t, hardeningCFLAGS, hardeningCFlagsLookup("cc (Whatever) 1.2.3", "linux", "i386"))

	// windows/mingw lacks -fcf-protection support
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningCFlagsLookup("cc (GCC) 2.3.4", "windows", "i386"))

	// darwin/arm64 doesn't do -fcf-protection
	assert.Equal(t, "-D_FORTIFY_SOURCE=3 -fstack-protector-strong", hardeningCFlagsLookup("Apple clang version 17.0.0 (clang-1700.0.13.5)", "darwin", "arm64"))
}

func Test_gccDefaultFlags(t *testing.T) {
	in := `
GGC heuristics: --param ggc-min-expand=100 --param ggc-min-heapsize=131072
options passed:  -v -auxbase
options enabled:  -falign-loops -fargument-alias
 -fasynchronous-unwind-tables -fbranch-count-reg -fcommon -fearly-inlining
 -feliminate-unused-debug-types -femit-class-debug-always -ffunction-cse
 -fgcse-lm -finline-functions-called-once -fivopts -fkeep-static-consts
 -fleading-underscore -fmath-errno -fmove-loop-invariants -fpeephole -fpic
 -fpie -freg-struct-return -fsched-interblock -fsched-spec
 -fsched-stalled-insns-dep -fshow-column -fsplit-ivs-in-unroller
 -ftoplevel-reorder -ftrapping-math -ftree-loop-im -ftree-loop-ivcanon
 -ftree-loop-optimize -ftree-vect-loop-version -funwind-tables
 -fvar-tracking -m128bit-long-double -m64 -m80387
 -maccumulate-outgoing-args -malign-stringops -mfancy-math-387
 -mfp-ret-in-387 -mieee-fp -mmmx -mpush-args -mred-zone -msse -msse2
Compiler executable checksum: 1e242b9c9b970b8bb2f3765bba204637
`
	x := gccDefaultFlags(in)
	assert.NotNil(t, x)
}

func Test_ccVersionAndDefaultFlags(t *testing.T) {
	s, err := ccVersionAndDefaultFlags()
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}
