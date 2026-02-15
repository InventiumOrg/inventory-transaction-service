#!/bin/bash

# Extract service name from image name and convert to underscore format
# Usage: ./extract-service-name.sh <org>/<service-name>
# Example: ./extract-service-name.sh inventium/inventory-transaction-service
# Output: inventory_transaction_service

IMAGE_NAME="${1:-inventium/inventory-transaction-service}"

# Extract service name (part after /)
SERVICE_NAME=$(echo "$IMAGE_NAME" | cut -d'/' -f2)

# Convert hyphens to underscores
SERVICE_NAME_UNDERSCORE=$(echo "$SERVICE_NAME" | tr '-' '_')

echo "$SERVICE_NAME_UNDERSCORE"
