# pawbar
A kitten-panel based (uncustomisable) desktop panel for your desktop

![image](https://github.com/user-attachments/assets/4e34cdc8-bdf5-4247-8feb-83716e095ed7)



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


By default the bar is configured with only a clock. You can add modules by editing `$HOME/.config/pawbar/pawbar.yaml`.

It has 4 modules with two of them being hyprland specific:
 - `clock`: A simple date-time module
 - `battery`: A battery module with dynamic icons and colors
 - `space`: A single space
 - `sep`: A single full height vertical bar
 - `hyprws`: A dynamic workspace switcher (with mouse events)
 - `hyprtitle`: A window class & title display

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

## Contribution
Project is in very early stages, any contribution is very much appreciated. 
The codebase (to my best effort, so not very much) clean, modular and extendable. 
