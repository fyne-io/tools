package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/tools/cmd/fyne/internal/mobile"
	"fyne.io/tools/cmd/fyne/internal/templates"
	"fyne.io/tools/cmd/fyne/internal/util"
)

// splashIconSize is the point size of the centred launch screen icon, matching the 288dp
// drawable size of the Android 12+ system splash so the launch looks the same on both platforms.
const splashIconSize = 288

func (p *Packager) packageAndroid(arch string, tags []string) error {
	iconFG, iconBG, iconMono := "", "", ""
	if p.AdaptiveIcon != nil {
		iconFG = p.AdaptiveIcon.Foreground
		iconBG = p.AdaptiveIcon.Background
		iconMono = p.AdaptiveIcon.Monochrome
	}

	return mobile.RunNewBuild(arch, p.AppID, p.icon, p.Name, p.AppVersion, p.AppBuild, p.release, p.distribution,
		"", "", tags, iconFG, iconBG, iconMono, p.verbose, p.appData.Splash)
}

func (p *Packager) packageIOS(target string, tags []string) error {
	err := mobile.RunNewBuild(target, p.AppID, p.icon, p.Name, p.AppVersion, p.AppBuild, p.release, p.distribution,
		p.certificate, p.profile, tags, "", "", "", p.verbose, nil)
	if err != nil {
		return err
	}

	assetDir := pkgUtil.EnsureSubDir(p.dir, "Images.xcassets")
	defer os.RemoveAll(assetDir)
	err = os.WriteFile(filepath.Join(assetDir, "Contents.json"), []byte(`{
  "info" : {
    "author" : "xcode",
    "version" : 1
  }
}`), util.FilePermDefault)
	if err != nil {
		fyne.LogError("Content err", err)
	}

	iconDir := pkgUtil.EnsureSubDir(assetDir, "AppIcon.appiconset")
	contentFile, _ := os.Create(filepath.Join(iconDir, "Contents.json"))

	err = templates.XCAssetsDarwin.Execute(contentFile, nil)
	if err != nil {
		return fmt.Errorf("failed to write xcassets content template: %w", err)
	}

	iconSizes := []int{76, 120, 152, 180, 1024} //revive:disable-line:add-constant
	for _, iconSize := range iconSizes {
		if err = copyResizeIcon(iconSize, iconDir, p.icon); err != nil {
			return err
		}
	}

	if p.appData.Splash != nil {
		splashIcon := p.appData.Splash.Icon
		if splashIcon == "" {
			splashIcon = p.icon
		}
		if err = writeSplashIconSet(assetDir, splashIcon); err != nil {
			return err
		}
	}

	appDir := filepath.Join(p.dir, mobile.AppOutputName(p.os, p.Name, p.release))
	if p.verbose {
		fmt.Println("Creating icons for", relPath(appDir))
	}
	err = runCmdCaptureOutput("xcrun", "actool", "Images.xcassets", "--compile", appDir, "--platform",
		"iphoneos", "--target-device", "iphone", "--minimum-deployment-target", "9.0", "--app-icon", "AppIcon",
		"--output-format", "human-readable-text", "--output-partial-info-plist", "/dev/null")
	if err != nil {
		return err
	}

	if p.appData.Splash != nil {
		if err = compileLaunchScreen(p.dir, appDir, p.appData.Splash.Background); err != nil {
			return err
		}
	}

	if target == "iossimulator" {
		// The build signed the bundle before the resources above were added, and iOS ignores
		// a launch storyboard that is not sealed by the signature, so sign it again.
		return runCmdCaptureOutput("codesign", "--force", "--sign", "-", appDir)
	}
	return nil
}

// writeSplashIconSet adds the launch screen icon to the asset catalog as SplashIcon, scaled for 3x displays.
func writeSplashIconSet(assetDir, icon string) error {
	imageDir := util.EnsureSubDir(assetDir, "SplashIcon.imageset")
	if err := util.WriteScaledPNG(icon, filepath.Join(imageDir, "splash.png"), splashIconSize*3); err != nil {
		return fmt.Errorf("failed to write splash icon: %w", err)
	}

	return os.WriteFile(filepath.Join(imageDir, "Contents.json"), []byte(`{
  "images" : [
    {
      "filename" : "splash.png",
      "idiom" : "universal",
      "scale" : "3x"
    }
  ],
  "info" : {
    "author" : "xcode",
    "version" : 1
  }
}`), 0o644)
}

// compileLaunchScreen builds the LaunchScreen storyboard named in Info.plist into the app bundle,
// with the splash background colour and the centred SplashIcon.
func compileLaunchScreen(dir, appDir, background string) error {
	r, g, b, err := parseHexColor(background)
	if err != nil {
		return fmt.Errorf("invalid splash background colour %q: %w", background, err)
	}
	data := struct {
		Red, Green, Blue float64
		IconSize         int
	}{Red: r, Green: g, Blue: b, IconSize: splashIconSize}

	storyboard := filepath.Join(dir, "LaunchScreen.storyboard")
	out, err := os.Create(storyboard)
	if err != nil {
		return err
	}
	defer os.Remove(storyboard)
	err = templates.LaunchScreenIOS.Execute(out, data)
	_ = out.Close()
	if err != nil {
		return fmt.Errorf("failed to write launch screen storyboard: %w", err)
	}

	return runCmdCaptureOutput("xcrun", "ibtool", "--compile", filepath.Join(appDir, "LaunchScreen.storyboardc"),
		"--target-device", "iphone", "--minimum-deployment-target", "9.0", storyboard)
}

// parseHexColor converts a "#RRGGBB" colour into 0-1 components, defaulting to white.
func parseHexColor(hex string) (r, g, b float64, err error) {
	if hex == "" {
		return 1, 1, 1, nil
	}
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, errors.New("expected #RRGGBB")
	}
	var channels [3]float64
	for i := range channels {
		v, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return 0, 0, 0, err
		}
		channels[i] = float64(v) / 255
	}
	return channels[0], channels[1], channels[2], nil
}

func copyResizeIcon(size int, dir, source string) error {
	strSize := strconv.Itoa(size)
	path := dir + "/Icon_" + strSize + ".png"
	return runCmdCaptureOutput("sips", "-o", path, "-Z", strSize, source)
}

// runCmdCaptureOutput is a exec.Command wrapper that offers better error messages from stdout and stderr.
func runCmdCaptureOutput(name string, args ...string) error {
	var (
		outbuf = &bytes.Buffer{}
		errbuf = &bytes.Buffer{}
	)
	cmd := exec.Command(name, args...)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf
	err := cmd.Run()
	if err != nil {
		outstr := outbuf.String()
		errstr := errbuf.String()
		if outstr != "" {
			err = fmt.Errorf(outbuf.String()+": %w", err)
		}
		if errstr != "" {
			err = fmt.Errorf(outbuf.String()+": %w", err)
		}
		return err
	}
	return nil
}
