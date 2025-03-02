# Fyne CLI tools

Toolchain and helpful commands for building and managing Fyne apps.

## Installation

Using the standard go tools you can install Fyne's command line tool using:

    go install fyne.io/tools/cmd/fyne@latest

## Synopsis

To list all available commands enter `fyne help`:

    NAME:
       fyne - A command line helper for various Fyne tools.
    
    USAGE:
       fyne [global options] command [command options]
    
    DESCRIPTION:
       The fyne command provides tooling for fyne applications and to assist in their development.
    
    COMMANDS:
       init             Initializes a new Fyne project
       env, e           Prints the Fyne module and environment information
       build, b         Builds an application
       package, p       Packages an application for distribution
       release, r       Prepares an application for public distribution
       install, get, i  Packages and installs an application
       serve, s         Packages an application using WebAssembly and exposes it via a web server
       translate, t     Scans for new translation strings
       version, v       Shows version information for fyne
       bundle           Embeds static content into your go application
       help, h          Shows a list of commands or help for one command
    
    GLOBAL OPTIONS:
       --help, -h  show help
