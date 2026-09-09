#!/bin/bash

CONFIG_FILE="config.yaml"

CURRENT_STORE_KEY=$(grep -oP '(?<=storeKey: !<tag:yaml.org,2002:str> ").*(?=")' "$CONFIG_FILE")

if [ -n "$CURRENT_STORE_KEY" ]; then
  read -p "STORE Key is already set in the config file. Do you want to generate a new one? (y/n): " answer
  if [[ "$answer" != "y" ]]; then
    echo "Exiting..."
    exit 0
  fi
fi


STORE_KEY=$(openssl rand -base64 32)

sed -i "s|storeKey: !<tag:yaml.org,2002:str> \".*\"|storeKey: !<tag:yaml.org,2002:str> \"$STORE_KEY\"|" "$CONFIG_FILE"

echo "STORE Key has been successfully updated."
