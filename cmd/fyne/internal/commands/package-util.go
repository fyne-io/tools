package commands

import (
	"os"

	"fyne.io/tools/cmd/fyne/internal/goos"
	"fyne.io/tools/cmd/fyne/internal/util"
)

type packagerUtil interface {
	Exists(path string) bool
	CopyFile(source string, target string) error
	CopyExeFile(src, tgt string) error
	WriteFile(target string, data []byte) error
	EnsureSubDir(parent, name string) string
	EnsureAbsPath(path string) string
	MakePathRelativeTo(root, path string) string

	RequireAndroidSDK() error
	AndroidBuildToolsPath() string

	IsAndroid(os string) bool
	IsIOS(os string) bool
	IsMobile(os string) bool
}

type defaultUtil struct{}

func (d defaultUtil) Exists(path string) bool {
	return util.Exists(path)
}

func (d defaultUtil) CopyFile(source string, target string) error {
	return util.CopyFile(source, target)
}

func (d defaultUtil) CopyExeFile(src, tgt string) error {
	return util.CopyExeFile(src, tgt)
}

func (d defaultUtil) WriteFile(target string, data []byte) error {
	return os.WriteFile(target, data, util.FilePermDefault)
}

func (d defaultUtil) EnsureSubDir(parent, name string) string {
	return util.EnsureSubDir(parent, name)
}

func (d defaultUtil) EnsureAbsPath(path string) string {
	return util.EnsureAbsPath(path)
}

func (d defaultUtil) MakePathRelativeTo(root, path string) string {
	return util.MakePathRelativeTo(root, path)
}

func (d defaultUtil) RequireAndroidSDK() error {
	return util.RequireAndroidSDK()
}

func (d defaultUtil) AndroidBuildToolsPath() string {
	return util.AndroidBuildToolsPath()
}

func (d defaultUtil) IsAndroid(os string) bool {
	return goos.IsAndroid(os)
}

func (d defaultUtil) IsIOS(os string) bool {
	return goos.IsIOS(os)
}

func (d defaultUtil) IsMobile(os string) bool {
	return goos.IsMobile(os)
}

var pkgUtil packagerUtil

func init() {
	pkgUtil = defaultUtil{}
}
