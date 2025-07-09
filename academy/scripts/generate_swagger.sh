#!/bin/bash
# Script to generate Swagger docs

cd /home/yunus/Projects/summer-academy/academy
# Use the full path to the swag binary
$HOME/go/bin/swag init -g cmd/main.go -o docs
echo "Swagger documentation generated!"
