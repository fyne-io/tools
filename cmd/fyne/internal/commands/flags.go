package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

type keyValueFlag struct {
	m map[string]string
}

func (k *keyValueFlag) Set(value string) error {
	if k.m == nil {
		k.m = make(map[string]string)
	}
	parts := strings.Split(value, "=")

	if len(parts) < 2 {
		return fmt.Errorf("expected format: key=value, got %s", value)
	}

	k.m[parts[0]] = strings.Join(parts[1:], "=")
	return nil
}

func (k *keyValueFlag) String() string {
	result := ""

	for key, value := range k.m {
		if result == "" {
			result = "\"" + key + "=" + value + "\""
		} else {
			result = result + ", \"" + key + "=" + value + "\""
		}
	}

	return result
}

var stringFlags = map[string]func(*string) cli.Flag{
	"app-id": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "app-id",
			Aliases:     []string{"id"},
			Usage:       "set app-id in reversed domain notation for android, darwin, and windows targets, or a valid provisioning profile for ios",
			Destination: dst,
		}
	},
	"app-version": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "app-version",
			Usage:       "set version number in the form x, x.y or x.y.z semantic version",
			Destination: dst,
		}
	},
	"category": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "category",
			Usage:       "macos: category of the app for store listing",
			Destination: dst,
		}
	},
	"certificate": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "certificate",
			Aliases:     []string{"cert"},
			Usage:       "ios/macos/windows: name of the certificate to sign the build",
			Destination: dst,
		}
	},
	"developer": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "developer",
			Aliases:     []string{"dev"},
			Usage:       "windows: the developer identity for your Microsoft store account",
			Destination: dst,
		}
	},
	"executable": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "executable",
			Aliases:     []string{"exe"},
			Usage:       "set path to executable",
			DefaultText: "current directory main binary",
			Destination: dst,
		}
	},
	"icon": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "icon",
			Usage:       "set name of the application icon file",
			Destination: dst,
		}
	},
	"install-dir": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "install-dir",
			Aliases:     []string{"o"},
			Usage:       "specify location to install to, rather than the OS default",
			Destination: dst,
		}
	},
	"key-name": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "key-name",
			Usage:       "android: alias for the signer's private key, which is needed when reading a .keystore file",
			Destination: dst,
		}
	},
	"key-pass": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "key-pass",
			Usage:       "android: password for the signer's private key, which is needed if the private key is password-protected",
			DefaultText: "read from stdin",
			Destination: dst,
		}
	},
	"key-store": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "key-store",
			Usage:       "android: location of .keystore file containing signing information",
			Destination: dst,
		}
	},
	"key-store-pass": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "key-store-pass",
			Usage:       "android: password for the .keystore file",
			DefaultText: "read from stdin",
			Destination: dst,
		}
	},
	"name": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "name",
			Usage:       "set name of the application",
			DefaultText: "executable file name",
			Destination: dst,
		}
	},
	"output": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "output",
			Usage:       "specify name for the output file",
			DefaultText: "based on current directory",
			Destination: dst,
		}
	},
	"password": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "password",
			Aliases:     []string{"pass"},
			Usage:       "windows: password for the certificate used to sign the build",
			Destination: dst,
		}
	},
	"profile": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "profile",
			Usage:       "ios/macos: name of the provisioning profile for this release build",
			Destination: dst,
			Value:       "XCWildcard",
		}
	},
	"source-dir": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "source-dir",
			Aliases:     []string{"src"},
			Usage:       "set directory to package, if executable is not set",
			Destination: dst,
		}
	},
	"tags": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "tags",
			Usage:       "set comma-separated list of build tags",
			Destination: dst,
		}
	},
	"target": func(dst *string) cli.Flag {
		return &cli.StringFlag{
			Name:        "target",
			Aliases:     []string{"os"},
			Usage:       "set operating system to target (android, android/arm, android/arm64, android/amd64, android/386, darwin, freebsd, ios, linux, netbsd, openbsd, windows)",
			Destination: dst,
		}
	},
}

var boolFlags = map[string]func(*bool) cli.Flag{
	"release": func(dst *bool) cli.Flag {
		return &cli.BoolFlag{
			Name:        "release",
			Usage:       "enable installation in release mode, disable debug, etc",
			Destination: dst,
		}
	},
	"use-raw-icon": func(dst *bool) cli.Flag {
		return &cli.BoolFlag{
			Name:        "use-raw-icon",
			Usage:       "skip any OS-specific icon pre-processing",
			Value:       false,
			Destination: dst,
		}
	},
}

var intFlags = map[string]func(*int) cli.Flag{
	"app-build": func(dst *int) cli.Flag {
		return &cli.IntFlag{
			Name:        "app-build",
			Usage:       "set build number (integer >0, increasing with each build)",
			Destination: dst,
		}
	},
}

var genericFlags = map[string]func(cli.Generic) cli.Flag{
	"metadata": func(dst cli.Generic) cli.Flag {
		return &cli.GenericFlag{
			Name:  "metadata",
			Usage: "specify custom metadata key value pair that you do not want to store in your FyneApp.toml (key=value)",
			Value: dst,
		}
	},
}
