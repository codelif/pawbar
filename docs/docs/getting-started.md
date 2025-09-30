# What is `pawbar`?

`pawbar` is a customizable status bar for GNU/Linux systems. Currently it supports wayland compositors but compatibility may be added for X.org and macOS systems.

# Installation

Currently, you have to manually compile and install `pawbar`.

## Manual
The following dependencies are required at compile time:
### Dependencies
 - `go`
 - `udev`
 - `librsvg`


Clone and compile `pawbar`
```sh
git clone https://github.com/codelif/pawbar
cd pawbar
go build ./cmd/pawbar
```

Install using the installation script:
```sh
./install.sh
```
