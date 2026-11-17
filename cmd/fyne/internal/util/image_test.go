package util

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteScaledPNG(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.png")
	f, err := os.Create(src)
	assert.Nil(t, err)
	assert.Nil(t, png.Encode(f, image.NewRGBA(image.Rect(0, 0, 64, 64))))
	assert.Nil(t, f.Close())

	dst := filepath.Join(dir, "out.png")
	assert.Nil(t, WriteScaledPNG(src, dst, 16))

	f, err = os.Open(dst)
	assert.Nil(t, err)
	defer f.Close()
	conf, err := png.DecodeConfig(f)
	assert.Nil(t, err)
	assert.Equal(t, 16, conf.Width)
	assert.Equal(t, 16, conf.Height)

	assert.NotNil(t, WriteScaledPNG(filepath.Join(dir, "missing.png"), dst, 16))
}
