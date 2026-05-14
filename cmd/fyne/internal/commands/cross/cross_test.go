package cross

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTargetImage(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		want    string
		wantErr bool
	}{
		{
			name:    "linux target",
			goos:    "linux",
			want:    "fyneio/fyne-cross-images:linux",
			wantErr: false,
		},
		{
			name:    "darwin target",
			goos:    "darwin",
			want:    "fyneio/fyne-cross-images:darwin",
			wantErr: false,
		},
		{
			name:    "windows target",
			goos:    "windows",
			want:    "fyneio/fyne-cross-images:windows",
			wantErr: false,
		},
		{
			name:    "unsupported target",
			goos:    "ios",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty target",
			goos:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TargetImage(tt.goos)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "", got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestTargetEnv(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		goarch    string
		checkEnv  []string // env vars that should be present
		wantErr   bool
	}{
		{
			name:     "linux amd64",
			goos:     "linux",
			goarch:   "amd64",
			checkEnv: []string{"CC", "CXX", "CGO_ENABLED", "GOOS", "GOARCH"},
			wantErr:  false,
		},
		{
			name:     "linux arm64",
			goos:     "linux",
			goarch:   "arm64",
			checkEnv: []string{"CC", "CXX", "CGO_ENABLED", "GOOS", "GOARCH"},
			wantErr:  false,
		},
		{
			name:     "windows amd64",
			goos:     "windows",
			goarch:   "amd64",
			checkEnv: []string{"CC", "CXX", "CGO_ENABLED", "GOOS", "GOARCH"},
			wantErr:  false,
		},
		{
			name:     "linux arm with GOARM",
			goos:     "linux",
			goarch:   "arm",
			checkEnv: []string{"CC", "CXX", "GOARM", "CGO_ENABLED", "GOOS", "GOARCH"},
			wantErr:  false,
		},
		{
			name:     "unsupported target",
			goos:     "darwin",
			goarch:   "arm64",
			checkEnv: []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := TargetEnv(tt.goos, tt.goarch)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, env)
			} else {
				require.NoError(t, err)
				require.NotNil(t, env)

				// Check that required env vars are present
				for _, varName := range tt.checkEnv {
					assert.Contains(t, env, varName, "env var %s not found", varName)
				}

				// Verify standard vars are set correctly
				assert.Equal(t, tt.goos, env["GOOS"])
				assert.Equal(t, tt.goarch, env["GOARCH"])
				assert.Equal(t, "1", env["CGO_ENABLED"])
			}
		})
	}
}
