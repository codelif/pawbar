#!/usr/bin/bash

sudo cp bar pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
cp kitty.conf "$HOME/.config/pawbar/"
echo -e "right:\n - clock" > "$HOME/.config/pawbar/pawbar.yaml"
