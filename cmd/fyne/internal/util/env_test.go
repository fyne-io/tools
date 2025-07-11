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
		ldFlags, goFlags := extractLdFlags(test.value)
		assert.Equal(t, test.wantLdFlags, ldFlags)
		assert.Equal(t, test.wantGoFlags, goFlags)
	}
}
