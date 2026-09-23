module fyne.io/tools

go 1.22.0

require (
	fyne.io/fyne/v2 v2.8.1
	github.com/BurntSushi/toml v1.6.0
	github.com/fogleman/gg v1.3.0
	github.com/fyne-io/image v0.1.1
	github.com/go-ole/go-ole v1.3.0
	github.com/jackmordaunt/icns/v2 v2.2.7
	github.com/josephspurrier/goversioninfo v1.7.0
	github.com/lucor/goinfo v0.9.0
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/natefinch/atomic v1.0.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/stretchr/testify v1.11.1
	github.com/urfave/cli/v2 v2.27.7
	golang.org/x/mod v0.23.0
	golang.org/x/tools v0.30.0
	golang.org/x/tools/go/vcs v0.1.0-deprecated
)

require (
	github.com/akavel/rsrc v0.10.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fyne-io/gl-js v0.2.1-0.20260315212741-029c47fd27e8 // indirect
	github.com/fyne-io/oksvg v0.2.0 // indirect
	github.com/go-gl/gl v0.0.0-20260331235117-4566fea9a276 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jsummers/gobmp v0.0.0-20230614200233-a9de23ed2e25 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/image v0.24.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v1.26.1 // Contains only retraction as v1.6.1 was ignored
	v1.26.0 // Published accidentally.
)
