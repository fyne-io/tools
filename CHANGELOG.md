# Changelog

This file lists the main changes with each version of the Fyne tools project.
More detailed release notes can be found on the [releases page](https://github.com/fyne-io/tools/releases).

## 1.7.0 - 16 Oct 2025

### Added

* Support Description metadata
* Allow more than one bundle argument
* Full support for semver 2.0 spec
* Translation setup for newly generated apps

### Changed

* Use embed instead of generated []byte for bundle command
* Make library version output deterministic and more forgiving
* Better error messages when trying to install remote apps without FyneApp.toml
* Replace dashes with underscores in generated app ids

### Fixed

* Remove legacy usage and files
* toml file should not be indented
* Correct lookup path for tools repo alongside fyne
* Fix up some keyboard issues on Android (fyne-io/fyne#5806)
* Entry with mobile.NumberKeyboard does not Type comma and separators (fyne-io/fyne#5101)
* Support d8 instead of dx for generating dex
* Make sure translation files end with a newline to prevent warnings from git and other tools
* Fix missing Migrations in compiled metadata
* Support installing Fyne apps in subdirectories

### New Contributors

Code in v1.7.0 contains work from the following first time contributors:

* @ErikKalkoken


## 1.6.2 - 22 August 2025

### Fixed

* Resolve compile issue with Go 1.25.0 caused by golang.org/x/tools conflict


## 1.6.1 - 15 April 2025

### Changed

 * New apps from "fyne init" will have the fyneDo migration turned on by default


## 1.6.0 - 11 April 2025

This is the beginning of the Fyne tools releases having migrated the `fyne` command line
tool from the [fyne-io/fyne](/fyne-io/fyne) repository.

### Added

 * Added new "fyne init" command to set up new apps
 * Update to support all new features of Fyne v2.6.0
 * Show the Fyne library version number next to the tool info in "fyne version"

