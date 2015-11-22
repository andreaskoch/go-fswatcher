// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/andreaskoch/go-fswatch"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// VERSION contains the version number of the application
	VERSION = "0.2.0"
)

const checkIntervalInSeconds = 2

var usage = func() {
	message("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	// print application info
	message("%s (Version: %s)\n\n", os.Args[0], VERSION)

	// print usage information if no arguments are supplied
	if len(os.Args) == 1 {
		usage()
		os.Exit(1)
	}

	// parse the flags
	flag.Parse()

	// check if the supplied path exists
	if !pathExists(Settings.Path) {
		message("Path %q does not exist.", Settings.Path)
		os.Exit(1)
	}

	// clean the path
	Settings.Path = filepath.Clean(Settings.Path)

	// normalize the path
	absolutePath, err := filepath.Abs(Settings.Path)
	if err != nil {
		message("Cannot determine the absolute path from %q.", Settings.Path)
		os.Exit(1)
	}

	Settings.Path = absolutePath

	// distinguish between files and directories
	stopFilesystemWatcher := make(chan bool, 1)
	if ok, _ := isDirectory(Settings.Path); ok {

		message("Watching directory %q%s.", Settings.Path, (func() string {
			if Settings.Recurse {
				return " (recursive)"
			}
			return ""
		})())

		watchDirectory(Settings.Path, Settings.Recurse, Settings.Command, stopFilesystemWatcher)

	} else {
		message("Watching file %q.", Settings.Path)
		watchFile(Settings.Path, Settings.Command, stopFilesystemWatcher)
	}

	// stop checker
	message(`Write "stop" and press <Enter> to stop.`)

	stopApplication := make(chan bool, 1)
	go func() {
		input := bufio.NewReader(os.Stdin)

		for {

			userInput, err := input.ReadString('\n')
			if err != nil {
				fmt.Printf("%s\n", err)
			}

			if command := strings.ToLower(strings.TrimSpace(userInput)); command == "stop" {

				// empty line
				message("")

				stopFilesystemWatcher <- true
				stopApplication <- true
			}
		}
	}()

	select {
	case <-stopApplication:
		debug("Stopped watching %q.", Settings.Path)
	}

	os.Exit(0)
}

func watchDirectory(directory string, recurse bool, commandText string, stop chan bool) {

	skipFiles := func(path string) bool {
		return false
	}

	go func() {
		folderWatcher := fswatch.NewFolderWatcher(directory, recurse, skipFiles, checkIntervalInSeconds)
		folderWatcher.Start()

		for folderWatcher.IsRunning() {

			select {
			case <-folderWatcher.Modified():
				debug("Directory %q changed.", directory)

				go func() {
					execute(directory, commandText)
				}()

			case <-stop:
				debug("Stopping directory watcher for %q.", directory)
				folderWatcher.Stop()

			case <-folderWatcher.Stopped():
				break
			}
		}

		debug("Watcher for directory %q stopped.", directory)
	}()
}

func watchFile(file string, commandText string, stop chan bool) {

	directory := filepath.Dir(file)

	go func() {
		fileWatcher := fswatch.NewFileWatcher(file, checkIntervalInSeconds)
		fileWatcher.Start()

		for fileWatcher.IsRunning() {

			select {
			case <-fileWatcher.Modified():
				debug("File %q has been modified.", file)

				go func() {
					execute(directory, commandText)
				}()

			case <-fileWatcher.Moved():
				debug("File %q has been moved.", file)

				go func() {
					execute(directory, commandText)
				}()

			case <-stop:
				debug("Stopping file watcher for %q.", file)
				fileWatcher.Stop()

			case <-fileWatcher.Stopped():
				break
			}
		}

		debug("Watcher for file %q stopped.", file)
	}()
}

func execute(directory, commandText string) {

	// get the command
	command := getCmd(directory, commandText)

	// execute the command
	if err := command.Start(); err != nil {
		fmt.Println(err)
	}

	// wait for the command to finish
	command.Wait()

	fmt.Println()
}

func getCmd(directory, commandText string) *exec.Cmd {
	if commandText == "" {
		return nil
	}

	components := strings.Split(commandText, " ")

	// get the command name
	commandName := components[0]

	// get the command arguments
	var arguments []string
	if len(components) > 1 {
		arguments = components[1:]
	}

	// create the command
	command := exec.Command(commandName, arguments...)

	// set the working directory
	command.Dir = directory

	// redirect command io
	redirectCommandIO(command)

	return command
}

func redirectCommandIO(cmd *exec.Cmd) (*os.File, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	//direct. Masked passwords work OK!
	cmd.Stdin = os.Stdin
	return nil, err
}

func debug(text string, args ...interface{}) {
	if !Settings.Verbose {
		return
	}

	message(text, args)
}

func message(text string, args ...interface{}) {

	// append newline character
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}

	fmt.Printf(text, args...)
}
