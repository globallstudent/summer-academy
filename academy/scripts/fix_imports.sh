#!/bin/bash

# Script to fix import paths from yourusername to globallstudent
find /home/yunus/Projects/summer-academy/academy -type f -name "*.go" -exec sed -i 's|github.com/yourusername/academy|github.com/globallstudent/academy|g' {} \;

echo "Import paths updated successfully."
