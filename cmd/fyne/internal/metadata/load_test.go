package metadata

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAppMetadata(t *testing.T) {
	r, err := os.Open("./testdata/FyneApp.toml")
	assert.Nil(t, err)
	defer r.Close()

	data, err := Load(r)
	assert.Nil(t, err)
	assert.Equal(t, "https://apps.fyne.io", data.Website)
	assert.Equal(t, "io.fyne.fyne", data.Details.ID)
	assert.Equal(t, "1.0.0", data.Details.Version)
	assert.Equal(t, 1, data.Details.Build)
	assert.Equal(t, data.Release["Test"], "Value1")
	assert.Equal(t, data.Release["InReleaseOnly"], "Value3")
	assert.NotContains(t, data.Release, "InDevelopmentOnly")
	assert.Equal(t, data.Development["Test"], "Value2")
	assert.Equal(t, data.Development["InDevelopmentOnly"], "Value4")
	assert.NotContains(t, data.Development, "InReleaseOnly")

	assert.NotNil(t, data.Source)
	assert.Equal(t, data.Source.Repo, "https://github.com/fyne-io/fyne")
	assert.Equal(t, data.Source.Dir, "internal/metadata/testdata")
}

func TestLoadAppMetadata_Permissions(t *testing.T) {
	data, err := Load(strings.NewReader("[Details]\nName = \"Fyne App\"\n"))
	assert.Nil(t, err)
	assert.Nil(t, data.Permissions)

	data, err = Load(strings.NewReader("[Permissions]\nMicrophone = true\n"))
	assert.Nil(t, err)
	assert.NotNil(t, data.Permissions)
	assert.True(t, data.Permissions.Microphone)
	assert.Equal(t, "Fyne App uses the microphone to capture audio", data.Permissions.MicrophoneUsageText("Fyne App"))

	data, err = Load(strings.NewReader("[Permissions]\nMicrophone = true\nMicrophoneUsage = \"Record voice notes\"\n"))
	assert.Nil(t, err)
	assert.Equal(t, "Record voice notes", data.Permissions.MicrophoneUsageText("Fyne App"))
}
