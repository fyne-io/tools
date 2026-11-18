package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type extractTest struct {
	value       string
	wantLdFlags string
	wantGoFlags string
}

func Test_ExtractLdFlags(t *testing.T) {
	goFlagsTests := []extractTest{
		{"-ldflags=-w", "-w", ""},
		{"-ldflags=-s", "-s", ""},
		{"-ldflags=-w -ldflags=-s", "-w -s", ""},
		{"-mod=vendor", "", "-mod=vendor"},
	}

	for _, test := range goFlagsTests {
		ldFlags, goFlags := ExtractLdFlags(test.value)
		assert.Equal(t, test.wantLdFlags, ldFlags)
		assert.Equal(t, test.wantGoFlags, goFlags)
	}
}

func Test_ExtractLdflagsFromGoFlags(t *testing.T) {
	ldFlags, env := ExtractLdflagsFromGoFlags([]string{"FOO=bar", "GOFLAGS=-ldflags=-X=main.version=1 -mod=mod"})
	assert.Equal(t, "-X=main.version=1", ldFlags)
	assert.Equal(t, []string{"FOO=bar", "GOFLAGS=-mod=mod"}, env)

	ldFlags, env = ExtractLdflagsFromGoFlags([]string{"FOO=bar"})
	assert.Equal(t, "", ldFlags)
	assert.Equal(t, []string{"FOO=bar"}, env)
}
