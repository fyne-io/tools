package mobile

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// embeddedDebugKey parses the package debug key that the signing code reuses.
func embeddedDebugKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	block, _ := pem.Decode([]byte(debugCert))
	if block == nil {
		t.Fatal("debugCert did not decode as PEM")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parsing embedded debug key: %v", err)
	}
	return priv
}

// argValue returns the value following flag in args, or "" if absent.
func argValue(args []string, flag string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

func assertArgPair(t *testing.T, args []string, flag, value string) {
	t.Helper()
	if got := argValue(args, flag); got != value {
		t.Fatalf("arg %s = %q, want %q; full args: %v", flag, got, value, args)
	}
}

func assertContains(t *testing.T, args []string, want string) {
	t.Helper()
	for _, a := range args {
		if a == want {
			return
		}
	}
	t.Fatalf("args %v do not contain %q", args, want)
}

// stubBuildTool creates a build-tools dir with an entry for the named tool and points
// ANDROID_HOME at it, so the existence check in zipalignAPK/signAPKModern passes. The entry
// is never executed, since tests substitute runBuildTool.
func stubBuildTool(t *testing.T, tool string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if tool == "apksigner" {
			tool += ".bat"
		} else {
			tool += ".exe"
		}
	}
	home := t.TempDir()
	bt := filepath.Join(home, "build-tools")
	if err := os.MkdirAll(bt, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bt, tool), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ANDROID_HOME", home)
}

// TestDebugCertificateDeterministic guards the stable signing identity: two builds must
// produce byte-identical certificates, else an updated debug APK signs as a different signer.
func TestDebugCertificateDeterministic(t *testing.T) {
	priv := embeddedDebugKey(t)
	a, err := debugCertificate(priv)
	if err != nil {
		t.Fatalf("debugCertificate: %v", err)
	}
	b, err := debugCertificate(priv)
	if err != nil {
		t.Fatalf("debugCertificate: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("debugCertificate is not deterministic; an updated debug APK would install as a different signer")
	}
}

// TestDebugCertificateContents checks the generated certificate is a valid self-signed
// certificate carrying the embedded debug key.
func TestDebugCertificateContents(t *testing.T) {
	priv := embeddedDebugKey(t)
	der, err := debugCertificate(priv)
	if err != nil {
		t.Fatalf("debugCertificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	if err := cert.CheckSignatureFrom(cert); err != nil {
		t.Fatalf("certificate is not validly self-signed: %v", err)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("certificate public key is %T, want *rsa.PublicKey", cert.PublicKey)
	}
	if pub.N.Cmp(priv.N) != 0 || pub.E != priv.E {
		t.Fatal("certificate public key does not match the embedded debug key")
	}
	if cert.Subject.CommonName != "Fyne Debug" {
		t.Fatalf("subject CN = %q, want %q", cert.Subject.CommonName, "Fyne Debug")
	}
	if !cert.IsCA {
		t.Fatal("certificate is not marked as a CA")
	}
	wantNotBefore := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !cert.NotBefore.Equal(wantNotBefore) {
		t.Fatalf("NotBefore = %v, want %v", cert.NotBefore, wantNotBefore)
	}
}

// TestWriteDebugSigningFiles checks the temporary key and certificate handed to apksigner
// are well-formed, match the embedded debug key, and are removed by the cleanup func.
func TestWriteDebugSigningFiles(t *testing.T) {
	keyPath, certPath, cleanup, err := writeDebugSigningFiles()
	if err != nil {
		t.Fatalf("writeDebugSigningFiles: %v", err)
	}
	want := embeddedDebugKey(t)

	keyDER, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}
	parsedAny, err := x509.ParsePKCS8PrivateKey(keyDER)
	if err != nil {
		t.Fatalf("key file is not valid PKCS#8: %v", err)
	}
	parsed, ok := parsedAny.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("key is %T, want *rsa.PrivateKey", parsedAny)
	}
	if parsed.N.Cmp(want.N) != 0 || parsed.E != want.E {
		t.Fatal("written key does not match the embedded debug key")
	}

	certDER, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read cert: %v", err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("cert file is not valid DER: %v", err)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok || pub.N.Cmp(want.N) != 0 {
		t.Fatal("written cert does not carry the embedded debug key")
	}

	cleanup()
	if _, err := os.Stat(filepath.Dir(keyPath)); !os.IsNotExist(err) {
		t.Fatalf("cleanup did not remove temp dir, stat err = %v", err)
	}
}

