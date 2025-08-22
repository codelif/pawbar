#!/usr/bin/bash

sudo cp pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
[ ! -f "$HOME/.config/pawbar/pawbar.yaml" ] && echo -e  "right:\n  - battery\n  - sep\n  - clock" > "$HOME/.config/pawbar/pawbar.yaml"
