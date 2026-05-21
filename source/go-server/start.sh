#!/bin/bash

if [ "$EUID" -ne 0 ]; then
  echo "Error: This script requires root privileges."
  echo "Please run with sudo: sudo $0"
  exit 1
fi

while true; do
  ./OpenTheatre prod 2>&1 | rotatelogs -n 5 ./log 86400
  echo 'Restarting OpenTheatre...'
  sleep 1
done