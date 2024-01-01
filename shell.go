package tools

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandInShell sets up a new command with the environment set up by users environment.
// In darwin or other Unix systems the shell will be loaded by running the terminal and accessing env command.
// The command is run directly so stdout and stderr can be read directly from the returnd `exec.Cmd`.
func CommandInShell(cmd string, args ...string) *exec.Cmd {
	var env []string
	switch runtime.GOOS {
	case "darwin": // darwin apps don't run in the user shell environment
		args = quoteArgs(args...)
		data, err := exec.Command(getDarwinShell(), "-i", "-c", "env").Output()
		if err == nil {
			env = strings.Split(string(data), "\n")
		}
	case "linux", "freebsd", "netbsd", "openbsd", "dragonflybsd": // unix environment may be set up in shell
		args = quoteArgs(args...)
		data, err := exec.Command(getUnixShell(), "-i", "-c", "env").Output()
		if err == nil {
			env = strings.Split(string(data), "\n")
		}
	}

	run := exec.Command(cmd, args...)
	run.Env = append(run.Env, env...)
	return run
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
	out, err := exec.Command("dscl", ".", "-read", home, "UserShell").Output()
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
