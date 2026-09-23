package commands

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/templates"
)

func (p *Packager) packageWasm() error {
	appDir := pkgUtil.EnsureSubDir(p.dir, "wasm")

	tpl := webData{
		AppName:    p.Name,
		AppVersion: p.AppVersion,
		WasmFile:   p.Name + ".wasm",
		IsReleased: p.release,
	}

	return tpl.packageWebInternal(appDir, p.exe, p.icon, p.release)
}

type webData struct {
	AppName    string
	AppVersion string
	WasmFile   string
	IsReleased bool
}

const fileWasmExecJs = "wasm_exec.js"

func (w webData) packageWebInternal(appDir string, exeWasmSrc string, icon string, release bool) error {
	var tpl bytes.Buffer
	err := templates.IndexHTML.Execute(&tpl, w)
	if err != nil {
		return err
	}

	index := filepath.Join(appDir, "index.html")
	err = pkgUtil.WriteFile(index, tpl.Bytes())
	if err != nil {
		return err
	}

	iconDst := filepath.Join(appDir, "icon.png")
	err = pkgUtil.CopyFile(icon, iconDst)
	if err != nil {
		return err
	}

	spinnerLightFile := filepath.Join(appDir, "spinner_light.gif")
	err = pkgUtil.WriteFile(spinnerLightFile, templates.SpinnerLight)
	if err != nil {
		return err
	}

	spinnerDarkFile := filepath.Join(appDir, "spinner_dark.gif")
	err = pkgUtil.WriteFile(spinnerDarkFile, templates.SpinnerDark)
	if err != nil {
		return err
	}

	lightCSSFile := filepath.Join(appDir, "light.css")
	err = pkgUtil.WriteFile(lightCSSFile, templates.CSSLight)
	if err != nil {
		return err
	}

	darkCSSFile := filepath.Join(appDir, "dark.css")
	err = pkgUtil.WriteFile(darkCSSFile, templates.CSSDark)
	if err != nil {
		return err
	}

	goroot, err := GOROOT()
	if err != nil {
		return err
	}

	wasmFileWasmExecJs := filepath.Join("wasm", fileWasmExecJs)
	wasmExecSrc := filepath.Join(goroot, "lib", wasmFileWasmExecJs)
	if !pkgUtil.Exists(wasmExecSrc) { // Fallback for Go < 1.24:
		wasmExecSrc = filepath.Join(goroot, "misc", wasmFileWasmExecJs)
	}

	wasmExecDst := filepath.Join(appDir, fileWasmExecJs)
	err = pkgUtil.CopyFile(wasmExecSrc, wasmExecDst)
	if err != nil {
		return err
	}

	exeWasmDst := filepath.Join(appDir, w.WasmFile)
	err = pkgUtil.CopyFile(exeWasmSrc, exeWasmDst)
	if err != nil {
		return err
	}

	// Download webgl-debug.js directly from the KhronosGroup repository when needed
	if !release {
		webglDebugFile := filepath.Join(appDir, "webgl-debug.js")
		err := pkgUtil.WriteFile(webglDebugFile, templates.WebGLDebugJs)
		if err != nil {
			return err
		}
	}

	return nil
}

// GOROOT returns the root of the go binary location.
func GOROOT() (string, error) {
	output, err := exec.Command("go", "env", "GOROOT").Output()
	return strings.TrimSuffix(string(output), "\n"), err
}
