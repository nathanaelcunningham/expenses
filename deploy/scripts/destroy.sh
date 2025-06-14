#!/bin/bash

set -e

echo "⚠️  Destroying Expenses App infrastructure..."

# Parse command line arguments
TFVARS_FILE="terraform.tfvars"
AUTO_APPROVE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--tfvars-file)
            TFVARS_FILE="$2"
            shift 2
            ;;
        -y|--auto-approve)
            AUTO_APPROVE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -f, --tfvars-file FILE    Use specific tfvars file (default: terraform.tfvars)"
            echo "  -y, --auto-approve        Skip interactive confirmation"
            echo "  -h, --help               Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option $1"
            exit 1
            ;;
    esac
done

# Navigate to terraform directory
cd "$(dirname "$0")/../terraform"

# Check if tfvars file exists
if [ ! -f "$TFVARS_FILE" ]; then
    echo "❌ Error: Terraform variables file '$TFVARS_FILE' not found"
    echo "Please create $TFVARS_FILE or copy from terraform.tfvars.example"
    exit 1
fi

echo "📁 Using variables file: $TFVARS_FILE"

echo "📋 Planning destruction..."
terraform plan -destroy -var-file="$TFVARS_FILE"

if [ "$AUTO_APPROVE" = false ]; then
    echo ""
    read -p "⚠️  Are you sure you want to destroy all infrastructure? This cannot be undone. [y/N]: " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Destruction cancelled."
        exit 0
    fi
fi

echo "🔥 Destroying infrastructure..."
terraform destroy -auto-approve -var-file="$TFVARS_FILE"

echo "✅ Infrastructure destroyed successfully!"