// TestZipalignAPKInvokesZipalign checks zipalignAPK aligns the APK via zipalign and
// replaces the original with the aligned output in place.
func TestZipalignAPKInvokesZipalign(t *testing.T) {
	stubBuildTool(t, "zipalign")

	var gotArgs []string
	orig := runBuildTool
	t.Cleanup(func() { runBuildTool = orig })
	runBuildTool = func(name string, args ...string) error {
		gotArgs = append([]string{name}, args...)
		// zipalign writes the aligned archive to the last argument.
		return os.WriteFile(args[len(args)-1], []byte("aligned"), 0o600)
	}

	apk := filepath.Join(t.TempDir(), "app.apk")
	if err := os.WriteFile(apk, []byte("unaligned"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := zipalignAPK(apk); err != nil {
		t.Fatalf("zipalignAPK: %v", err)
	}

	base := filepath.Base(gotArgs[0])
	if base != "zipalign" && base != "zipalign.exe" {
		t.Fatalf("expected zipalign to be invoked, got %q; args: %v", gotArgs[0], gotArgs)
	}
	assertContains(t, gotArgs, "-P") // 16 KB page-align uncompressed native libraries
	assertContains(t, gotArgs, "16")
	assertContains(t, gotArgs, "4") // 4-byte alignment required for resources.arsc
	if got, err := os.ReadFile(apk); err != nil || string(got) != "aligned" {
		t.Fatalf("apk not replaced with aligned output (content=%q, err=%v)", got, err)
	}
	if _, err := os.Stat(apk + ".aligned"); !os.IsNotExist(err) {
		t.Fatalf(".aligned intermediate not renamed away, stat err = %v", err)
	}
}

// TestSignAPKModernInvokesApksigner checks signAPKModern enables the v1/v2/v3/v4 schemes,
// signs the APK in place with the generated key and certificate, and keeps the .idsig
// sidecar (needed for adb's incremental install path).
func TestSignAPKModernInvokesApksigner(t *testing.T) {
	stubBuildTool(t, "apksigner")

	var gotArgs []string
	orig := runBuildTool
	t.Cleanup(func() { runBuildTool = orig })
	runBuildTool = func(name string, args ...string) error {
		gotArgs = append([]string{name}, args...)

		// The key and cert handed to apksigner must be well-formed.
		if b, err := os.ReadFile(argValue(args, "--key")); err != nil {
			t.Errorf("key file unreadable: %v", err)
		} else if _, err := x509.ParsePKCS8PrivateKey(b); err != nil {
			t.Errorf("key file is not PKCS#8: %v", err)
		}
		if b, err := os.ReadFile(argValue(args, "--cert")); err != nil {
			t.Errorf("cert file unreadable: %v", err)
		} else if _, err := x509.ParseCertificate(b); err != nil {
			t.Errorf("cert file is not a DER certificate: %v", err)
		}

		// apksigner writes an <apk>.idsig sidecar next to the signed APK.
		apk := args[len(args)-1]
		return os.WriteFile(apk+".idsig", []byte("idsig"), 0o600)
	}

	apk := filepath.Join(t.TempDir(), "app.apk")
	if err := os.WriteFile(apk, []byte("PK\x03\x04"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := signAPKModern(apk); err != nil {
		t.Fatalf("signAPKModern: %v", err)
	}

	if len(gotArgs) == 0 || gotArgs[1] != "sign" {
		t.Fatalf("apksigner not invoked with 'sign'; args: %v", gotArgs)
	}
	assertArgPair(t, gotArgs, "--v1-signing-enabled", "true")
	assertArgPair(t, gotArgs, "--v2-signing-enabled", "true")
	assertArgPair(t, gotArgs, "--v3-signing-enabled", "true")
	assertArgPair(t, gotArgs, "--v4-signing-enabled", "true")
	if gotArgs[len(gotArgs)-1] != apk {
		t.Fatalf("last arg = %q, want apk path %q", gotArgs[len(gotArgs)-1], apk)
	}
	// The .idsig sidecar must be kept so adb uses the incremental install path.
	if _, err := os.Stat(apk + ".idsig"); err != nil {
		t.Fatalf(".idsig sidecar should be kept next to the APK, stat err = %v", err)
	}
}

// TestSignAPKModernApksignerMissing checks a clear error is returned (and nothing is run)
// when apksigner cannot be located.
func TestSignAPKModernApksignerMissing(t *testing.T) {
	t.Setenv("ANDROID_HOME", t.TempDir()) // no build-tools/apksigner inside

	called := false
	orig := runBuildTool
	t.Cleanup(func() { runBuildTool = orig })
	runBuildTool = func(string, ...string) error {
		called = true
		return nil
	}

	err := signAPKModern(filepath.Join(t.TempDir(), "app.apk"))
	if err == nil {
		t.Fatal("expected an error when apksigner is missing")
	}
	if !strings.Contains(err.Error(), "apksigner") {
		t.Fatalf("error %q does not mention apksigner", err)
	}
	if called {
		t.Fatal("signCommand was invoked despite apksigner being missing")
	}
}
