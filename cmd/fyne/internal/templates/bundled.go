package templates

import _ "embed"

//go:embed data/AndroidManifest.xml
var androidManifestXML string

//go:embed data/Info.plist
var infoPlist string

//go:embed data/Makefile
var makefile string

//go:embed data/entry.desktop
var desktopEntry string

//go:embed data/app.manifest
var appManifest string

//go:embed data/appxmanifest.XML
var appxManifestXML string

//go:embed data/entitlements-darwin.plist
var entitlementsDarwinPlist string

//go:embed data/entitlements-ios.plist
var entitlementsIosPlist string

//go:embed data/hello_world.got
var helloWorld string

//go:embed data/fyne_metadata_init.got
var fyneMetadataInit string

//go:embed data/fyne_pprof.got
var fynePprof string

//go:embed data/xcassets.JSON
var xcassetsJSON string

//go:embed data/index.html
var indexHTML string

//go:embed data/dark.css
var darkCSS []byte

//go:embed data/light.css
var lightCSS []byte

//go:embed data/spinner_dark.gif
var spinnerDarkGIF []byte

//go:embed data/spinner_light.gif
var spinnerLightGIF []byte

//go:embed data/webgl-debug.js
var webGLDebug []byte
