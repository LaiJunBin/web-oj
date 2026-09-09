#!/bin/bash

CONFIG_FILE="config.yaml"

CURRENT_CSRF_KEY=$(grep -oP '(?<=csrfKey: !<tag:yaml.org,2002:str> ").*(?=")' "$CONFIG_FILE")

if [ -n "$CURRENT_CSRF_KEY" ]; then
  read -p "CSRF Key is already set in the config file. Do you want to generate a new one? (y/n): " answer
  if [[ "$answer" != "y" ]]; then
    echo "Exiting..."
    exit 0
  fi
fi


CSRF_KEY=$(openssl rand -base64 32)

sed -i "s|csrfKey: !<tag:yaml.org,2002:str> \".*\"|csrfKey: !<tag:yaml.org,2002:str> \"$CSRF_KEY\"|" "$CONFIG_FILE"

echo "CSRF Key has been successfully updated."
