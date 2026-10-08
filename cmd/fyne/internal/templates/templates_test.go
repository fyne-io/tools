package templates

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLaunchScreenIOS(t *testing.T) {
	buf := &bytes.Buffer{}
	err := LaunchScreenIOS.Execute(buf, struct {
		Red, Green, Blue float64
		IconSize         int
	}{Red: 0.5, Green: 0.25, Blue: 1, IconSize: 288})
	assert.Nil(t, err)

	out := buf.String()
	assert.Contains(t, out, `red="0.5" green="0.25" blue="1"`)
	assert.Contains(t, out, `image="SplashIcon"`)
	assert.Contains(t, out, `firstAttribute="width" constant="288"`)
	assert.Contains(t, out, `firstAttribute="height" constant="288"`)
	assert.Contains(t, out, `<image name="SplashIcon" width="288" height="288"/>`)
}
