package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildRejectsMobileTargets(t *testing.T) {
	for _, target := range []string{"android", "android/arm64", "android/386", "ios", "iossimulator"} {
		t.Run(target, func(t *testing.T) {
			b := &Builder{appData: &appData{}, os: target, runner: &testCommandRuns{t: t}}

			err := b.build()
			assert.EqualError(t, err, "the build command cannot target "+target+
				", mobile packages are created with \"fyne package --target "+target+"\"")
		})
	}
}

func Test_BuildWasmVersion(t *testing.T) {
	expected := []mockRunner{
		{
			expectedValue: expectedValue{
				args:  []string{"build"},
				env:   []string{"GOARCH=wasm", "GOOS=js", "CGO_ENABLED=0"},
				osEnv: true,
				dir:   "myTest",
			},
			mockReturn: mockReturn{ret: []byte("")},
		},
	}

	wasmBuildTest := &testCommandRuns{runs: expected, t: t}
	b := &Builder{appData: &appData{}, os: "wasm", srcdir: "myTest", runner: wasmBuildTest}
	err := b.build()
	assert.NoError(t, err)
	wasmBuildTest.verifyExpectation()
}

func Test_BuildWasmReleaseVersion(t *testing.T) {
	expected := []mockRunner{
		{
			expectedValue: expectedValue{
				args:  []string{"build", "-trimpath", "-ldflags", "-s -w", "-tags", "release"},
				env:   []string{"GOARCH=wasm", "GOOS=js", "CGO_ENABLED=0"},
				osEnv: true,
				dir:   "myTest",
			},
			mockReturn: mockReturn{
				ret: []byte(""),
			},
		},
	}

	wasmBuildTest := &testCommandRuns{runs: expected, t: t}
	b := &Builder{appData: &appData{}, os: "wasm", srcdir: "myTest", release: true, runner: wasmBuildTest}
	err := b.build()
	assert.Nil(t, err)
	wasmBuildTest.verifyExpectation()
}

func Test_BuildKeepsGoFlagsLdflags(t *testing.T) {
	goFlags := "-ldflags=-X=main.version=v9.9.9 -mod=mod"
	t.Setenv("GOFLAGS", goFlags)

	var env []string
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOFLAGS=") {
			env = append(env, v)
		}
	}
	env = append(env, "GOFLAGS=-mod=mod", "GOARCH=wasm", "GOOS=js", "CGO_ENABLED=0")

	run := mockRunner{
		expectedValue: expectedValue{
			args: []string{"build", "-ldflags", "-X=main.version=v9.9.9"},
			env:  env,
			dir:  "myTest",
		},
		mockReturn: mockReturn{ret: []byte("")},
	}

	// The Windows packager builds twice in one process, the second build needs the ldflags as well.
	wasmBuildTest := &testCommandRuns{runs: []mockRunner{run, run}, t: t}
	b := &Builder{appData: &appData{}, os: "wasm", srcdir: "myTest", runner: wasmBuildTest}
	assert.NoError(t, b.build())
	assert.Equal(t, goFlags, os.Getenv("GOFLAGS"))
	assert.NoError(t, b.build())
	assert.Equal(t, goFlags, os.Getenv("GOFLAGS"))
	wasmBuildTest.verifyExpectation()
}

func Test_BuildLinuxReleaseVersion(t *testing.T) {
	relativePath := "." + string(os.PathSeparator) + filepath.Join("cmd", "terminal")

	cflags, exists := os.LookupEnv("CGO_CFLAGS")
	if exists {
		cflags += " "
	}
	ldflags, exists := os.LookupEnv("CGO_LDFLAGS")
	if exists {
		ldflags += " "
	}

	archcflags := ""
	arch := targetArch()
	switch arch {
	case "arm64":
		archcflags = " -mbranch-protection=bti+pac-ret"
	}

	hardeningCFlags := hardeningCFlagsLookup(ccVersion(), "linux", arch)
	expected := []mockRunner{
		{
			expectedValue: expectedValue{
				args:  []string{"build", "-trimpath", "-ldflags", "-s -w", "-tags", "release", relativePath},
				env:   []string{"CGO_ENABLED=1", "GOOS=linux", fmt.Sprintf("CGO_CFLAGS=%s%s %s%s", cflags, baseCFLAGSRelease, hardeningCFlags, archcflags), fmt.Sprintf("CGO_LDFLAGS=%s%s", ldflags, hardeningLDFLAGSLinux)},
				osEnv: true,
				dir:   "myTest",
			},
			mockReturn: mockReturn{
				ret: []byte(""),
			},
		},
	}

	linuxBuildTest := &testCommandRuns{runs: expected, t: t}
	b := &Builder{appData: &appData{}, os: "linux", srcdir: "myTest", release: true, runner: linuxBuildTest, goPackage: relativePath}
	err := b.build()
	assert.NoError(t, err)
	linuxBuildTest.verifyExpectation()
}

func Test_AppendEnv(t *testing.T) {
	env := []string{
		"foo=bar",
		"bar=baz",
		"foo1=bar=baz",
	}

	appendEnv(&env, "foo2", "baz2")
	appendEnv(&env, "foo", "baz")
	appendEnv(&env, "foo1", "-bar")

	if assert.Len(t, env, 4) {
		assert.Equal(t, "foo=bar baz", env[0])
		assert.Equal(t, "bar=baz", env[1])
		assert.Equal(t, "foo1=bar=baz -bar", env[2])
		assert.Equal(t, "foo2=baz2", env[3])
	}
}

func Test_NormaliseVersion(t *testing.T) {
	assert.Equal(t, "master", normaliseVersion("master"))
	assert.Equal(t, "2.3.0.0", normaliseVersion("v2.3"))
	assert.Equal(t, "2.4.0.0", normaliseVersion("v2.4.0"))
	assert.Equal(t, "2.3.6.0-dev", normaliseVersion("v2.3.6-0.20230711180435-d4b95e1cb1eb"))
	assert.Equal(t, "2.4.1.0-dev", normaliseVersion("v2.4.1-rc7.0.20230711180435-d4b95e1cb1eb"))
}
