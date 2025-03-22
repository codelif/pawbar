# pawbar
A kitten-panel based (uncustomisable) desktop panel for your desktop

## Installing
A basic install script is there:
```sh
./install.sh
```
Though fair caution it installs to `/usr/local/bin`

Also if you want to you can compile pawbar yourself as well:
```sh
go build ./cmd/pawbar
```

## Usage
Run bar by calling the bar script after using the `install.sh` script:
```sh
bar
```


By default the bar is configured with only a clock. You can add modules by editing `$HOME/.config/pawbar/pawbar.yaml`.

It has 4 modules with two of them being hyprland specific:
 - `clock`: A simple date-time module
 - `space`: A single space
 - `hyprws`: A dynamic workspace switcher (with mouse events)
 - `hyprtitle`: A simple window title bar display

A typical config looks like:
```yaml
left:
  - hyprws
  - space
  - hyprtitle

right:
  - clock
```

## Contribution
Project is in very early stages, any contribution is very much appreciated. 
The codebase (to my best effort, so not very much) clean, modular and extendable. 
