package util

import (
	"errors"
	"os"
	"path/filepath"
)

// LookupDirWithGoMod takes a directory and checks for a go.mod file, traverses back towards the root,
// and returns the first directory with a match. In case of a relative path the traversal stops at
// the relative root
func LookupDirWithGoMod(workDir string) (string, error) {
	isRelative := !filepath.IsAbs(workDir)
	relDir := ""
	volName := filepath.VolumeName(workDir)

	if isRelative {
		relDir = workDir
		for {
			dir, file := filepath.Split(relDir)
			dir = filepath.Clean(dir)
			if dir == "" || dir == "." || dir == ".." || dir == volName || file == "" || file == "." || file == ".." {
				break
			}
			relDir = filepath.Clean(dir)
		}

		if absDir, err := filepath.Abs(relDir); err != nil {
			return "", err
		} else {
			relDir = absDir
		}

		if absDir, err := filepath.Abs(workDir); err != nil {
			return "", err
		} else {
			workDir = absDir
		}
	}

	for {
		fi, err := os.Stat(filepath.Join(workDir, "go.mod"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if fi != nil {
			break
		}
		parentDir := filepath.Dir(workDir)
		if parentDir == workDir || isRelative && parentDir == relDir {
			return "", os.ErrNotExist
		}
		workDir = parentDir
	}

	return workDir, nil
}
