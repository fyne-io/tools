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
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

var outfile = flag.String("o", "dex.go", "result will be written file")

var tmpdir string

func main() {
	flag.Parse()

	var err error
	tmpdir, err = os.MkdirTemp("", "gendex-")
	if err != nil {
		log.Fatal(err)
	}

	err = gendex()
	os.RemoveAll(tmpdir)
	if err != nil {
		log.Fatal(err)
	}
}

func gendex() error {
	src, err := os.ReadFile("classes.dex")
	if err != nil {
		return err
	}
	data := base64.StdEncoding.EncodeToString(src)

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, header)

	var piece string
	for len(data) > 0 {
		l := 70
		if l > len(data) {
			l = len(data)
		}
		piece, data = data[:l], data[l:]
		fmt.Fprintf(buf, "\t`%s` +\n", piece)
	}
	fmt.Fprintf(buf, "\t``")
	out := buf.Bytes()
	if err != nil {
		buf.WriteTo(os.Stderr)
		return err
	}

	w, err := os.Create(*outfile)
	if err != nil {
		return err
	}
	if _, err := w.Write(out); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
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
		binary.Write(&out, binary.BigEndian, uint16(count))
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

	return os.WriteFile(path, out.Bytes(), 0o644)
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
	binary.Write(out, binary.BigEndian, uint16(len(kept)))
	for _, e := range kept {
		out.Write(e)
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
