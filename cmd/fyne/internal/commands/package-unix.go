package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/metadata"
	"fyne.io/tools/cmd/fyne/internal/templates"
)

type unixData struct {
	Name        string
	AppID       string
	Exec        string
	Icon        string
	Local       string
	GenericName string
	Categories  string
	Comment     string
	Keywords    string
	ExecParams  string
	MimeTypes   string

	SourceRepo, SourceDir string
}

func (p *Packager) packageUNIX() error {
	var prefixDir string
	local := "local/"
	dirName := filepath.Base(p.exe)

	outDir := p.dir
	if !p.install {
		outDir = util.EnsureSubDir(util.EnsureSubDir(p.dir, "tmp-pkg"), dirName)
	}

	if _, err := os.Stat(filepath.Join("/", "usr", "local")); os.IsNotExist(err) {
		prefixDir = util.EnsureSubDir(outDir, "usr")
		local = ""
	} else {
		prefixDir = util.EnsureSubDir(util.EnsureSubDir(outDir, "usr"), "local")
	}

	shareDir := util.EnsureSubDir(prefixDir, "share")

	binDir := util.EnsureSubDir(prefixDir, "bin")
	binName := filepath.Join(binDir, filepath.Base(p.exe))
	err := util.CopyExeFile(p.exe, binName)
	if err != nil {
		return fmt.Errorf("failed to copy application binary file: %w", err)
	}

	appIDOrName := p.AppID
	if appIDOrName == "" {
		appIDOrName = p.Name
	}

	iconDir := util.EnsureSubDir(shareDir, "pixmaps")
	iconName := appIDOrName + filepath.Ext(p.icon)
	iconPath := filepath.Join(iconDir, iconName)
	err = util.CopyFile(p.icon, iconPath)
	if err != nil {
		return fmt.Errorf("failed to copy icon: %w", err)
	}

	mimes := ""
	openWith := ""
	if p.CanOpen != nil && p.CanOpen.MimeTypes != "" {
		mimes = p.CanOpen.MimeTypes
		openWith = " %F"
	}

	appsDir := util.EnsureSubDir(shareDir, "applications")
	desktop := filepath.Join(appsDir, appIDOrName+".desktop")
	deskFile, err := os.Create(desktop)
	if err != nil {
		return fmt.Errorf("failed to create desktop file: %w", err)
	}

	defer deskFile.Close()

	linuxBSD := metadata.LinuxAndBSD{}
	if p.linuxAndBSDMetadata != nil {
		linuxBSD = *p.linuxAndBSDMetadata
	}
	tplData := unixData{
		Name:        p.Name,
		AppID:       p.AppID,
		Exec:        filepath.Base(p.exe) + openWith,
		Icon:        appIDOrName,
		Local:       local,
		GenericName: linuxBSD.GenericName,
		Keywords:    formatDesktopFileList(linuxBSD.Keywords),
		Comment:     linuxBSD.Comment,
		Categories:  formatDesktopFileList(linuxBSD.Categories),
		ExecParams:  linuxBSD.ExecParams,
		MimeTypes:   mimes,
	}

	if p.sourceMetadata != nil {
		tplData.SourceRepo = p.sourceMetadata.Repo
		tplData.SourceDir = p.sourceMetadata.Dir
	}

	err = templates.DesktopFileUNIX.Execute(deskFile, tplData)
	if err != nil {
		return fmt.Errorf("failed to write desktop entry string: %w", err)
	}

	if !p.install {
		parent := filepath.Dir(outDir)
		defer os.RemoveAll(parent)

		makefile, _ := os.Create(filepath.Join(outDir, "Makefile"))
		err := templates.MakefileUNIX.Execute(makefile, tplData)
		if err != nil {
			return fmt.Errorf("failed to write Makefile string: %w", err)
		}

		tarCmdArgs := []string{"-Jcf", filepath.Join(p.dir, p.Name+".tar.xz")}
		if p.os == "openbsd" {
			tarCmdArgs = []string{"-zcf", filepath.Join(p.dir, p.Name+".tar.gz")}
		}
		tarCmdArgs = append(tarCmdArgs, "-C", parent, dirName)

		var buf bytes.Buffer
		tarCmd := exec.Command("tar", tarCmdArgs...)
		tarCmd.Stderr = &buf
		if err = tarCmd.Run(); err != nil {
			return fmt.Errorf("failed to create archive with tar: %s - %w", buf.String(), err)
		}
	}

	return nil
}

func formatDesktopFileList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	return strings.Join(items, ";") + ";"
}
