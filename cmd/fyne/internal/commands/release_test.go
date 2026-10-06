package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/tools/cmd/fyne/internal/util"
)

func TestReleaser_nameFromCertInfo(t *testing.T) {
	rel := &Releaser{}
	cert := "CN=Company, O=Company, L=City, S=State, C=Country"
	assert.Equal(t, "Company", rel.nameFromCertInfo(cert))
	assert.Equal(t, "Fallback", rel.nameFromCertInfo("Fallback"))
	assert.Equal(t, "Fallback", rel.nameFromCertInfo("Fallback, extra"))

	badCase := "Cn=Company, O=Company, L=City, S=State, C=Country"
	assert.Equal(t, "Company", rel.nameFromCertInfo(badCase))
}

func TestIsValidMacOSCategory(t *testing.T) {
	assert.True(t, isValidMacOSCategory("games"))
	assert.True(t, isValidMacOSCategory("utilities"))
	assert.True(t, isValidMacOSCategory("Games"))

	assert.False(t, isValidMacOSCategory("sporps"))
	assert.False(t, isValidMacOSCategory("android-games"))
	assert.False(t, isValidMacOSCategory(""))
}

func Test_ReleaserZipAlignVerbose(t *testing.T) {
	dir := t.TempDir()
	apk := filepath.Join(dir, "myapp.apk")
	require.NoError(t, os.WriteFile(apk, []byte("data"), util.FilePermDefault))

	useMockPkgUtil(t)
	utilAndroidBuildToolsPathMock = func() string { return dir } // no zipalign in there

	r := &Releaser{
		Packager: Packager{
			appData: &appData{},
		},
	}
	r.verbose = true

	out := captureStdout(t, func() {
		err := r.zipAlign(apk)
		assert.Error(t, err)
	})

	assert.Contains(t, out, "Aligning "+apk)
}

func Test_ReleaserSignAndroidVerbose(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // no jarsigner found, so the signing fails

	r := &Releaser{
		Packager: Packager{
			appData: &appData{},
			release: true,
		},
		keyName: "alias",
	}
	r.verbose = true

	out := captureStdout(t, func() {
		err := r.signAndroid("myapp.apk")
		assert.Error(t, err)
	})

	assert.Contains(t, out, "Signing ./myapp.apk")
}

func Test_ReleaserPackageWindowsVerbose(t *testing.T) {
	dir := t.TempDir()

	useMockPkgUtil(t)
	utilCopyFileMock = func(source string, target string) error { return nil }

	r := &Releaser{
		Packager: Packager{
			appData: &appData{Name: "myapp"},
		},
	}
	r.dir = dir
	r.verbose = true

	appx := filepath.Join(dir, "myapp.appx")
	out := captureStdout(t, func() {
		err := r.packageWindowsRelease(appx)
		assert.Error(t, err) // the Windows SDK tools cannot be found
	})

	assert.Contains(t, out, "Packaging "+appx)
}
