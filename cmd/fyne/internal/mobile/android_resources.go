package mobile

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/tools/cmd/fyne/internal/mobile/binres"
	"fyne.io/tools/cmd/fyne/internal/util"
)

// generateAdaptiveIconXML creates the XML definition for an adaptive icon
func generateAdaptiveIconXML(hasMonochrome bool) string {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@mipmap/ic_launcher_background"/>
    <foreground android:drawable="@mipmap/ic_launcher_foreground"/>`

	if hasMonochrome {
		xml += `
    <monochrome android:drawable="@mipmap/ic_launcher_monochrome"/>`
	}

	xml += `
</adaptive-icon>
`
	return xml
}

// writeAdaptiveIconResources creates all necessary resource files for adaptive icons
func writeAdaptiveIconResources(resDir, foregroundPath, backgroundPath, monochromePath string) error {
	xxxhdpiDir := filepath.Join(resDir, "mipmap-xxxhdpi")
	if err := os.MkdirAll(xxxhdpiDir, 0o755); err != nil {
		return fmt.Errorf("failed to create mipmap-xxxhdpi directory: %w", err)
	}

	if err := util.CopyFile(foregroundPath, filepath.Join(xxxhdpiDir, "ic_launcher_foreground.png")); err != nil {
		return fmt.Errorf("failed to copy foreground icon: %w", err)
	}

	if backgroundPath != "" {
		if err := util.CopyFile(backgroundPath, filepath.Join(xxxhdpiDir, "ic_launcher_background.png")); err != nil {
			return fmt.Errorf("failed to copy background icon: %w", err)
		}
	}

	hasMonochrome := monochromePath != ""
	if hasMonochrome {
		if err := util.CopyFile(monochromePath, filepath.Join(xxxhdpiDir, "ic_launcher_monochrome.png")); err != nil {
			return fmt.Errorf("failed to copy monochrome icon: %w", err)
		}
	}

	// Create mipmap-anydpi-v26 metadata for adaptive icon XML
	anydpiV26Dir := filepath.Join(resDir, "mipmap-anydpi-v26")
	if err := os.MkdirAll(anydpiV26Dir, 0o755); err != nil {
		return fmt.Errorf("failed to create mipmap-anydpi-v26 directory: %w", err)
	}

	adaptiveXML := generateAdaptiveIconXML(false) // v26 doesn't include monochrome
	if err := os.WriteFile(filepath.Join(anydpiV26Dir, "ic_launcher.xml"), []byte(adaptiveXML), 0o644); err != nil {
		return fmt.Errorf("failed to write ic_launcher.xml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(anydpiV26Dir, "ic_launcher_round.xml"), []byte(adaptiveXML), 0o644); err != nil {
		return fmt.Errorf("failed to write ic_launcher_round.xml: %w", err)
	}

	if !hasMonochrome {
		return nil
	}
	// If monochrome provided, create v33 resources with monochrome support
	anydpiV33Dir := filepath.Join(resDir, "mipmap-anydpi-v33")
	if err := os.MkdirAll(anydpiV33Dir, 0o755); err != nil {
		return fmt.Errorf("failed to create mipmap-anydpi-v33 directory: %w", err)
	}

	adaptiveXMLWithMono := generateAdaptiveIconXML(true)
	if err := os.WriteFile(filepath.Join(anydpiV33Dir, "ic_launcher.xml"), []byte(adaptiveXMLWithMono), 0o644); err != nil {
		return fmt.Errorf("failed to write v33 ic_launcher.xml: %w", err)
	}

	if err := os.WriteFile(filepath.Join(anydpiV33Dir, "ic_launcher_round.xml"), []byte(adaptiveXMLWithMono), 0o644); err != nil {
		return fmt.Errorf("failed to write v33 ic_launcher_round.xml: %w", err)
	}
	return nil
}

// compileAndroidResources compiles Android resources using aapt2
// Returns: resources.arsc path, res/ directory path, compiled AndroidManifest.xml path, error
func compileAndroidResources(tempDir string, manifestData []byte, foregroundPath, backgroundPath, monochromePath string, targetSDK, versionCode int, versionName, packageName string) (arscPath string, resDir string, manifestPath string, err error) {
	aapt2, err := util.Aapt2Path()
	if err != nil {
		return "", "", "", err
	}

	resDir = filepath.Join(tempDir, "res")
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("failed to create res directory: %w", err)
	}

	if err := writeAdaptiveIconResources(resDir, foregroundPath, backgroundPath, monochromePath); err != nil {
		return "", "", "", err
	}

	compiledDir := filepath.Join(tempDir, "compiled")
	if err := os.MkdirAll(compiledDir, 0o755); err != nil {
		return "", "", "", fmt.Errorf("failed to create compiled directory: %w", err)
	}

	tempManifestPath := filepath.Join(tempDir, "AndroidManifest.xml")
	if err := os.WriteFile(tempManifestPath, manifestData, 0o644); err != nil {
		return "", "", "", fmt.Errorf("failed to write AndroidManifest.xml: %w", err)
	}

	err = filepath.Walk(resDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		cmd := exec.Command(aapt2, "compile", "-o", compiledDir, path)
		if buildV {
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("aapt2 compile failed for %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return "", "", "", err
	}

	androidJar, err := binres.LatestAPIResourcesPath()
	if err != nil || !util.Exists(androidJar) {
		return "", "", "", fmt.Errorf("android.jar not found for API %d at %s", binres.MinSDK, androidJar)
	}

	outputAPK := filepath.Join(tempDir, "resources.apk")
	linkArgs := []string{
		"link",
		"-o", outputAPK,
		"--manifest", tempManifestPath,
		"--min-sdk-version", "21", // Android 5.0, minimum for adaptive icons (v26) with backward compat
		"--target-sdk-version", fmt.Sprintf("%d", targetSDK),
		"--version-code", fmt.Sprintf("%d", versionCode),
		"--version-name", versionName,
		"-I", androidJar,
		"--auto-add-overlay",
	}

	flatFiles, err := filepath.Glob(filepath.Join(compiledDir, "*.flat"))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to find compiled flat files: %w", err)
	}
	linkArgs = append(linkArgs, flatFiles...)

	cmd := exec.Command(aapt2, linkArgs...)
	if buildV {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		printcmd("%s %v", aapt2, linkArgs)
	}
	if err := cmd.Run(); err != nil {
		return "", "", "", fmt.Errorf("aapt2 link failed: %w", err)
	}

	// Extract resources.arsc from the output APK
	arscPath = filepath.Join(tempDir, "resources.arsc")
	if err := extractFileFromZip(outputAPK, "resources.arsc", arscPath); err != nil {
		return "", "", "", fmt.Errorf("failed to extract resources.arsc: %w", err)
	}

	// Extract compiled AndroidManifest.xml from the output APK
	manifestPath = filepath.Join(tempDir, "AndroidManifest.xml")
	if err := extractFileFromZip(outputAPK, "AndroidManifest.xml", manifestPath); err != nil {
		return "", "", "", fmt.Errorf("failed to extract AndroidManifest.xml: %w", err)
	}

	// Extract res/ directory from the output APK
	extractedResDir := filepath.Join(tempDir, "extracted_res")
	if err := extractDirFromZip(outputAPK, "res/", extractedResDir); err != nil {
		return "", "", "", fmt.Errorf("failed to extract res/ directory: %w", err)
	}

	return arscPath, extractedResDir, manifestPath, nil
}

// extractFileFromZip extracts a single file from a zip archive
func extractFileFromZip(zipPath, fileName, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != fileName {
			continue
		}

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
	return fmt.Errorf("file %s not found in zip", fileName)
}

// extractDirFromZip extracts all files with a given prefix from a zip archive
func extractDirFromZip(zipPath, prefix, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
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
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		dest, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(dest, rc)
		rc.Close()
		dest.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
