# WBFY Docker Images

This directory contains Dockerfiles for building the WBFY terminal environments used in the Summer Academy platform.

## Image Structure

- `wbfy-base`: Base image with common Linux tools and utilities
- `wbfy-python`: Python environment with common libraries
- `wbfy-golang`: Go environment with workspace setup
- `wbfy-node`: Node.js environment with npm and common tools

## Building and Pushing Images

To build and push all Docker images to Docker Hub:

1. Make sure you have Docker installed and are logged in to Docker Hub:
   ```bash
   docker login
   ```

2. Run the build script:
   ```bash
   chmod +x build-and-push.sh
   ./build-and-push.sh
   ```

## Using the Images

The images are automatically used by the WBFY terminal integration in the Summer Academy platform. Each terminal session launches a container based on the language selected by the user.

## Environment Variables

- `WBFY_CMD`: The command to run when the container starts
- `PROBLEM_TYPE`: The type of problem (e.g., "linux", "dsa", "build")
- `SESSION_ID`: A unique identifier for the terminal session

## Adding a New Language Environment

To add support for a new language:

1. Create a new directory with the language name: `wbfy-[language]`
2. Create a Dockerfile that extends the base image
3. Install the language runtime and common tools
4. Set the default command via the `WBFY_CMD` environment variable
5. Add the new language to the `build-and-push.sh` script
6. Update the `getDockerImage` function in `wbfy.go`
