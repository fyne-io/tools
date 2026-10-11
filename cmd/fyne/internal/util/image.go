package util

import (
	"fmt"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

// WriteScaledPNG decodes the PNG at src, scales it to size pixels square and writes it to dst.
func WriteScaledPNG(src, dst string, size int) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	img, err := png.Decode(in)
	_ = in.Close()
	if err != nil {
		return fmt.Errorf("failed to decode PNG %s: %w", src, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	err = png.Encode(out, resize.Resize(uint(size), uint(size), img, resize.Lanczos3))
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
}
