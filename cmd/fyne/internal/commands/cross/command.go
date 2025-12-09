package cross

import (
	"strings"
)

// Command represents a command that can be executed in a container
type Command interface {
	Command() []string
}

// Ensure GoBuildCmd implements the Command interface.
var _ Command = (*GoBuildCmd)(nil)

// GoBuildCmd represents a "go build" command with configurable options.
type GoBuildCmd struct {
	goPackage string
	ldFlags   []string
	output    string
	tags      []string
	trimpath  bool
	verbose   bool
}

// NewGoBuildCmd creates a new GoBuildCmd with default empty values.
func NewGoBuildCmd() *GoBuildCmd {
	return &GoBuildCmd{
		ldFlags: []string{},
		tags:    []string{},
	}
}

// Command builds and returns the complete "go build" command with all configured options.
func (g *GoBuildCmd) Command() []string {
	cmd := []string{"go", "build"}

	if g.verbose {
		cmd = append(cmd, "-v")
	}

	if g.trimpath {
		cmd = append(cmd, "-trimpath")
	}

	if g.output != "" {
		cmd = append(cmd, "-o", g.output)
	}

	if len(g.ldFlags) > 0 {
		cmd = append(cmd, "-ldflags="+strings.Join(g.ldFlags, " "))
	}

	if len(g.tags) > 0 {
		cmd = append(cmd, "-tags", strings.Join(g.tags, ","))
	}

	if g.goPackage != "" {
		cmd = append(cmd, g.goPackage)
	}

	return cmd
}

// String returns a formatted string representation of the command.
// Implements the Stringer interface for display and debugging purposes.
func (g *GoBuildCmd) String() string {
	return strings.Join(g.Command(), " ")
}

// WithLDFlag adds a linker flag to the command.
func (g *GoBuildCmd) WithLDFlag(ldFlag string) {
	g.ldFlags = append(g.ldFlags, ldFlag)
}

// WithOutput sets the output file path (-o flag).
func (g *GoBuildCmd) WithOutput(output string) {
	g.output = output
}

// WithPackage sets the Go package to build.
func (g *GoBuildCmd) WithPackage(pkg string) {
	g.goPackage = pkg
}

// WithTag adds a build tag to the command.
func (g *GoBuildCmd) WithTag(tag string) {
	g.tags = append(g.tags, tag)
}

// WithTrimpath sets whether to enable -trimpath for reproducible builds.
func (g *GoBuildCmd) WithTrimpath(trimpath bool) {
	g.trimpath = trimpath
}

// WithVerbose sets whether to enable verbose output (-v flag).
func (g *GoBuildCmd) WithVerbose(verbose bool) {
	g.verbose = verbose
}
