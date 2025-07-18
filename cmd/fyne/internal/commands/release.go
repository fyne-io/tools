package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"fyne.io/fyne/v2"
	"fyne.io/tools/cmd/fyne/internal/mobile"
	"fyne.io/tools/cmd/fyne/internal/templates"

	"github.com/urfave/cli/v2"
)

var macAppStoreCategories = []string{
	"business", "developer-tools", "education", "entertainment", "finance", "games", "action-games",
	"adventure-games", "arcade-games", "board-games", "card-games", "casino-games", "dice-games",
	"educational-games", "family-games", "kids-games", "music-games", "puzzle-games", "racing-games",
	"role-playing-games", "simulation-games", "sports-games", "strategy-games", "trivia-games", "word-games",
	"graphics-design", "healthcare-fitness", "lifestyle", "medical", "music", "news", "photography",
	"productivity", "reference", "social-networking", "sports", "travel", "utilities", "video", "weather",
}

// Release returns the cli command for bundling release builds of fyne applications
func Release() *cli.Command {
	r := NewReleaser()

	return &cli.Command{
		Name:    "release",
		Aliases: []string{"r"},
		Usage:   "Prepares an application for public distribution",
		Flags: []cli.Flag{
			stringFlags["target"](&r.os),
			stringFlags["key-store"](&r.keyStore),
			stringFlags["key-store-pass"](&r.keyStorePass),
			stringFlags["key-name"](&r.keyName),
			stringFlags["key-pass"](&r.keyStorePass),
			stringFlags["name"](&r.Name),
			stringFlags["tags"](&r.tags),
			stringFlags["app-version"](&r.AppVersion),
			intFlags["app-build"](&r.AppBuild),
			stringFlags["app-id"](&r.AppID),
			stringFlags["certificate"](&r.certificate),
			stringFlags["profile"](&r.profile),
			stringFlags["developer"](&r.developer),
			stringFlags["password"](&r.password),
			stringFlags["category"](&r.category),
			stringFlags["icon"](&r.icon),
			boolFlags["use-raw-icon"](&r.rawIcon),
			genericFlags["metadata"](&r.customMetadata),
		},
		Action: r.releaseAction,
	}
}

// Releaser adapts app packages form distribution.
type Releaser struct {
	Packager

	keyStore     string
	keyStorePass string
	keyName      string
	keyPass      string
	developer    string
	password     string
}

// NewReleaser returns a command that can handle the packaging a GUI apps for release from local Fyne source code.
func NewReleaser() *Releaser {
	r := &Releaser{}
	r.appData = &appData{}
	return r
}

// AddFlags adds the flags for interacting with the release command.
//
// Deprecated: Access to the individual cli commands are being removed.
func (r *Releaser) AddFlags() {
	flag.StringVar(&r.os, "os", "", "The operating system to target (android, android/arm, android/arm64, android/amd64, android/386, darwin, freebsd, ios, linux, netbsd, openbsd, windows)")
	flag.StringVar(&r.Name, "name", "", "The name of the application, default is the executable file name")
	flag.StringVar(&r.icon, "icon", "", "The name of the application icon file")
	flag.StringVar(&r.AppID, "app-id", "", "For ios or darwin targets an app-id is required, for ios this must \nmatch a valid provisioning profile")
	flag.StringVar(&r.AppVersion, "app-version", "", "Version number in the form x, x.y or x.y.z semantic version")
	flag.IntVar(&r.AppBuild, "app-build", 0, "Build number, should be greater than 0 and incremented for each build")
	flag.StringVar(&r.keyStore, "keyStore", "", "Android: location of .keystore file containing signing information")
	flag.StringVar(&r.keyStorePass, "keyStorePass", "", "Android: password for the .keystore file, default take the password from stdin")
	flag.StringVar(&r.keyName, "keyName", "", "Android: alias for the signer's private key, which is needed when reading a .keystore file")
	flag.StringVar(&r.keyPass, "keyPass", "", "Android: password for the signer's private key, which is needed if the private key is password-protected. Default take the password from stdin")
	flag.StringVar(&r.certificate, "certificate", "", "iOS/macOS/Windows: name of the certificate to sign the build")
	flag.StringVar(&r.profile, "profile", "", "iOS/macOS: name of the provisioning profile for this release build")
	flag.StringVar(&r.developer, "developer", "", "Windows: the developer identity for your Microsoft store account")
	flag.StringVar(&r.password, "password", "", "Windows: password for the certificate used to sign the build")
	flag.StringVar(&r.tags, "tags", "", "A comma-separated list of build tags")
	flag.StringVar(&r.category, "category", "", "macOS: category of the app for store listing")
}

