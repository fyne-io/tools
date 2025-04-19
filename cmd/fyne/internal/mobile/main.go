// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobile

import (
	"flag"
)

type command struct {
	run  func(*command) error
	Flag flag.FlagSet
	Name string

	IconPath, AppName      string
	Version, Cert, Profile string
	Build                  int
}
