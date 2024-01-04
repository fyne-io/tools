package tools

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type staticString string

// CommandInShell sets up a new command with the environment set up by user's environment.
// In darwin or other Unix systems the shell will be loaded by running the terminal and accessing env command.
// The command is run directly so stdout and stderr can be read directly from the returned `exec.Cmd`.
// The cmd parameter must be a string constant to avoid possible code injection attacks.
func CommandInShell(cmd staticString, args ...string) *exec.Cmd {
	// lookup command path in target env
	path := string(cmd)
	if runtime.GOOS == "darwin" {
		data, err := runInShell("which", path).Output()

		if err == nil {
			// parse lines as shell can output header
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 {
				path = lines[len(lines)-1]

				// possibly trailing empty line
				if path == "" && len(lines) > 1 {
					path = lines[len(lines)-2]
				}
			}
		}
	}

	return runInShell(path, args...)
}

func runInShell(cmd string, args ...string) *exec.Cmd {
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
