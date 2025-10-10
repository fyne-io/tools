package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"foo", "bar"}, "foo"))
	assert.False(t, Contains([]string{"foo", "bar"}, "baz"))
	assert.False(t, Contains([]string{"foo"}, ""))
	assert.False(t, Contains([]string{}, "foo"))
}
