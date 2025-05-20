#!/bin/bash

FILE=~/.config/pawbar/quotes.txt

# Remove empty lines and comments
quotes=$(grep '[^[:space:]]' "$FILE" | sed '/^#/d')

# Convert to array
IFS=$'\n' read -rd '' -a lines <<< "$quotes"

# Get number of lines
count=${#lines[@]}

# Pick random line
if [ "$count" -gt 0 ]; then
    index=$((RANDOM % count))
    echo "${lines[$index]}"
else
    echo "No quotes available"
    exit 1
fi

