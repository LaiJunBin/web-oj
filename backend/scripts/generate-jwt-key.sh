#!/bin/bash

CONFIG_FILE="config.yaml"

CURRENT_JWT_KEY=$(grep -oP '(?<=jwtKey: !<tag:yaml.org,2002:str> ").*(?=")' "$CONFIG_FILE")

if [ -n "$CURRENT_JWT_KEY" ]; then
  read -p "JWT Key is already set in the config file. Do you want to generate a new one? (y/n): " answer
  if [[ "$answer" != "y" ]]; then
    echo "Exiting..."
    exit 0
  fi
fi


JWT_KEY=$(openssl rand -base64 32)

sed -i "s|jwtKey: !<tag:yaml.org,2002:str> \".*\"|jwtKey: !<tag:yaml.org,2002:str> \"$JWT_KEY\"|" "$CONFIG_FILE"

echo "JWT Key has been successfully updated."
