// Package cross provides functionality for cross-compiling Fyne applications
// using containerized build environments. It manages container configuration,
// go build command construction, and execution of builds inside Docker/Moby containers
// with support for different target OS/architecture combinations.
package cross

import "fmt"

// osImages maps operating systems to their container images
var osImages = map[string]string{
	"linux":   "fyneio/fyne-cross-images:tools-linux",
	"windows": "fyneio/fyne-cross-images:tools-windows",
	"darwin":  "fyneio/fyne-cross-images:tools-darwin",
}

// targetEnvMap defines OS/arch-specific environment variables for cross-compilation
var targetEnvMap = map[string]map[string]string{
	// Linux targets
	"linux/amd64": {
		"CC":  "zig cc -target x86_64-linux-gnu -isystem /usr/include -L/usr/lib/x86_64-linux-gnu",
		"CXX": "zig c++ -target x86_64-linux-gnu -isystem /usr/include -L/usr/lib/x86_64-linux-gnu",
	},
	"linux/386": {
		"CC":  "zig cc -target x86-linux-gnu -isystem /usr/include -L/usr/lib/i386-linux-gnu",
		"CXX": "zig c++ -target x86-linux-gnu -isystem /usr/include -L/usr/lib/i386-linux-gnu",
	},
	"linux/arm": {
		"CC":    "zig cc -target arm-linux-gnueabihf -isystem /usr/include -L/usr/lib/arm-linux-gnueabihf",
		"CXX":   "zig c++ -target arm-linux-gnueabihf -isystem /usr/include -L/usr/lib/arm-linux-gnueabihf",
		"GOARM": "7",
	},
	"linux/arm64": {
		"CC":  "zig cc -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu",
		"CXX": "zig c++ -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu",
	},
	// Windows targets
	"windows/amd64": {
		"CC":  "zig cc -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
		"CXX": "zig c++ -target x86_64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
	},
	"windows/386": {
		"CC":  "zig cc -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
		"CXX": "zig c++ -target x86-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
	},
	"windows/arm64": {
		"CC":  "zig cc -target aarch64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
		"CXX": "zig c++ -target aarch64-windows-gnu -Wdeprecated-non-prototype -Wl,--subsystem,windows",
	},
}

// TargetImage returns the container image name for a given operating system.
// Returns an error if the OS is not supported.
func TargetImage(goos string) (string, error) {
	image, ok := osImages[goos]
	if !ok {
		return "", fmt.Errorf("unsupported OS: %s", goos)
	}
	return image, nil
}

// TargetEnv returns the environment variables for cross-compiling to a given OS/arch combination.
// The returned map includes CGO compiler settings, GOOS, GOARCH, and any arch-specific variables.
// Returns an error if the target combination is not supported.
func TargetEnv(goos, goarch string) (map[string]string, error) {
	target := goos + "/" + goarch
	env, ok := targetEnvMap[target]
	if !ok {
		return nil, fmt.Errorf("unsupported target: %s", target)
	}

	env["CGO_ENABLED"] = "1"
	env["GOOS"] = goos
	env["GOARCH"] = goarch
	return env, nil
}
