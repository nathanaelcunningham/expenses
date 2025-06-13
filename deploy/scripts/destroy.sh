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

print_warning "WARNING: This will destroy all resources in the $ENVIRONMENT environment!"
print_warning "This action cannot be undone!"

# Ask for confirmation
echo
read -p "Are you sure you want to destroy the $ENVIRONMENT environment? Type 'destroy' to confirm: " -r
echo
if [[ "$REPLY" != "destroy" ]]; then
    print_warning "Destruction cancelled"
    exit 0
fi

print_status "Destroying $ENVIRONMENT environment..."

# Change to terraform directory
cd "$(dirname "$0")/../terraform"

# Initialize Terraform
print_status "Initializing Terraform..."
terraform init

# Plan destruction
print_status "Planning destruction..."
terraform plan -destroy -var-file="../environments/${ENVIRONMENT}.tfvars" -out=destroy-plan

# Apply destruction
print_status "Destroying resources..."
terraform apply -auto-approve destroy-plan

print_status "Environment destroyed successfully!"