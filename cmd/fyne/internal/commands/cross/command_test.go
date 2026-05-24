package cross

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGoBuildCmd(t *testing.T) {
	cmd := NewGoBuildCmd()
	require.NotNil(t, cmd)
	assert.False(t, cmd.verbose)
	assert.Empty(t, cmd.output)
	assert.Empty(t, cmd.ldFlags)
	assert.Empty(t, cmd.tags)
	assert.Empty(t, cmd.goPackage)
	assert.False(t, cmd.trimpath)
}

func TestGoBuildCmdWithVerbose(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithVerbose(true)
	assert.True(t, cmd.verbose)
}

func TestGoBuildCmdWithOutput(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithOutput("myapp")
	assert.Equal(t, "myapp", cmd.output)
}

func TestGoBuildCmdWithLDFlag(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithLDFlag("-s")
	cmd.WithLDFlag("-w")
	assert.Equal(t, []string{"-s", "-w"}, cmd.ldFlags)
}

func TestGoBuildCmdWithTag(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithTag("release")
	cmd.WithTag("custom")
	assert.Equal(t, []string{"release", "custom"}, cmd.tags)
}

func TestGoBuildCmdWithPackage(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithPackage("./cmd/app")
	assert.Equal(t, "./cmd/app", cmd.goPackage)
}

func TestGoBuildCmdWithTrimpath(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithTrimpath(true)
	assert.True(t, cmd.trimpath)
}

func TestGoBuildCmdCommand_Basic(t *testing.T) {
	cmd := NewGoBuildCmd()
	result := cmd.Command()
	assert.Equal(t, []string{"go", "build"}, result)
}

func TestGoBuildCmdCommand_WithVerbose(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithVerbose(true)
	result := cmd.Command()
	assert.Contains(t, result, "-v")
}

func TestGoBuildCmdCommand_WithTrimpath(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithTrimpath(true)
	result := cmd.Command()
	assert.Contains(t, result, "-trimpath")
}

func TestGoBuildCmdCommand_WithOutput(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithOutput("myapp")
	result := cmd.Command()
	assert.Contains(t, result, "-o")
	assert.Contains(t, result, "myapp")
}

func TestGoBuildCmdCommand_WithLDFlags(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithLDFlag("-s")
	cmd.WithLDFlag("-w")
	result := cmd.Command()

	// Should be single -ldflags with space-separated values
	assert.Contains(t, result, "-ldflags=-s -w", "expected -ldflags=-s -w in command")
}

func TestGoBuildCmdCommand_WithTags(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithTag("release")
	cmd.WithTag("custom")
	result := cmd.Command()

	// Should be -tags with comma-separated values
	assert.Contains(t, result, "-tags")
	assert.Contains(t, result, "release,custom")
}

func TestGoBuildCmdCommand_WithPackage(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithPackage("./cmd/app")
	result := cmd.Command()
	assert.Contains(t, result, "./cmd/app")
}

func TestGoBuildCmdCommand_ComplexScenario(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithVerbose(true)
	cmd.WithTrimpath(true)
	cmd.WithOutput("myapp")
	cmd.WithLDFlag("-s")
	cmd.WithLDFlag("-w")
	cmd.WithTag("release")
	cmd.WithPackage("./cmd/app")

	result := cmd.Command()

	// Verify all components are present
	assert.Contains(t, result, "go")
	assert.Contains(t, result, "build")
	assert.Contains(t, result, "-v")
	assert.Contains(t, result, "-trimpath")
	assert.Contains(t, result, "-o")
	assert.Contains(t, result, "myapp")
	assert.Contains(t, result, "-ldflags=-s -w")
	assert.Contains(t, result, "-tags")
	assert.Contains(t, result, "release")
	assert.Contains(t, result, "./cmd/app")
}

func TestCommandInterface(t *testing.T) {
	// Verify GoBuildCmd implements Command interface
	cmd := NewGoBuildCmd()
	var _ Command = cmd
}

func TestGoBuildCmdString_Basic(t *testing.T) {
	cmd := NewGoBuildCmd()
	result := cmd.String()
	assert.Equal(t, "go build", result)
}

func TestGoBuildCmdString_WithVerbose(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithVerbose(true)
	result := cmd.String()
	assert.Contains(t, result, "go build")
	assert.Contains(t, result, "-v")
}

func TestGoBuildCmdString_WithOutput(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithOutput("myapp")
	result := cmd.String()
	assert.Contains(t, result, "-o myapp")
}

func TestGoBuildCmdString_WithLDFlags(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithLDFlag("-s")
	cmd.WithLDFlag("-w")
	result := cmd.String()
	assert.Contains(t, result, "-ldflags=-s -w")
}

func TestGoBuildCmdString_WithTags(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithTag("release")
	cmd.WithTag("custom")
	result := cmd.String()
	assert.Contains(t, result, "-tags release,custom")
}

func TestGoBuildCmdString_WithPackage(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithPackage("./cmd/app")
	result := cmd.String()
	assert.Contains(t, result, "./cmd/app")
	// Package should be at the end
	assert.True(t, strings.HasSuffix(result, "./cmd/app"), "package should be at the end")
}

func TestGoBuildCmdString_ComplexScenario(t *testing.T) {
	cmd := NewGoBuildCmd()
	cmd.WithVerbose(true)
	cmd.WithTrimpath(true)
	cmd.WithOutput("output/myapp")
	cmd.WithLDFlag("-s")
	cmd.WithLDFlag("-w")
	cmd.WithTag("release")
	cmd.WithTag("custom")
	cmd.WithPackage("./cmd/app")

	result := cmd.String()

	// Verify structure: go build [flags] package
	assert.Contains(t, result, "go build")
	assert.Contains(t, result, "-v")
	assert.Contains(t, result, "-trimpath")
	assert.Contains(t, result, "-o output/myapp")
	assert.Contains(t, result, "-ldflags=-s -w")
	assert.Contains(t, result, "-tags release,custom")
	assert.Contains(t, result, "./cmd/app")

	// Verify it's a valid space-separated command
	assert.NotEmpty(t, result)
	// Should be executable command format: go build [args...]
	parts := strings.Fields(result)
	assert.GreaterOrEqual(t, len(parts), 2)
	assert.Equal(t, "go", parts[0])
	assert.Equal(t, "build", parts[1])
}
