#!/bin/bash

# Deploy the URL shortener application to ArgoCD
# This script should be run from the root of the project

set -e

# Get the IP of the K3s node (replace with your server IP)
SERVER_IP=${1:-"your-server-ip"}
REPO_URL=${2:-"https://github.com/your-username/go-url-shortener.git"}

# Update the ArgoCD application with the correct repository URL
echo "Updating application.yaml with repository URL: ${REPO_URL}"
sed -i "s|https://github.com/your-username/go-url-shortener.git|${REPO_URL}|g" k8s/application.yaml

# Update the ingress host with the server IP
echo "Updating ingress.yaml with server IP: ${SERVER_IP}"
sed -i "s|url.your-server-ip.nip.io|url.${SERVER_IP}.nip.io|g" k8s/ingress.yaml

# Update the image in deployment.yaml
echo "Updating deployment.yaml with registry address: ${SERVER_IP}:30500"
sed -i "s|your-registry:30500/url-shortener:latest|${SERVER_IP}:30500/url-shortener:latest|g" k8s/deployment.yaml

# Update the image in kustomization.yaml
echo "Updating kustomization.yaml with registry address: ${SERVER_IP}:30500"
sed -i "s|your-registry:30500/url-shortener|${SERVER_IP}:30500/url-shortener|g" k8s/kustomization.yaml

echo "Configuration updated successfully!"
echo "Now you can apply the ArgoCD application:"
echo "kubectl apply -f k8s/application.yaml"
