package mobile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EnvInitPassesGoFlagsLdflags(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Skip("no NDK toolchain for this host:", r)
			}
		}()
		archNDK()
	}()

	oldN, oldLdflags, oldTarget, oldAndroidEnv := buildN, buildLdflags, buildTarget, androidEnv
	t.Cleanup(func() {
		buildN, buildLdflags, buildTarget, androidEnv = oldN, oldLdflags, oldTarget, oldAndroidEnv
	})
	t.Setenv("GOFLAGS", "-ldflags=-checklinkname=0 -mod=mod")
	buildN = true // skips the NDK lookup
	buildTarget = "android"
	buildLdflags = "-w" // as set for a release build

	assert.NoError(t, envInit())
	assert.Equal(t, "-w -checklinkname=0", buildLdflags)
	for arch, env := range androidEnv {
		assert.Contains(t, env, "GOFLAGS=-mod=mod", arch)
	}
}
