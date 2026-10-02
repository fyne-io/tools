package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ShellQuote(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"", "''"},
		{"myapp", "myapp"},
		{"myapp.tar.xz", "myapp.tar.xz"},
		{"/usr/local/bin/my-app_2.0", "/usr/local/bin/my-app_2.0"},
		{"Fyne Demo.tar.xz", "'Fyne Demo.tar.xz'"},
		{"two words", "'two words'"},
		{"say \"hi\"", `'say "hi"'`},
		{"it's", `'it'\''s'`},
		{"$HOME", "'$HOME'"},
		{"a*b?c", "'a*b?c'"},
	}

	for _, test := range tests {
		assert.Equal(t, test.out, ShellQuote(test.in), "quoting %q", test.in)
	}
}

func Test_ShellQuoteCommand(t *testing.T) {
	assert.Equal(t, "go build", ShellQuoteCommand("go", []string{"build"}))
	assert.Equal(t, "go build -o ./myapp", ShellQuoteCommand("go", []string{"build", "-o", "./myapp"}))
	assert.Equal(t, "go build -o './Fyne Demo'", ShellQuoteCommand("go", []string{"build", "-o", "./Fyne Demo"}))
	assert.Equal(t, "'/opt/go dir/bin/go' build", ShellQuoteCommand("/opt/go dir/bin/go", []string{"build"}))
}
