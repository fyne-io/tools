// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Gendex generates a dex file used by Go apps created with gomobile.
//
// The dex is a thin extension of NativeActivity, providing access to
// a few platform features (not the SDK UI) not easily accessible from
// NDK headers. Long term these could be made part of the standard NDK,
// however that would limit gomobile to working with newer versions of
// the Android OS, so we do this while we wait.
//
// Requires ANDROID_HOME be set to the path of the Android SDK, and
// javac must be on the PATH.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/util"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/modfile"
)

const (
	androidRepo   = "https://dl.google.com/android/maven2/"
	javaFilesGlob = "internal/driver/mobile/app/*.java"
	javaDepsFile  = "internal/driver/mobile/app/java-dependencies.txt"
)

func main() {
	app := &cli.App{
		Name:  "gendex",
		Usage: "A Fyne tools command line helper to generate dex.go file.",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "set output file",
				Value:   "dex.go",
			},
			&cli.PathFlag{
				Name:        "source-dir",
				Aliases:     []string{"s"},
				Usage:       "set Fyne source directory",
				DefaultText: "fyne path in go.mod cache",
			},
			&cli.PathFlag{
				Name:        "build-dir",
				Aliases:     []string{"b"},
				Usage:       "set working directory for the build process",
				DefaultText: "temporary directory",
			},
			&cli.BoolFlag{
				Name:    "keep-build",
				Aliases: []string{"k"},
				Usage:   "set to prevent cleanup of build directory",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "enable verbose output",
			},
		},
		Action: doAction,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func doAction(c *cli.Context) error {
	buildDir := c.Path("build-dir")
	if buildDir == "" {
		dir, err := os.MkdirTemp("", "gendex-")
		if err != nil {
			return err
		}
		buildDir = dir
	}

	if !c.Bool("keep-build") {
		defer func() {
			if err := os.RemoveAll(buildDir); err != nil {
				log.Print(err)
			}
		}()
	}

	if c.Bool("verbose") {
		log.Printf("using build directory: %s", buildDir)
	}

	fyneSourceDir := c.Path("source-dir")
	if fyneSourceDir == "" {
		dir, err := util.LookupDirWithGoMod(".")
		if err != nil {
			return err
		}

		fyneSourceDir, err = lookupFyneDir(filepath.Join(dir, "go.mod"))
		if err != nil {
			return err
		}
	}

	return gendex(fyneSourceDir, buildDir, c.Path("output"), c.Bool("verbose"))
}

func lookupFyneDir(file string) (string, error) {
	var mod *modfile.File
	if data, err := os.ReadFile(file); err != nil {
		return "", err
	} else if mod, err = modfile.Parse(file, data, nil); err != nil {
		return "", err
	}

	for _, req := range mod.Require {
		if req.Mod.Path != "fyne.io/fyne/v2" {
			continue
		}
		out, err := exec.Command("go", "env", "GOPATH").CombinedOutput()
		if err != nil {
			return "", err
		}
		return filepath.Join(
			strings.TrimSpace(string(out)),
			"pkg",
			"mod",
			req.Mod.Path+"@"+req.Mod.Version,
		), nil
	}
	return "", fmt.Errorf("failed to find fyne source path")
}

func gendex(fyneSourceDir, buildDir, outfile string, verbose bool) error {
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		return errors.New("ANDROID_HOME not set")
	}
	buildTools, err := findLast(filepath.Join(androidHome, "build-tools"))
	if err != nil {
		return err
	}
	platform, err := findLast(filepath.Join(androidHome, "platforms"))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(buildDir, "work/org/golang/app"), util.DirPermDefault|util.PermGroupWrite); err != nil {
		return err
	}
	javaFiles, err := filepath.Glob(filepath.Join(fyneSourceDir, javaFilesGlob))
	if err != nil {
		return err
	}
	if len(javaFiles) == 0 {
		return errors.New("could not find files: " + javaFilesGlob)
	}
	if verbose {
		log.Printf("found java files: %v", javaFiles)
	}

	androidJar, err := getAndroidJar(platform, buildDir, verbose)
	if err != nil {
		return err
	}

	androidDeps, err := readAndroidDeps(filepath.Join(fyneSourceDir, javaDepsFile))
	if err != nil {
		return err
	}

	depDir := filepath.Join(buildDir, "deps")
	if err := os.MkdirAll(depDir, util.DirPermDefault); err != nil {
		return err
	}
	if err := downloadDeps(androidDeps, depDir, verbose); err != nil {
		return err
	}

	cmd := exec.Command(
		"javac",
		"-source", "1.8",
		"-target", "1.8",
		"-bootclasspath", androidJar,
		"-classpath", filepath.Join(depDir, "*"),
		"-d", filepath.Join(buildDir, "work"),
	)
	cmd.Args = append(cmd.Args, javaFiles...)
	if verbose {
		log.Printf("compiling java sources: %v", cmd.Args)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println(cmd.Args)
		os.Stderr.Write(out)
		return err
	}

	classFiles, err := filepath.Glob(filepath.Join(buildDir, "work/org/golang/app/*.class"))
	if err != nil {
		return err
	}

	// Strip the MethodParameters attribute from every method. javac emits this
	// for `mandated` synthetic enclosing-instance parameters of inner classes
	// with name_index=0; the AOSP-bundled R8 NPEs reading those entries. Since
	// MethodParameters is purely metadata for reflection, dropping it is safe.
	// Seems to be because of a compatibility issue between javac and d8
	for _, f := range classFiles {
		if err := stripMethodParameters(f); err != nil {
			return fmt.Errorf("strip MethodParameters %s: %w", f, err)
		}
	}

	cmd = exec.Command(
		filepath.Join(buildTools, "d8"),
		"--output", buildDir,
		"--lib", androidJar,
	)
	jarFiles, err := filepath.Glob(filepath.Join(depDir, "*.jar"))
	if err != nil {
		return err
	}
	cmd.Args = append(cmd.Args, jarFiles...)
	cmd.Args = append(cmd.Args, classFiles...)

	if verbose {
		log.Printf("building dex file: %v", cmd.Args)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		return err
	}

	if err := updateDexGo(filepath.Join(buildDir, "classes.dex"), outfile); err != nil {
		return err
	}

	return generateChecksums("SHA256SUMS", append([]string{androidJar}, jarFiles...), verbose)
}

