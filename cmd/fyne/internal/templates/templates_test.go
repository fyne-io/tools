package templates

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestMicrophoneCapability(t *testing.T) {
	tests := map[string]struct {
		tmpl *template.Template
		want string
	}{
		"darwin info":         {InfoPlistDarwin, "NSMicrophoneUsageDescription"},
		"darwin entitlements": {EntitlementsDarwin, "com.apple.security.device.audio-input"},
		"android manifest":    {ManifestAndroid, "android.permission.RECORD_AUDIO"},
		"windows appx":        {AppxManifestWindows, `<DeviceCapability Name="microphone" />`},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for _, enabled := range []bool{false, true} {
				buf := &bytes.Buffer{}
				err := tt.tmpl.Execute(buf, map[string]any{"Microphone": enabled, "MicrophoneUsage": "Notes & memos"})
				assert.NoError(t, err)
				assert.Equal(t, enabled, bytes.Contains(buf.Bytes(), []byte(tt.want)))
			}
		})
	}

	buf := &bytes.Buffer{}
	err := InfoPlistDarwin.Execute(buf, map[string]any{"Microphone": true, "MicrophoneUsage": "Notes & memos"})
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "<string>Notes &amp; memos</string>")
}

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
