package commands

import (
	"fmt"
	"image"
	"image/color"
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

func darwinLangs(langs []string) []string {
	r := make([]string, len(langs))
	for n, lang := range langs {
		r[n] = strings.Replace(lang, "-", "_", 1)
	}
	return r
}

func (p *Packager) packageDarwin() (err error) {
	appDir := pkgUtil.EnsureSubDir(p.dir, p.Name+".app")
	exeName := filepath.Base(p.exe)

	contentsDir := pkgUtil.EnsureSubDir(appDir, "Contents")
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

	macOSDir := pkgUtil.EnsureSubDir(contentsDir, "MacOS")
	binName := filepath.Join(macOSDir, exeName)
	if err := pkgUtil.CopyExeFile(p.exe, binName); err != nil {
		return fmt.Errorf("failed to copy executable: %w", err)
	}

	resDir := pkgUtil.EnsureSubDir(contentsDir, "Resources")
	icnsPath := filepath.Join(resDir, "icon.icns")

	img, err := os.Open(p.icon)
	if err != nil {
		return fmt.Errorf("failed to open source image \"%s\": %w", p.icon, err)
	}
	defer img.Close()
	srcImg, _, err := image.Decode(img)
	if err != nil {
		return fmt.Errorf("failed to decode source image: %w", err)
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
	if !p.rawIcon {
		srcImg = processMacOSIcon(srcImg)
	}
	if err := icns.Encode(dest, srcImg); err != nil {
		return fmt.Errorf("failed to encode icns: %w", err)
	}

	return nil
}

const (
	macosIconSize   = 1024
	macosIconBorder = 100.0
	macosIconRadius = 185.4
)

func processMacOSIcon(in image.Image) image.Image {
	innerSize := float64(macosIconSize) - macosIconBorder*2 // how many pixels inside border
	sized := resize.Resize(uint(innerSize), uint(innerSize), in, resize.Lanczos3)

	dc := gg.NewContext(macosIconSize, macosIconSize)
	dc.DrawRoundedRectangle(macosIconBorder, macosIconBorder, innerSize, innerSize, macosIconRadius)
	dc.SetColor(color.Black)
	dc.Fill()
	mask := dc.AsMask()

	dc = gg.NewContext(macosIconSize, macosIconSize)
	_ = dc.SetMask(mask) // ignore error if size was not equal, as it is
	dc.DrawImage(sized, int(macosIconBorder), int(macosIconBorder))

	return dc.Image()
}
