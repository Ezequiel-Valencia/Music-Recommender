#!/bin/sh

# Check if the USER_ID environment variable is set
if [ -n "$USER_ID" ]; then
  usermod -u $USER_ID threemix
fi

# Gosu https://github.com/tianon/gosu
if [ -n "$GROUP_ID" ]; then
  groupmod -g $GROUP_ID threemix
fi

chown -R threemix:threemix /app

if [ "$(id -u)" -eq 0 ]; then
  gosu threemix:threemix air -c .air.toml
fi

