package commands

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/templates"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"

	"github.com/jackmordaunt/icns/v2"
)

type darwinData struct {
	Name, ExeName string
	AppID         string
	Version       string
	Build         int
	Category      string
	Languages     []string
}

const (
	icnsHeaderSize = 8
	icnsMagicSize  = 4
	icnsOSTypeDark = "\xFD\xD9\x2F\xA8"
)

func darwinLangs(langs []string) []string {
	r := make([]string, len(langs))
	for n, lang := range langs {
		r[n] = strings.Replace(lang, "-", "_", 1)
	}
	return r
}

func (p *Packager) packageDarwin() (err error) {
	appDir := util.EnsureSubDir(p.dir, p.Name+".app")
	exeName := filepath.Base(p.exe)

	contentsDir := util.EnsureSubDir(appDir, "Contents")
	info := filepath.Join(contentsDir, "Info.plist")
	infoFile, err := os.Create(info)
	if err != nil {
		return fmt.Errorf("failed to create plist template: %w", err)
	}
	defer func() {
		if r := infoFile.Close(); r != nil && err == nil {
			err = r
		}
	}()

	tplData := darwinData{
		Name: p.Name, ExeName: exeName, AppID: p.AppID, Version: p.AppVersion, Build: p.AppBuild,
		Category: strings.ToLower(p.category), Languages: darwinLangs(p.langs),
	}
	if err := templates.InfoPlistDarwin.Execute(infoFile, tplData); err != nil {
		return fmt.Errorf("failed to write plist template: %w", err)
	}

	macOSDir := util.EnsureSubDir(contentsDir, "MacOS")
	binName := filepath.Join(macOSDir, exeName)
	if err := util.CopyExeFile(p.exe, binName); err != nil {
		return fmt.Errorf("failed to copy executable: %w", err)
	}

	resDir := util.EnsureSubDir(contentsDir, "Resources")
	icnsPath := filepath.Join(resDir, "icon.icns")

	srcImg, err := loadIcon(p.icon)
	if err != nil {
		return err
	}
	if !p.rawIcon {
		srcImg = processMacOSIcon(srcImg)
	}
	buf := &bytes.Buffer{}
	if err := icns.Encode(buf, srcImg); err != nil {
		return fmt.Errorf("failed to encode icns: %w", err)
	}

	if p.darkIcon != "" {
		darkSrcImg, err := loadIcon(p.darkIcon)
		if err != nil {
			return err
		}
		if !p.rawIcon {
			darkSrcImg = processMacOSIcon(darkSrcImg)
		}
		if err := injectDarkIcon(darkSrcImg, buf); err != nil {
			return err
		}
	}
	dest, err := os.Create(icnsPath)
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	defer func() {
		if r := dest.Close(); r != nil && err == nil {
			err = r
		}
	}()
	if _, err := io.Copy(dest, buf); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

func loadIcon(icon string) (image.Image, error) {
	f, err := os.Open(icon)
	if err != nil {
		return nil, fmt.Errorf("failed to open source image %q: %w", icon, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode source image: %w", err)
	}

	return img, nil
}

func processMacOSIcon(in image.Image) image.Image {
	size := 1024
	border := 100
	radius := 185.4

	innerSize := int(float64(size) - float64(border*2)) // how many pixels inside border
	sized := resize.Resize(uint(innerSize), uint(innerSize), in, resize.Lanczos3)

	dc := gg.NewContext(size, size)
	dc.DrawRoundedRectangle(float64(border), float64(border), float64(innerSize), float64(innerSize), radius)
	dc.SetColor(color.Black)
	dc.Fill()
	mask := dc.AsMask()

	dc = gg.NewContext(size, size)
	_ = dc.SetMask(mask) // ignore error if size was not equal, as it is
	dc.DrawImage(sized, border, border)

	return dc.Image()
}

func injectDarkIcon(darkImg image.Image, buf *bytes.Buffer) error {
	buf2 := &bytes.Buffer{}
	if err := icns.Encode(buf2, darkImg); err != nil {
		return fmt.Errorf("failed to encode icns: %w", err)
	}

	if _, err := buf.WriteString(icnsOSTypeDark); err != nil {
		return fmt.Errorf("failed to append dark icon header: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, uint32(buf2.Len()-icnsHeaderSize)); err != nil {
		return fmt.Errorf("failed to append dark icon length: %w", err)
	}
	if _, err := buf.Write(buf2.Bytes()[icnsHeaderSize:]); err != nil {
		return fmt.Errorf("failed to append dark icon data: %w", err)
	}

	if _, err := binary.Encode(buf.Bytes()[icnsMagicSize:], binary.BigEndian, uint32(buf.Len())); err != nil {
		return fmt.Errorf("failed to adjust icon length: %w", err)
	}

	return nil
}
