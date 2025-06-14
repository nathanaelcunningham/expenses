#!/bin/bash

set -e

echo "🚀 Deploying Expenses App to Railway..."

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
            echo "  -y, --auto-approve        Skip interactive approval of the Terraform plan"
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
    echo "Example: cp terraform.tfvars.example terraform.tfvars"
    exit 1
fi

echo "📁 Using variables file: $TFVARS_FILE"

echo "🔧 Initializing Terraform..."
terraform init

echo "📋 Planning deployment..."
terraform plan -var-file="$TFVARS_FILE"

if [ "$AUTO_APPROVE" = true ]; then
    echo "🚀 Applying configuration (auto-approved)..."
    terraform apply -auto-approve -var-file="$TFVARS_FILE"
else
    echo "🚀 Applying configuration..."
    terraform apply -var-file="$TFVARS_FILE"
fi

echo "✅ Deployment complete!"
echo ""
echo "📊 Outputs:"
terraform output

echo ""
echo "🎉 Your Expenses App has been deployed to Railway!"
echo "Check the Railway dashboard for service status: $(terraform output -raw project_url)"