func getAndroidJar(platform, buildDir string, verbose bool) (string, error) {
	androidVer := append(strings.Split(filepath.Base(platform), "android-"), "")[1]

	if verbose {
		log.Printf("found android version: %v", androidVer)
	}

	f, err := os.Open(filepath.Join(platform, "android.jar"))
	if err != nil {
		return "", err
	}

	androidJar := filepath.Join(buildDir, "android-"+androidVer+".jar")
	g, err := os.Create(androidJar)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(g, f); err != nil {
		_ = g.Close()
		return "", err
	}
	if err := g.Close(); err != nil {
		return "", err
	}

	return androidJar, nil
}

func readAndroidDeps(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := []string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		r = append(r, fields[0])
	}

	return r, nil
}

func downloadDeps(androidDeps []string, dir string, verbose bool) error {
	for _, dep := range androidDeps {
		parts := strings.Split(dep, ":")
		if len(parts) < 3 {
			return fmt.Errorf("invalid dependency: %v", dep)
		}
		jarFile := filepath.Join(dir, parts[1]+"-"+parts[2]+".jar")

		if fi, err := os.Stat(jarFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		} else if fi != nil && fi.Size() > 0 {
			if verbose {
				log.Printf("found existing file: %v", jarFile)
			}
			continue
		}

		pomPath := pomPathFromParts(parts)
		pomUrl := androidRepo + pomPath
		depUrl, err := getDownloadUrl(pomUrl)
		if err != nil {
			return err
		}

		dlFile := filepath.Join(dir, filepath.Base(depUrl))
		if verbose {
			log.Printf("downloading dependency: %v: %v", dlFile, depUrl)
		}
		b, err := download(depUrl)
		if err != nil {
			return err
		}

		switch filepath.Ext(dlFile) {
		case ".jar":
			if err := os.WriteFile(dlFile, b, util.FilePermDefault); err != nil {
				return err
			}
		case ".aar":
			if err := util.ExtractFileFromZipBytes(b, "classes.jar", jarFile); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown dependency format: %v", dlFile)
		}
	}
	return nil
}

func pomPathFromParts(parts []string) string {
	return strings.ReplaceAll(parts[0], ".", "/") + "/" + parts[1] + "/" + parts[2] + "/" + parts[1] + "-" + parts[2] + ".pom"
}

func getDownloadUrl(pomUrl string) (string, error) {
	pom, err := download(pomUrl)
	if err != nil {
		return "", err
	}
	var prj struct {
		Packaging string `xml:"packaging"`
	}
	if err := xml.Unmarshal(pom, &prj); err != nil {
		return "", err
	}
	ext := ".jar"
	if prj.Packaging != "" {
		ext = "." + prj.Packaging
	}
	return strings.TrimSuffix(pomUrl, ".pom") + ext, nil
}

