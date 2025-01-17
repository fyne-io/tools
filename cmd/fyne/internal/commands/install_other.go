//go:build !windows

package commands

func postInstall(_ *Installer) error {
	return nil
}
