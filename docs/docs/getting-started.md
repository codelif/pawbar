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
git clone --recurse-submodules https://github.com/codelif/pawbar
cd pawbar
go build ./cmd/pawbar
```

Install using the installation script:
```sh
./install.sh
```

# Configuration

`pawbar` is configured using a configuration file.


Configuration file is placed at `$HOME/.config/pawbar/pawbar.yaml`

Simplest configuration file can be:
```yaml
right:
  - clock
```

It sets a right anchored `clock` module with default configuration

A useful default configuration can be:
```yaml
bar:
  truncate_priority: 
    - middle
    - right
    - left
left:
  - ws
  - title
middle:
  - clock:
      format: "%a %H:%M" 
      tick: 1m
      onmouse:
        hover:
          config:
            format: "%a %H:%M:%S"
            tick: 1s
            fg: indianred
right:
  - volume:
      onmouse:
        left:
          run: "pavucontrol"
  - sep
  - backlight
  - sep
  - battery
```
