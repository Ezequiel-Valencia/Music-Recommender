#!/bin/sh

# Check if the USER_ID environment variable is set
echo "Checking if USER_ID is Set"
if [ -n "$USER_ID" ]; then
  echo "User id set to: $USER_ID"
  usermod -u $USER_ID threemix
fi

# Gosu https://github.com/tianon/gosu
echo "Checking if GROUP_ID is Set"
if [ -n "$GROUP_ID" ]; then
  echo "Group ID set to: $GROUP_ID"
  groupmod -g $GROUP_ID threemix
fi

chown -R threemix:threemix /app

echo "Starting go application as internal threemix user"
gosu threemix:threemix bin/threemix-executable

