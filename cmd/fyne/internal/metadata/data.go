package metadata

// FyneApp describes the top level metadata for building a fyne application
type FyneApp struct {
	Website      string `toml:",omitempty"`
	Description  string `toml:",omitempty"`
	Details      AppDetails
	AdaptiveIcon *AdaptiveIcon `toml:",omitempty"`

	Development map[string]string `toml:",omitempty"`
	Release     map[string]string `toml:",omitempty"`
	Source      *AppSource        `toml:",omitempty"`
	CanOpen     *CanOpen          `toml:",omitempty"`
	LinuxAndBSD *LinuxAndBSD      `toml:",omitempty"`
	Languages   []string          `toml:",omitempty"`
	Migrations  map[string]bool   `toml:",omitempty"`
	Permissions *Permissions      `toml:",omitempty"`
}

// AppDetails describes the build information, this group may be OS or arch specific
type AppDetails struct {
	Icon     string `toml:",omitempty"`
	Name, ID string `toml:",omitempty"`
	Version  string `toml:",omitempty"`
	Build    int    `toml:",omitempty"`
}

type AdaptiveIcon struct {
	Foreground string `toml:",omitempty"`
	Background string `toml:",omitempty"`
	Monochrome string `toml:",omitempty"`
}

type AppSource struct {
	Repo, Dir string `toml:",omitempty"`
}

// LinuxAndBSD describes specific metadata for desktop files on Linux and BSD.
type LinuxAndBSD struct {
	GenericName string   `toml:",omitempty"`
	Categories  []string `toml:",omitempty"`
	Comment     string   `toml:",omitempty"`
	Keywords    []string `toml:",omitempty"`
	ExecParams  string   `toml:",omitempty"`
}

// Permissions lists the system integrations an application uses that the OS requires to be declared at package time.
type Permissions struct {
	Microphone bool `toml:",omitempty"`
	// MicrophoneUsage optionally overrides the default explanation shown when the OS asks the user for access.
	MicrophoneUsage string `toml:",omitempty"`
}

// MicrophoneUsageText returns the explanation of why the named app uses the microphone,
// this is the default text unless MicrophoneUsage was set.
func (p Permissions) MicrophoneUsageText(appName string) string {
	if p.MicrophoneUsage != "" {
		return p.MicrophoneUsage
	}
	return appName + " uses the microphone to capture audio"
}

// CanOpen represents a selection of file types (mime etc) that this application can open.
type CanOpen struct {
	MimeTypes string `toml:",omitempty"`
}