// PrintHelp prints the help message for the release command.
//
// Deprecated: Access to the individual cli commands are being removed.
func (r *Releaser) PrintHelp(indent string) {
	fmt.Println(indent, "The release command prepares an application for public distribution.")
}

// Run runs the release command.
//
// Deprecated: A better version will be exposed in the future.
func (r *Releaser) Run(params []string) {
	r.Packager.distribution = true
	r.Packager.release = true

	if err := r.validate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	if err := r.beforePackage(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}
	r.Packager.Run(params)
	if err := r.afterPackage(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func (r *Releaser) releaseAction(_ *cli.Context) error {
	r.Packager.distribution = true
	r.Packager.release = true

	if err := r.validate(); err != nil {
		return err
	}

	if err := r.beforePackage(); err != nil {
		return err
	}

	if err := r.Packager.packageWithoutValidate(); err != nil {
		return err
	}

	if err := r.afterPackage(); err != nil {
		return err
	}

	return nil
}

func (r *Releaser) afterPackage() error {
	if util.IsAndroid(r.os) {
		target := mobile.AppOutputName(r.os, r.Packager.Name, r.release)
		apk := filepath.Join(r.dir, target)
		if err := r.zipAlign(apk); err != nil {
			return err
		}
		return r.signAndroid(apk)
	}
	if r.os == "darwin" {
		return r.packageMacOSRelease()
	}
	if r.os == "ios" {
		return r.packageIOSRelease()
	}
	if r.os == "windows" {
		outName := r.Name + ".appx"
		if pos := strings.LastIndex(r.Name, ".exe"); pos > 0 {
			outName = r.Name[:pos] + ".appx"
		}

		outPath := filepath.Join(r.dir, outName)
		os.Remove(outPath) // MakeAppx will hang if the file exists... ignore result
		if err := r.packageWindowsRelease(outPath); err != nil {
			return err
		}
		return r.signWindows(outPath)
	}
	return nil
}

func (r *Releaser) beforePackage() error {
	if util.IsAndroid(r.os) {
		if err := util.RequireAndroidSDK(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Releaser) nameFromCertInfo(info string) string {
	// format should be "CN=Company, O=Company, L=City, S=State, C=Country"
	parts := strings.Split(info, ",")
	cn := parts[0]
	pos := strings.Index(strings.ToUpper(cn), "CN=")
	if pos == -1 {
		return cn // not what we were expecting, but should be OK
	}

	return cn[pos+3:]
}

func (r *Releaser) packageIOSRelease() error {
	team, err := mobile.DetectIOSTeamID(r.certificate)
	if err != nil {
		return errors.New("failed to determine team ID")
	}

	payload := filepath.Join(r.dir, "Payload")
	_ = os.Mkdir(payload, 0o750)
	defer os.RemoveAll(payload)
	appName := mobile.AppOutputName(r.os, r.Name, r.release)
	payloadAppDir := filepath.Join(payload, appName)
	if err := os.Rename(filepath.Join(r.dir, appName), payloadAppDir); err != nil {
		return err
	}

	cleanup, err := r.writeEntitlements(templates.EntitlementsDarwinMobile, struct{ TeamID, AppID string }{
		TeamID: team,
		AppID:  r.AppID,
	})
	if err != nil {
		return errors.New("failed to write entitlements plist template")
	}
	defer cleanup()

	cmd := exec.Command("codesign", "-f", "-vv", "-s", r.certificate, "--entitlements",
		"entitlements.plist", "Payload/"+appName+"/")
	if err := cmd.Run(); err != nil {
		fyne.LogError("Codesign failed", err)
		return errors.New("unable to codesign application bundle")
	}

	return exec.Command("zip", "-r", appName[:len(appName)-4]+".ipa", "Payload/").Run()
}

func (r *Releaser) packageMacOSRelease() error {
	// try to derive two certificates from one name (they will be consistent)
	appCert := strings.Replace(r.certificate, "Installer", "Application", 1)
	installCert := strings.Replace(r.certificate, "Application", "Installer", 1)
	unsignedPath := r.Name + "-unsigned.pkg"

	defer os.RemoveAll(r.Name + ".app") // this was the output of package and it can get in the way of future builds

	cleanup, err := r.writeEntitlements(templates.EntitlementsDarwin, nil)
	if err != nil {
		return errors.New("failed to write entitlements plist template")
	}
	defer cleanup()

	cmd := exec.Command("codesign", "-vfs", appCert, "--entitlement", "entitlements.plist", r.Name+".app")
	err = cmd.Run()
	if err != nil {
		fyne.LogError("Codesign failed", err)
		return errors.New("unable to codesign application bundle")
	}

	cmd = exec.Command("productbuild", "--component", r.Name+".app", "/Applications/",
		"--product", r.Name+".app/Contents/Info.plist", unsignedPath)
	err = cmd.Run()
	if err != nil {
		fyne.LogError("Product build failed", err)
		return errors.New("unable to build macOS app package")
	}
	defer os.Remove(unsignedPath)

	cmd = exec.Command("productsign", "--sign", installCert, unsignedPath, r.Name+".pkg")
	return cmd.Run()
}

func (r *Releaser) packageWindowsRelease(outFile string) error {
	payload := filepath.Join(r.dir, "Payload")
	_ = os.Mkdir(payload, 0o750)
	defer os.RemoveAll(payload)

	manifestPath := filepath.Join(payload, "appxmanifest.xml")
	manifest, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	manifestData := struct{ AppID, Developer, DeveloperName, Name, Version string }{
		AppID: r.AppID,
		// TODO read this info
		Developer:     encodeXMLString(r.developer),
		DeveloperName: encodeXMLString(r.nameFromCertInfo(r.developer)),
		Name:          encodeXMLString(r.Name),
		Version:       r.combinedVersion(),
	}
	err = templates.AppxManifestWindows.Execute(manifest, manifestData)
	manifest.Close()
	if err != nil {
		return errors.New("failed to write application manifest template")
	}

	util.CopyFile(r.icon, filepath.Join(payload, "Icon.png"))
	util.CopyFile(r.Name, filepath.Join(payload, r.Name))

	binDir, err := findWindowsSDKBin()
	if err != nil {
		return errors.New("cannot find makeappx.exe, make sure you have installed the Windows SDK")
	}

	cmd := exec.Command(filepath.Join(binDir, "makeappx.exe"), "pack", "/d", payload, "/p", outFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *Releaser) signAndroid(path string) error {
	signer := "jarsigner"
	var args []string
	if r.release {
		args = []string{"-keystore", r.keyStore}
	} else {
		signer = filepath.Join(util.AndroidBuildToolsPath(), "/apksigner")
		args = []string{"sign", "--ks", r.keyStore}
	}

	if r.keyStorePass != "" {
		if r.release {
			args = append(args, "-storepass", r.keyStorePass)
		} else {
			args = append(args, "--ks-pass", "pass:"+r.keyStorePass)
		}
	}
	if r.keyPass != "" {
		if r.release {
			args = append(args, "-keypass", r.keyPass)
		} else {
			args = append(args, "--key-pass", "pass:"+r.keyPass)
		}
	}
	args = append(args, path)
	if r.release {
		if r.keyName == "" { // Required to sign Google Play .aab
			return errors.New("missing required -keyName (alias) parameter")
		}
		args = append(args, r.keyName)
	}

	cmd := exec.Command(signer, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *Releaser) signWindows(appx string) error {
	binDir, err := findWindowsSDKBin()
	if err != nil {
		return errors.New("cannot find signtool.exe, make sure you have installed the Windows SDK")
	}

	cmd := exec.Command(filepath.Join(binDir, "signtool.exe"),
		"sign", "/a", "/v", "/fd", "SHA256", "/f", r.certificate, "/p", r.password, appx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *Releaser) validate() error {
	if r.os == "" {
		r.os = targetOS()
	}
	err := r.Packager.validate()
	if err != nil {
		return err
	}

	if util.IsMobile(r.os) || r.os == "windows" {
		if r.AppVersion == "" { // Here it is required, if provided then package validate will check format
			return errors.New("missing required --app-version parameter")
		}
		if r.AppBuild <= 0 {
			return errors.New("missing required --app-build parameter")
		}
	}
	if r.os == "windows" {
		if r.developer == "" {
			return errors.New("missing required --developer parameter for windows release,\n" +
				"use data from Partner Portal, format \"CN=Company, O=Company, L=City, S=State, C=Country\"")
		}
		if r.certificate == "" {
			return errors.New("missing required --certificate parameter for windows release")
		}
		if r.password == "" {
			return errors.New("missing required --password parameter for windows release")
		}
	}
	if util.IsAndroid(r.os) {
		if r.keyStore == "" {
			return errors.New("missing required --key-store parameter for android release")
		}
	} else if r.os == "darwin" {
		if r.certificate == "" {
			r.certificate = "3rd Party Mac Developer Application"
		}
		if r.profile == "" {
			return errors.New("missing required --profile parameter for macOS release")
		}
		if r.category == "" {
			return errors.New("missing required --category parameter for macOS release")
		} else if !isValidMacOSCategory(r.category) {
			return errors.New("category does not match one of the supported list: " +
				strings.Join(macAppStoreCategories, ", "))
		}
	} else if r.os == "ios" {
		if r.certificate == "" {
			r.certificate = "Apple Distribution"
		}
		if r.profile == "" {
			return errors.New("missing required --profile parameter for iOS release")
		}
	}
	return nil
}

func (r *Releaser) writeEntitlements(tmpl *template.Template, entitlementData any) (cleanup func(), err error) {
	entitlementPath := filepath.Join(r.dir, "entitlements.plist")
	entitlements, err := os.Create(entitlementPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := entitlements.Close(); r != nil && err == nil {
			err = r
		}
	}()

	if err := tmpl.Execute(entitlements, entitlementData); err != nil {
		return nil, err
	}
	return func() {
		_ = os.Remove(entitlementPath)
	}, nil
}

func (r *Releaser) zipAlign(path string) error {
	unaligned := filepath.Join(filepath.Dir(path), "unaligned.apk")
	err := os.Rename(path, unaligned)
	if err != nil {
		return nil
	}

	cmd := filepath.Join(util.AndroidBuildToolsPath(), "zipalign")
	err = exec.Command(cmd, "4", unaligned, path).Run()
	if err != nil {
		_ = os.Rename(path, unaligned) // ignore error, return previous
		return err
	}
	return os.Remove(unaligned)
}

func findWindowsSDKBin() (string, error) {
	inPath, err := exec.LookPath("makeappx.exe")
	if err == nil {
		return inPath, nil
	}

	matches, err := filepath.Glob("C:\\Program Files (x86)\\Windows Kits\\*\\bin\\*\\*\\makeappx.exe")
	if err != nil || len(matches) == 0 {
		return "", errors.New("failed to look up standard locations for makeappx.exe")
	}

	return filepath.Dir(matches[0]), nil
}

func isValidMacOSCategory(in string) bool {
	found := false
	for _, cat := range macAppStoreCategories {
		if cat == strings.ToLower(in) {
			found = true
			break
		}
	}
	return found
}
