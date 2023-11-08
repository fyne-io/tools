package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_QuoteArgs(t *testing.T) {
	assert.Equal(t, []string{"a", "\"b c\""}, quoteArgs([]string{"a", "b c"}...))
}

func Test_QuoteString(t *testing.T) {
	assert.Equal(t, "\"cat\"", quoteString("cat"))
	assert.Equal(t, "\"my-file.txt\"", quoteString("my-file.txt"))
	assert.Equal(t, "\"\\\"quoted\\\".svg\"", quoteString("\"quoted\".svg"))
}
