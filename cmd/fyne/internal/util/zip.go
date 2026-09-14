package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"bytes"
)

func copyZipFileToPath(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, rc)
	return err
}

// ExtractFileFromZip extracts a single file from a zip archive
func ExtractFileFromZip(zipPath, fileName, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	return extractFileFromZip(&r.Reader, fileName, destPath)
}

// ExtractFileFromZipReader extracts a single file from an open zip archive handle
func ExtractFileFromZipBytes(b []byte, fileName, destPath string) error {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	return extractFileFromZip(r, fileName, destPath)
}

func extractFileFromZip(r *zip.Reader, fileName, destPath string) error {
	for _, f := range r.File {
		if f.Name != fileName {
			continue
		}

		return copyZipFileToPath(f, destPath)
	}
	return fmt.Errorf("file %s not found in zip", fileName)
}

// ExtractDirFromZip extracts all files with a given prefix from a zip archive
func ExtractDirFromZip(zipPath, prefix, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, DirPermDefault); err != nil {
		return err
	}

	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, prefix) {
			continue
		}

		relPath := f.Name[len(prefix):]
		if relPath == "" {
			continue
		}
		destPath := filepath.Join(destDir, relPath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, DirPermDefault); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), DirPermDefault); err != nil {
			return err
		}

		// Extract file
		if err := copyZipFileToPath(f, destPath); err != nil {
			return err
		}
	}

	return nil
}
