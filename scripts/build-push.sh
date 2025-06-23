#!/bin/bash

# Build and push the URL shortener image to the local registry
# This script should be run from the root of the project

set -e

# Get the IP of the K3s node (replace with your server IP)
SERVER_IP=${1:-"your-server-ip"}
IMAGE_NAME="url-shortener"
IMAGE_TAG="latest"
REGISTRY="${SERVER_IP}:30500"
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "Building Docker image: ${FULL_IMAGE_NAME}"
docker build -t "${FULL_IMAGE_NAME}" .

echo "Logging into registry at ${REGISTRY}"
# Note: You need to create a user first using ./scripts/docker-registry.sh create-user
docker login "${REGISTRY}"

echo "Pushing image to registry: ${FULL_IMAGE_NAME}"
docker push "${FULL_IMAGE_NAME}"

echo "Image pushed successfully!"
echo "Now you can deploy the application with ArgoCD using:"
echo "kubectl apply -f k8s/application.yaml"