func download(u string) ([]byte, error) {
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func updateDexGo(infile, outfile string) error {
	src, err := os.ReadFile(infile)
	if err != nil {
		return err
	}
	data := base64.StdEncoding.EncodeToString(src)

	buf := new(bytes.Buffer)
	buf.WriteString(header)
	buf.WriteRune('`')

	var piece string
	for len(data) > 0 {
		l := 70
		if l > len(data) {
			l = len(data)
		}
		piece, data = data[:l], data[l:]
		buf.WriteString(piece)
		buf.WriteRune('\n')
	}
	buf.Bytes()[buf.Len()-1] = '`'
	buf.WriteRune('\n')

	out, err := format.Source(buf.Bytes())
	if err != nil {
		_, _ = buf.WriteTo(os.Stderr)
		return err
	}

	w, err := os.Create(outfile)
	if err != nil {
		return err
	}
	if _, err := w.Write(out); err != nil {
		return err
	}
	return w.Close()
}

func generateChecksums(sumFile string, files []string, verbose bool) error {
	w, err := os.Create(sumFile)
	if err != nil {
		return err
	}

	var ww io.Writer = w
	if verbose {
		ww = io.MultiWriter(w, os.Stdout)
		log.Printf("updating checksum file: %v", sumFile)
	}
	for _, jarFile := range files {
		f, err := os.Open(jarFile)
		if err != nil {
			return err
		}
		h := sha256.New()
		_, err = io.Copy(h, f)
		_ = f.Close()
		if err != nil {
			return err
		}
		fmt.Fprintf(ww, "%x  %s\n", h.Sum(nil), filepath.Base(jarFile))
	}
	return w.Close()
}

// stripMethodParameters rewrites a .class file in place, removing every
// MethodParameters attribute from every method and field. The file's overall
// structure (constant pool, methods, fields, class attributes) is preserved.
func stripMethodParameters(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Locate the constant-pool entry (if any) named "MethodParameters" so we
	// can identify attributes by name index. If the pool has no such Utf8
	// entry the file cannot reference the attribute and we leave it alone.
	r := &classReader{buf: data}
	if r.u4() != 0xCAFEBABE {
		return errors.New("bad class magic")
	}
	r.skip(4) // minor + major
	cpCount := int(r.u2())
	mpIndex := uint16(0)
	for i := 1; i < cpCount; i++ {
		tag := r.u1()
		switch tag {
		case 1: // Utf8
			n := int(r.u2())
			s := r.bytes(n)
			if string(s) == "MethodParameters" {
				mpIndex = uint16(i)
			}
		case 7, 8, 16, 19, 20: // 2-byte payload
			r.skip(2)
		case 15: // MethodHandle
			r.skip(3)
		case 3, 4, 9, 10, 11, 12, 17, 18: // 4-byte payload
			r.skip(4)
		case 5, 6: // Long, Double - take 2 slots
			r.skip(8)
			i++
		default:
			return fmt.Errorf("unknown constant pool tag %d at index %d", tag, i)
		}
	}
	cpEnd := r.pos
	if mpIndex == 0 {
		return nil // attribute name not in pool; nothing to strip
	}

	// Build the rewritten file in a buffer.
	var out bytes.Buffer
	out.Write(data[:cpEnd])

	// access_flags + this_class + super_class
	out.Write(data[r.pos : r.pos+6])
	r.skip(6)

	// interfaces
	ic := int(r.u2())
	out.Write(data[cpEnd+6 : cpEnd+6+2+ic*2])
	r.skip(ic * 2)

	rewriteMembers := func() error {
		count := int(r.u2())
		if err := binary.Write(&out, binary.BigEndian, uint16(count)); err != nil {
			return nil
		}
		for i := 0; i < count; i++ {
			// access_flags(2) + name_index(2) + descriptor_index(2)
			out.Write(r.bytes(6))
			if err := rewriteAttributes(r, &out, mpIndex); err != nil {
				return err
			}
		}
		return nil
	}

	if err := rewriteMembers(); err != nil { // fields
		return err
	}
	if err := rewriteMembers(); err != nil { // methods
		return err
	}
	if err := rewriteAttributes(r, &out, mpIndex); err != nil { // class attrs
		return err
	}

	return os.WriteFile(path, out.Bytes(), util.FilePermDefault)
}

func rewriteAttributes(r *classReader, out *bytes.Buffer, dropName uint16) error {
	count := int(r.u2())
	kept := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		nameIdx := r.u2()
		length := r.u4()
		body := r.bytes(int(length))
		if nameIdx == dropName {
			continue
		}
		entry := make([]byte, 0, 6+len(body))
		var hdr [6]byte
		binary.BigEndian.PutUint16(hdr[0:2], nameIdx)
		binary.BigEndian.PutUint32(hdr[2:6], length)
		entry = append(entry, hdr[:]...)
		entry = append(entry, body...)
		kept = append(kept, entry)
	}
	if err := binary.Write(out, binary.BigEndian, uint16(len(kept))); err != nil {
		return err
	}
	for _, e := range kept {
		if _, err := out.Write(e); err != nil {
			return err
		}
	}
	return nil
}

type classReader struct {
	buf []byte
	pos int
}

func (r *classReader) u1() byte {
	v := r.buf[r.pos]
	r.pos++
	return v
}

func (r *classReader) u2() uint16 {
	v := binary.BigEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v
}

func (r *classReader) u4() uint32 {
	v := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v
}

func (r *classReader) skip(n int) { r.pos += n }

func (r *classReader) bytes(n int) []byte {
	b := r.buf[r.pos : r.pos+n]
	r.pos += n
	return b
}

func findLast(path string) (string, error) {
	dir, err := os.Open(path)
	if err != nil {
		return "", err
	}
	children, err := dir.Readdirnames(-1)
	if err != nil {
		return "", err
	}
	sort.Strings(children)
	return path + "/" + children[len(children)-1], nil
}

var header = `// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Code generated by gendex.go. DO NOT EDIT.

package mobile

var dexStr = `
