# go-fswatcher

Watch a file or folder for changes and execute a custom command.

## Usage

### Watch a file for changes

```bash
go-fswatcher -path some-file.txt -command "some-command"
```

### Watch a folder for changes

```bash
go-fswatcher -path some-folder -command "some-command"
```

If you want to watch for changes recursivly you can add the `recurse` option:

```bash
go-fswatcher -path some-folder -recurse -command "some-command"
```

## Examples

### Using go-fswatcher to autobuild your go applications

Build your go application everytime a file changes in your package directory.

```bash
go-fswatcher -path $GOPATH/src/github.com/andreaskoch/allmark -recurse -command "go install"
```

## Build Status

[![Build Status](https://travis-ci.org/andreaskoch/go-fswatcher.png?branch=master)](https://travis-ci.org/andreaskoch/go-fswatcher)

## Dependencies

- [github.com/andreaskoch/go-fswatch](https://github.com/andreaskoch/go-fswatch) for monitoring the filesystem for changes

## Contribute

If you have an idea how to make this little tool better please send me a message or a pull request.

All contributions are welcome.
