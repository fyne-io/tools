# Changelog

This file lists the main changes with each version of the Fyne tools project.
More detailed release notes can be found on the [releases page](https://github.com/fyne-io/tools/releases).

## 1.7.3 - 30 Sep 2026

### Added

* Compiler/platform based optimisation flags (#144, #147, #152)
* Fixing store compatibility for September (#151)

### Changed

* Update Fyne version, sync Go version, and work around macOS/CI issue (#157)
* Make sure version argument and command do the same (#158)
* gendex: use Fyne module source, CLI improvements, dependency download and checksum support (#166)

### Fixed

* Fix 'fyne package' embedding of Windows metadata and icon when outputting to a different directory (#134)
* Correct the icon name for Makefiles (#140)
* Ensure we list the most recent versions last to resolve correctly (#145)
* Don't rely on an empty AppID, it's optional (#146)

### New Contributors

Code in v1.7.3 contains work from the following first time contributors:

* @GraysonTinker


## 1.7.2 - 29 May 2026

### Changed

* Compatibility mode for fyne-cross-images (#130)

### Fixed

* Pass flag to compiler instead of preprocessor directly (#129)
* Build with compilers missing support flow-control protection (#133)


## 1.7.1 - 12 Apr 2026

### Added

* Start support for open with in metadata (#104)
* Support for adaptive and themed android icons (#115)
* Parsing of concatenated strings with fyne translate (#117)

### Changed

* Apply hardening C/LD-FLAGS and optimize (#100)
* Improve help to show that at least one file is needed (#102)
* Set StartupWMClass in Unix/Linux .desktop files (#112)
* gendex: Sort directory names in findLast (#123)

### Fixed

* Update Fyne version (#97)
* Don't generate Migration code for pre-2.6 library usage (#98)
* Support remote install with icons that are simply Icon.png (#99)
* Fix release flags not being set for WASM (#101)
* Don't make tarbombs (#105)
* Fallback to app name if AppID is not present for package_unix (#114)
* Update Android to API 21 in to keep the builds for Play Store happy (#119)
* Make sure that CGO is not enabled when building in WASM mode (#121)
* fyne serve: fix blank wasm loading screen while app is starting (#122)

### New Contributors

Code in v1.7.1 contains work from the following first time contributors:

* @dweymouth
* @gezb
* @MaxGyver83
* @ncw
* @typenoob


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
* Renamed release flag --key-store and --key-store-pass to --keystore and --keystore-pass
* Create 16k aligned files for Google Play

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

