#!/usr/bin/bash

sudo cp bar pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
cp kitty.conf "$HOME/.config/pawbar/"
echo -e  "right:\n  - battery\n  - space\n  - sep\n  - space\n  - clock" > "$HOME/.config/pawbar/pawbar.yaml"
