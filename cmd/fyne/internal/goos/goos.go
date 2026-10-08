package goos

import (
	"strings"
)

// OS constants (to compare to runtime.GOOS)
const (
	Android    = "android"
	Darwin     = "darwin"
	FreeBSD    = "freebsd"
	IOS        = "ios"
	JavaScript = "js"
	Linux      = "linux"
	NetBSD     = "netbsd"
	OpenBSD    = "openbsd"
	WASM       = "wasm"
	Web        = "web"
	Windows    = "windows"
)

// IsBSD returns whether the specified OS is a supported BSD variant.
func IsBSD(os string) bool {
	return os == FreeBSD || os == NetBSD || os == OpenBSD
}

// IsIOS returns true if the given os parameter represents one of the iOS targets (ios, iossimulator)
func IsIOS(os string) bool {
	return strings.HasPrefix(os, "ios")
}

// IsMobile returns true if the given os parameter represents a platform handled by gomobile.
func IsMobile(os string) bool {
	return IsIOS(os) || IsAndroid(os)
}

// IsAndroid returns true if the given os parameter represents one of the Android targets.
func IsAndroid(os string) bool {
	return strings.HasPrefix(os, "android")
}

// IsWASM returns true if the given os parameter represents one of the web targets.
func IsWASM(os string) bool {
	return os == WASM || os == Web
}

// IsWeb returns true if the given os parameter represents one of the web targets.
func IsWeb(os string) bool {
	return os == JavaScript || os == WASM || os == Web
}
