// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"os"
)

const (
	defaultRecurse = false
)

type WatcherSettings struct {
	Path    string
	Recurse bool
}

var Settings WatcherSettings = WatcherSettings{}

func init() {

	// use the current directory as the default path
	defaultPath, err := os.Getwd()
	if err != nil {
		defaultPath = "."
	}

	flag.StringVar(&Settings.Path, "path", defaultPath, "The file or directory to watch")

	flag.BoolVar(&Settings.Recurse, "recurse", defaultRecurse, "A flag indicating whether to recurse or not")
}
