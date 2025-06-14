# Expenses App - Railway Deployment

This directory contains Terraform configuration for deploying the Expenses tracking application to Railway.

## Architecture

The application consists of three main components:
- **Database**: PostgreSQL with persistent volume and TCP proxy for external access
- **Backend**: Go gRPC service connected to the database
- **Frontend**: React application that communicates with the backend

## Prerequisites

1. [Terraform](https://www.terraform.io/downloads.html) >= 1.0
2. [Railway CLI](https://docs.railway.app/develop/cli) (optional, for local development)
3. Railway account and API token

## Configuration

Create a `terraform.tfvars` file from the example template:

```bash
cd deploy/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with your values:

```hcl
# Railway API token
railway_token = "your_railway_api_token_here"

# Project configuration
project_name        = "expenses-app"
project_description = "Expenses tracking application with Go backend and React frontend"

# Repository configuration
github_repo   = "https://github.com/your-username/expenses"
github_branch = "main"

# Database configuration
postgres_user     = "postgres"
postgres_password = "your_secure_password_here"
postgres_db       = "expenses"
```

**Note:** The `terraform.tfvars` file is automatically ignored by Git to keep your secrets safe.

## Deployment

### Quick Deploy
```bash
# Deploy with default terraform.tfvars
./scripts/deploy.sh

# Deploy with custom tfvars file
./scripts/deploy.sh -f my-config.tfvars

# Deploy with auto-approval (no interactive prompts)
./scripts/deploy.sh -y
```

### Manual Deployment
```bash
cd terraform
terraform init
terraform plan -var-file="terraform.tfvars"
terraform apply -var-file="terraform.tfvars"
```

### Script Options
Both deployment scripts support the following options:
- `-f, --tfvars-file FILE`: Use specific tfvars file (default: `terraform.tfvars`)
- `-y, --auto-approve`: Skip interactive approval prompts
- `-h, --help`: Show help message

## Accessing Your Application

After deployment, Terraform will output URLs for:
- Frontend application
- Backend API
- Database connection details
- Railway project dashboard

## Database Access

The PostgreSQL database is accessible both internally (within Railway) and externally:
- **Internal**: Uses Railway's private domain for service-to-service communication
- **External**: Uses TCP proxy for connections from outside Railway

Connection details are available in the Terraform outputs.

## Cleanup

To destroy all infrastructure:
```bash
# Destroy with default terraform.tfvars
./scripts/destroy.sh

# Destroy with custom tfvars file
./scripts/destroy.sh -f my-config.tfvars

# Destroy with auto-approval (no confirmation prompt)
./scripts/destroy.sh -y
```

## Directory Structure

```
deploy/
├── terraform/
│   ├── main.tf                    # Main configuration with locals
│   ├── variables.tf               # Input variables
│   ├── outputs.tf                 # Output values
│   ├── terraform.tf               # Provider configuration
│   ├── terraform.tfvars.example   # Example variables file
│   ├── .gitignore                 # Git ignore for sensitive files
│   └── modules/
│       ├── database/              # PostgreSQL module
│       ├── backend/               # Go service module
│       └── frontend/              # React app module
├── scripts/
│   ├── deploy.sh                  # Enhanced deployment script
│   └── destroy.sh                 # Enhanced cleanup script
└── README.md                      # This file
```

## Troubleshooting

1. **Missing tfvars file**: Copy `terraform.tfvars.example` to `terraform.tfvars` and fill in your values
2. **Authentication Issues**: Ensure your Railway API token in `terraform.tfvars` is valid and has necessary permissions
3. **GitHub Access**: Make sure your repository is public or Railway has access to private repos
4. **Build Failures**: Check that your `railway.json` files in backend/frontend directories are properly configured
5. **Database Connection**: Verify the database service is running before backend deployment
6. **Variable Validation**: Use `terraform validate` to check for syntax errors in your configuration

For more help, check the [Railway documentation](https://docs.railway.app/) or the Terraform Railway provider [documentation](https://registry.terraform.io/providers/terraform-community-providers/railway/latest/docs).