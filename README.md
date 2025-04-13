# pawbar
A kitten-panel based desktop panel for your desktop

![image](https://github.com/user-attachments/assets/b8cdfd44-ca66-45df-a8eb-d8142d0e4ffb)



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

> [!NOTE]
> I will add other installation methods when I am satisfied with this project.
## Usage
Run bar by calling the bar script after using the `install.sh` script:
```sh
bar
```


By default the bar is configured with only a clock and a battery. You can add modules by editing `$HOME/.config/pawbar/pawbar.yaml`.

It has 6 modules with two of them being hyprland specific:
 - `clock`: A simple date-time module
 - `battery`: A battery module with dynamic icons and colors
 - `backlight`: A screen brightness indicator
 - `ram`: RAM usage
 - `cpu`: CPU usage
 - `disk`: Disk usage
 - `locale`: Current locale
 - `hyprws`: A dynamic workspace switcher (with mouse events)
 - `hyprtitle`: A window class & title display
 - `space`: A single space
 - `sep`: A full height vertical bar and a space on either side

A typical(my) config looks like:
```yaml
left:
  - hyprws
  - hyprtitle

right:
  - battery
  - space
  - sep
  - space
  - clock
```

## Roadmap
 - [x] Running
 - [ ] Modules and Services:
     - [ ] volume (service is done, module remaining)
     - [ ] wifi
     - [ ] bluetooth
     - [ ] tray
     - [ ] Suggest more
 - [ ] Extended module config
 - [ ] Extended bar config
 - [ ] Menu support

## Contribution
Project is in very early stages, any contribution is very much appreciated. 
