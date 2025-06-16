// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobile

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func mkdir(dir string) error {
	if buildX || buildN {
		printcmd("mkdir -p %s", dir)
	}
	if buildN {
		return nil
	}
	return os.MkdirAll(dir, 0o750)
}

func removeAll(path string) error {
	if buildX || buildN {
		printcmd(`rm -r -f "%s"`, path)
	}
	if buildN {
		return nil
	}

	return os.RemoveAll(path)
}

func goEnv(name string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	val, err := exec.Command("go", "env", name).Output()
	if err != nil {
		panic(err) // the Go tool was tested to work earlier
	}
	return strings.TrimSpace(string(val))
}

func runCmd(cmd *exec.Cmd) error {
	if buildX || buildN {
		dir := ""
		if cmd.Dir != "" {
			dir = "PWD=" + cmd.Dir + " "
		}
		env := strings.Join(cmd.Env, " ")
		if env != "" {
			env += " "
		}
		printcmd("%s%s%s", dir, env, strings.Join(cmd.Args, " "))
	}

	buf := new(bytes.Buffer)
	buf.WriteByte('\n')
	if buildV {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = buf
		cmd.Stderr = buf
	}

	if buildWork {
		if runtime.GOOS == "windows" {
			cmd.Env = append(cmd.Env, `TEMP=`+tmpdir, `TMP=`+tmpdir)
		} else {
			cmd.Env = append(cmd.Env, `TMPDIR=`+tmpdir)
		}
	}

	if !buildN {
		cmd.Env = environ(cmd.Env)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s failed: %v%s", strings.Join(cmd.Args, " "), err, buf)
		}
	}
	return nil
}
