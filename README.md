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
       bundle           Embeds static content into your go application.
       env, e           The env command prints the Fyne module and environment information
       install, get, i  Packages and installs an application.
       package, p       Packages an application for distribution.
       release, r       Prepares an application for public distribution.
       version, v       Shows version information for fyne.
       serve, s         Package an application using WebAssembly and expose it via a web server.
       translate, t     Scans for new translation strings.
       build, b         Build an application.
       help, h          Shows a list of commands or help for one command
    
    GLOBAL OPTIONS:
       --help, -h  show help
