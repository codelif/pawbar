#!/usr/bin/bash

sudo cp bar pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
cp kitty.conf "$HOME/.config/pawbar/"
cp pickQuotes.sh "$HOME/.config/pawbar/"
cp quotes.txt "$HOME/.config/pawbar/"
echo -e  "right:\n  - battery\n  - sep\n  - clock" > "$HOME/.config/pawbar/pawbar.yaml"
