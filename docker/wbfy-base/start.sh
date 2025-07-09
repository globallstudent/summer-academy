#!/bin/bash

# Get environment variables or set defaults
WBFY_CMD=${WBFY_CMD:-"bash"}
PROBLEM_TYPE=${PROBLEM_TYPE:-"linux"}
SESSION_ID=${SESSION_ID:-"default"}

echo "Starting terminal environment for $PROBLEM_TYPE"
echo "Session ID: $SESSION_ID"
echo "Command: $WBFY_CMD"

# If there's a setup.sh in the workspace, run it
if [ -f "/workspace/setup.sh" ]; then
    echo "Running setup script..."
    chmod +x /workspace/setup.sh
    /workspace/setup.sh
fi

# Execute the command
exec $WBFY_CMD
