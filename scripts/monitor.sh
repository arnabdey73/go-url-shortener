#!/bin/bash

# Monitor the URL shortener application deployment
# This script should be run from the root of the project

set -e

# Function to show usage
usage() {
  echo "Usage: $0 [COMMAND]"
  echo "Commands:"
  echo "  status    Show deployment status"
  echo "  logs      Show application logs"
  echo "  metrics   Port-forward to metrics endpoint"
  echo "  argocd    Show ArgoCD application status"
  echo "  help      Show this help message"
  exit 1
}

# Check if a command was provided
if [ $# -eq 0 ]; then
  usage
fi

COMMAND=$1

case $COMMAND in
  status)
    echo "Checking URL Shortener deployment status..."
    kubectl get deployment,pod,svc,ingress,pvc -n url-shortener
    ;;
    
  logs)
    echo "Showing URL Shortener logs..."
    kubectl logs -f -l app=url-shortener -n url-shortener
    ;;
    
  metrics)
    echo "Port-forwarding to metrics endpoint..."
    kubectl port-forward svc/url-shortener -n url-shortener 8080:80
    echo "Metrics available at http://localhost:8080/metrics"
    ;;
    
  argocd)
    echo "Checking ArgoCD application status..."
    kubectl get application url-shortener -n argocd -o wide
    ;;
    
  help)
    usage
    ;;
    
  *)
    echo "Unknown command: $COMMAND"
    usage
    ;;
esac
