#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if environment is provided
if [ $# -eq 0 ]; then
    print_error "Please provide environment (dev or prod)"
    echo "Usage: $0 <environment>"
    echo "Example: $0 dev"
    exit 1
fi

ENVIRONMENT=$1

# Validate environment
if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
    print_error "Environment must be 'dev' or 'prod'"
    exit 1
fi

# Check required environment variables
if [ -z "$RAILWAY_TOKEN" ]; then
    print_error "RAILWAY_TOKEN environment variable is required"
    exit 1
fi

if [ -z "$TF_VAR_db_password" ]; then
    print_error "TF_VAR_db_password environment variable is required"
    exit 1
fi

print_status "Deploying to $ENVIRONMENT environment..."

# Change to terraform directory
cd "$(dirname "$0")/../terraform"

# Initialize Terraform
print_status "Initializing Terraform..."
terraform init

# Plan deployment
print_status "Planning deployment..."
terraform plan -var-file="../environments/${ENVIRONMENT}.tfvars" -out=tfplan

# Ask for confirmation
echo
read -p "Do you want to apply this plan? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Deployment cancelled"
    exit 0
fi

# Apply deployment
print_status "Applying deployment..."
terraform apply -auto-approve tfplan

print_status "Deployment completed successfully!"

# Show outputs
print_status "Deployment outputs:"
terraform output