package mobile

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"fyne.io/tools/cmd/fyne/internal/metadata"
)

func writeTestPNG(t *testing.T, path string, size int) {
	f, err := os.Create(path)
	assert.Nil(t, err)
	assert.Nil(t, png.Encode(f, image.NewRGBA(image.Rect(0, 0, size, size))))
	assert.Nil(t, f.Close())
}

func pngSize(t *testing.T, path string) int {
	f, err := os.Open(path)
	assert.Nil(t, err)
	defer f.Close()
	conf, err := png.DecodeConfig(f)
	assert.Nil(t, err)
	assert.Equal(t, conf.Width, conf.Height)
	return conf.Width
}

func Test_writeSplashResources(t *testing.T) {
	dir := t.TempDir()
	icon := filepath.Join(dir, "icon.png")
	custom := filepath.Join(dir, "custom.png")
	writeTestPNG(t, icon, 64)
	writeTestPNG(t, custom, 32)

	res := filepath.Join(dir, "res")
	err := writeSplashResources(res, &metadata.Splash{Background: "#015952"}, icon)
	assert.Nil(t, err)

	// the fallback icon is scaled to the fixed splash size
	assert.Equal(t, splashIconSize, pngSize(t, filepath.Join(res, "drawable-xxxhdpi", "splash_icon.png")))

	data, err := os.ReadFile(filepath.Join(res, "values", "colors.xml"))
	assert.Nil(t, err)
	assert.Contains(t, string(data), `<color name="splash_background">#015952</color>`)
	data, err = os.ReadFile(filepath.Join(res, "drawable", "splash_background.xml"))
	assert.Nil(t, err)
	assert.Contains(t, string(data), "@color/splash_background")
	assert.Contains(t, string(data), "@drawable/splash_icon")
	data, err = os.ReadFile(filepath.Join(res, "values", "styles.xml"))
	assert.Nil(t, err)
	assert.Contains(t, string(data), "android:windowBackground")
	data, err = os.ReadFile(filepath.Join(res, "values-v31", "styles.xml"))
	assert.Nil(t, err)
	assert.Contains(t, string(data), "android:windowSplashScreenAnimatedIcon")

	// a custom icon is preferred and the colour defaults to white
	res = filepath.Join(dir, "res2")
	assert.Nil(t, writeSplashResources(res, &metadata.Splash{Icon: custom}, icon))
	assert.Equal(t, splashIconSize, pngSize(t, filepath.Join(res, "drawable-xxxhdpi", "splash_icon.png")))
	data, err = os.ReadFile(filepath.Join(res, "values", "colors.xml"))
	assert.Nil(t, err)
	assert.Contains(t, string(data), "#FFFFFF")

	// a missing icon is an error
	assert.NotNil(t, writeSplashResources(filepath.Join(dir, "res3"), &metadata.Splash{}, filepath.Join(dir, "none.png")))
}
