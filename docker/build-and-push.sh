#!/bin/bash

set -e

# Docker Hub username
DOCKER_USER=globalstudent

# Build base image first
echo "Building base image..."
cd $(dirname "$0")/wbfy-base
docker build -t $DOCKER_USER/wbfy-base:latest .
docker push $DOCKER_USER/wbfy-base:latest

# Build language-specific images
for lang in python golang node; do
    echo "Building $lang image..."
    cd $(dirname "$0")/wbfy-$lang
    docker build -t $DOCKER_USER/wbfy-$lang:latest .
    docker push $DOCKER_USER/wbfy-$lang:latest
done

echo "All images built and pushed successfully!"
