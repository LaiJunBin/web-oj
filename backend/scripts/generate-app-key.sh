#!/bin/bash

CONFIG_FILE="config.yaml"

CURRENT_APP_KEY=$(grep -oP '(?<=appKey: !<tag:yaml.org,2002:str> ").*(?=")' "$CONFIG_FILE")

if [ -n "$CURRENT_APP_KEY" ]; then
  read -p "APP Key is already set in the config file. Do you want to generate a new one? (y/n): " answer
  if [[ "$answer" != "y" ]]; then
    echo "Exiting..."
    exit 0
  fi
fi


APP_KEY=$(openssl rand -base64 32)

sed -i "s|appKey: !<tag:yaml.org,2002:str> \".*\"|appKey: !<tag:yaml.org,2002:str> \"$APP_KEY\"|" "$CONFIG_FILE"

echo "APP Key has been successfully updated."
