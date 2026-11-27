package mobile

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/tools/cmd/fyne/internal/util"
)

// runBuildTool runs an Android SDK build-tool (zipalign, apksigner). It is a package var so
// tests can substitute it and assert its arguments.
var runBuildTool = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// zipalignAPK aligns the APK's ZIP entries with zipalign so resources.arsc is uncompressed
// on a 4-byte boundary, as targetSdkVersion >= 30 requires. Must run before signAPKModern,
// as re-aligning would invalidate the signature.
func zipalignAPK(apkPath string) error {
	zipalign := filepath.Join(util.AndroidBuildToolsPath(), "zipalign")
	if runtime.GOOS == "windows" {
		zipalign += ".exe"
	}
	if !util.Exists(zipalign) {
		return fmt.Errorf("zipalign not found in Android SDK build-tools; it is needed to align the APK so it installs on targetSdkVersion >= 30")
	}

	// zipalign cannot align in place, so write an aligned copy and swap it in. -P 16 page-
	// aligns native libraries for 16 KB devices; 4 is the byte alignment resources.arsc needs.
	aligned := apkPath + ".aligned"
	if err := runBuildTool(zipalign, "-P", "16", "-f", "4", apkPath, aligned); err != nil {
		return fmt.Errorf("zipalign failed for %s: %w", apkPath, err)
	}
	return os.Rename(aligned, apkPath)
}

// signAPKModern adds APK Signature Scheme v2/v3/v4 signatures with apksigner, on top of the
// v1 (JAR) signature that targetSdkVersion >= 30 no longer accepts alone. It reuses the
// embedded debug key and a cert derived from it (see debugCertificate), so the debug signing
// identity stays stable across builds.
func signAPKModern(apkPath string) error {
	apksigner := filepath.Join(util.AndroidBuildToolsPath(), "apksigner")
	if runtime.GOOS == "windows" {
		apksigner += ".bat"
	}
	if !util.Exists(apksigner) {
		return fmt.Errorf("apksigner not found in Android SDK build-tools")
	}

	keyPath, certPath, cleanup, err := writeDebugSigningFiles()
	if err != nil {
		return err
	}
	defer cleanup()

	// v1 keeps pre-API-24 devices working; v2/v3 satisfy the modern requirement; v4 writes an
	// <apk>.idsig sidecar so `adb install` uses the incremental path (some 16 KB emulators
	// reject the streamed path).
	err = runBuildTool(apksigner, "sign",
		"--key", keyPath,
		"--cert", certPath,
		"--v1-signing-enabled", "true",
		"--v2-signing-enabled", "true",
		"--v3-signing-enabled", "true",
		"--v4-signing-enabled", "true",
		apkPath,
	)
	if err != nil {
		return fmt.Errorf("apksigner failed to sign %s: %w", apkPath, err)
	}
	return nil
}

// writeDebugSigningFiles writes the embedded debug key (PKCS#8 DER) and a self-signed cert
// for it (X.509 DER) to temp files for apksigner, returning their paths and a cleanup func.
func writeDebugSigningFiles() (keyPath, certPath string, cleanup func(), err error) {
	block, _ := pem.Decode([]byte(debugCert))
	if block == nil {
		return "", "", nil, errors.New("no debug cert")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", "", nil, err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", "", nil, err
	}
	certDER, err := debugCertificate(priv)
	if err != nil {
		return "", "", nil, err
	}

	dir, err := os.MkdirTemp("", "fyne-apksign")
	if err != nil {
		return "", "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }

	keyPath = filepath.Join(dir, "debug.pk8")
	if err := os.WriteFile(keyPath, keyDER, 0o600); err != nil {
		cleanup()
		return "", "", nil, err
	}
	certPath = filepath.Join(dir, "debug.x509.der")
	if err := os.WriteFile(certPath, certDER, 0o600); err != nil {
		cleanup()
		return "", "", nil, err
	}
	return keyPath, certPath, cleanup, nil
}

// debugCertificate builds a self-signed X.509 cert (DER) for the debug key. Its template is
// fixed and RSA PKCS#1 v1.5 signatures are deterministic, so the bytes are identical every
// build; that stable identity lets an updated debug APK install over a previous one.
func debugCertificate(priv *rsa.PrivateKey) ([]byte, error) {
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Fyne Debug", Organization: []string{"Fyne"}, Country: []string{"US"}},
		NotBefore:             time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:              time.Date(2125, 1, 1, 0, 0, 0, 0, time.UTC),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	return x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
}
