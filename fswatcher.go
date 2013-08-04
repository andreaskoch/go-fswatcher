// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	// "github.com/andreaskoch/go-fswatch"
)

const (
	VERSION = "0.1"
)

var usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	// print application info
	fmt.Printf("%s (Version: %s)\n\n", os.Args[0], VERSION)

	// print usage information if no arguments are supplied
	if len(os.Args) == 1 {
		usage()
		os.Exit(1)
	}

	// parse the flags
	flag.Parse()

	// check if the supplied path exists
	if !pathExists(Settings.Path) {
		fmt.Printf("Path %q does not exist.\n", Settings.Path)
		os.Exit(1)
	}

	// distinguish between files and directories
	if ok, _ := isDirectory(Settings.Path); ok {

		fmt.Printf("Watching directory %q%s.\n", Settings.Path, (func() string {
			if Settings.Recurse {
				return " (recursive)"
			}
			return ""
		})())

	} else {
		fmt.Printf("Watching file %q.\n", Settings.Path)
	}
}
