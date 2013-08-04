// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// la di da

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

var (
	workingDirectory string
)

const (
	VERSION = "0.1"
)

var usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func init() {

	// determine the current working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workingDirectory = wd

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

	// clean the path
	Settings.Path = filepath.Clean(Settings.Path)

	// normalize the path
	absolutePath, err := filepath.Abs(Settings.Path)
	if err != nil {
		fmt.Printf("Cannot determine the absolute path from %q.", Settings.Path)
		os.Exit(1)
	}

	Settings.Path = absolutePath

	// distinguish between files and directories
	stopFilesystemWatcher := make(chan bool, 1)
	if ok, _ := isDirectory(Settings.Path); ok {

		fmt.Printf("Watching directory %q%s.\n", Settings.Path, (func() string {
			if Settings.Recurse {
				return " (recursive)"
			}
			return ""
		})())

		watchDirectory(Settings.Path, Settings.Recurse, Settings.Command, stopFilesystemWatcher)

	} else {
		fmt.Printf("Watching file %q.\n", Settings.Path)
		watchFile(Settings.Path, Settings.Command, stopFilesystemWatcher)
	}

	// stop checker
	fmt.Println(`Write "stop" and press <Enter> to stop.`)

	stopApplication := make(chan bool, 1)
	go func() {
		input := bufio.NewReader(os.Stdin)

		for {

			userInput, err := input.ReadString('\n')
			if err != nil {
				fmt.Println("%s\n", err)
			}

			if command := strings.ToLower(strings.TrimSpace(userInput)); command == "stop" {

				fmt.Println()

				stopFilesystemWatcher <- true
				stopApplication <- true
			}
		}
	}()

	select {
	case <-stopApplication:
		fmt.Printf("Stopped watching %q.\n", Settings.Path)
	}

	os.Exit(0)
}

func watchDirectory(directoryPath string, recurse bool, commandText string, stop chan bool) {

	skipFiles := func(path string) bool {
		return false
	}

	go func() {
		folderWatcher := fswatch.NewFolderWatcher(directoryPath, recurse, skipFiles).Start()

		for folderWatcher.IsRunning() {

			select {
			case <-folderWatcher.Change:
				fmt.Printf("Directory %q changed.\n", directoryPath)

				go func() {
					execute(commandText)
				}()

			case <-stop:
				fmt.Printf("Stopping directory watcher for %q.\n", directoryPath)
				folderWatcher.Stop()

			case <-folderWatcher.Stopped:
				break
			}
		}

		fmt.Printf("Watcher for directory %q stopped.\n", directoryPath)
	}()
}

func watchFile(filePath string, commandText string, stop chan bool) {

	go func() {
		fileWatcher := fswatch.NewFileWatcher(filePath).Start()

		for fileWatcher.IsRunning() {

			select {
			case <-fileWatcher.Modified:
				fmt.Printf("File %q has been modified.\n", filePath)

				go func() {
					execute(commandText)
				}()

			case <-fileWatcher.Moved:
				fmt.Printf("File %q has been moved.\n", filePath)

				go func() {
					execute(commandText)
				}()

			case <-stop:
				fmt.Printf("Stopping file watcher for %q.\n", filePath)
				fileWatcher.Stop()

			case <-fileWatcher.Stopped:
				break
			}
		}

		fmt.Printf("Watcher for file %q stopped.\n", filePath)
	}()
}

func execute(commandText string) {

	fmt.Printf("Executing command: %s\n", commandText)
	fmt.Println()

	command := getCmd(commandText)
	if err := command.Start(); err != nil {
		fmt.Printf("Launch error: %s\n", err)
	}
}

func getCmd(commandText string) *exec.Cmd {
	if commandText == "" {
		return nil
	}

	components := strings.Split(commandText, " ")

	// get the command name
	commandName := components[0]

	// get the command arguments
	arguments := make([]string, 0)
	if len(components) > 1 {
		arguments = components[1:]
	}

	// create the command
	command := exec.Command(commandName, arguments...)

	// set the working directory
	command.Dir = workingDirectory

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
