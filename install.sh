#!/usr/bin/bash

sudo cp bar pawbar /usr/local/bin/
mkdir -p "$HOME/.config/pawbar/"
cp kitty.conf "$HOME/.config/pawbar/"
echo -e "right:\n\t- battery\n\t- space\n\t- sep\n\t- space\n\t- clock" > "$HOME/.config/pawbar/pawbar.yaml"
