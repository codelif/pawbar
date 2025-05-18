# pawbar
A kitten-panel based desktop panel for your desktop

![image](https://github.com/user-attachments/assets/b8cdfd44-ca66-45df-a8eb-d8142d0e4ffb)

## Why?
Due to the existence of modern terminal standards (especially in kitty), a tiny terminal windows support true colors, [text-sizing](https://sw.kovidgoyal.net/kitty/text-sizing-protocol/), [images](https://sw.kovidgoyal.net/kitty/graphics-protocol/), mouse support (clicking, dragging, hover and scrolling), etc. Which makes it a viable option for a status bar, instead of a GUI Toolkits (like GTK). Kitty enables this using its [panel](https://sw.kovidgoyal.net/kitty/kittens/panel/) mode/kitten, through which it offers all the customization/capabilities of a kitty terminal window but as a status bar on top of your screen. 


## Installing
A basic install script is there (you need to compile pawbar before running the script): 
```sh
go build ./cmd/pawbar
./install.sh
```
Though fair caution it installs to `/usr/local/bin`

> [!NOTE]
> I will add other installation methods when I am satisfied with this project.
## Usage
Run bar by calling the bar script after using the `install.sh` script:
```sh
bar
```


By default the bar is configured with only a clock and a battery. You can add modules by editing `$HOME/.config/pawbar/pawbar.yaml`.

It has 13 modules :
 - `clock`: A simple date-time module (format changable on click)
 - `wifi`: A simple wifi conenction indicator (for now, without interactive menu)
 - `bluetooth`: A simple bluetooth conenction indicator (for now, without interactive menu)
 - `battery`: A battery module with dynamic icons and colors
 - `backlight`: A screen brightness indicator (interactable)
 - `volume`: A Volume level indicator (interactable)
 - `ram`: RAM usage (format changable on click)
 - `cpu`: CPU usage
 - `disk`: Disk usage (format changable on click)
 - `locale`: Current locale
 - `ws`: A dynamic workspace switcher (hyprland/i3/sway) with (with mouse events) (change workspace on click)
 - `title`: A window class & title display ((hyprland/i3/sway)
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
     - [ ] wifi (with menu)
     - [ ] bluetooth (with menu)
     - [ ] tray
     - [ ] workspace and title for more WMs
     - [ ] Suggest more
 - [ ] Extended module config
 - [ ] Extended bar config
 - [ ] Menu support

## Contribution
Project is in very early stages, any contribution is very much appreciated. 
