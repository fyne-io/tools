package tools

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/execabs"
)

func CommandInShell(cmd string, args ...string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin": // darwin apps don't run in the user shell environment
		args = quoteArgs(args...)
		return execabs.Command(getDarwinShell(), "-i", "-c", cmd+" "+strings.Join(args, " "))
	case "linux", "freebsd", "netbsd", "openbsd", "dragonflybsd": // unix environment may be set up in shell
		args = quoteArgs(args...)
		return execabs.Command(getUnixShell(), "-i", "-c", cmd+" "+strings.Join(args, " "))
	default: // windows etc, just execute in global env
		return execabs.Command(cmd, args...)
	}
}

func quoteArgs(args ...string) []string {
	ret := make([]string, len(args))
	for i, s := range args {
		if strings.Contains(s, " ") {
			ret[i] = quoteString(s)
		} else {
			ret[i] = args[i]
		}
	}
	return ret
}

func quoteString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return "\"" + s + "\""
}

func getDarwinShell() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "zsh"
	}
	out, err := execabs.Command("dscl", ".", "-read", home, "UserShell").Output()
	if err != nil {
		return "zsh"
	}

	items := strings.Split(string(out), ":")
	if len(items) < 2 {
		return "zsh"
	}

	return strings.TrimSpace(items[1])
}

func getUnixShell() string {
	shell, ok := os.LookupEnv("SHELL")
	if !ok || shell == "" {
		return "/bin/sh"
	}

	return shell
}
