# Railway Deployment with Terraform

This directory contains Terraform configuration and GitHub Actions workflows for deploying the expenses application to Railway.

## Structure

```
deploy/
├── terraform/           # Terraform configuration
│   ├── main.tf         # Main configuration
│   ├── variables.tf    # Input variables
│   ├── outputs.tf      # Output values
│   ├── terraform.tf    # Provider requirements
│   └── modules/        # Reusable modules
│       ├── backend/    # Backend service module
│       ├── frontend/   # Frontend service module
│       └── database/   # Database service module
├── environments/       # Environment-specific variables
│   ├── dev.tfvars     # Development environment
│   └── prod.tfvars    # Production environment
├── scripts/           # Deployment scripts
│   ├── deploy.sh      # Manual deployment script
│   └── destroy.sh     # Cleanup script
└── README.md          # This file
```

## Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **Railway API Token**: Get your token from Railway dashboard
3. **Terraform**: Install Terraform >= 1.0
4. **Docker Images**: Ensure your backend/frontend images are built and pushed to GHCR

## Setup

### 1. Railway API Token

Get your Railway API token:
1. Go to [Railway Dashboard](https://railway.app/dashboard)
2. Click on your profile → Account Settings
3. Go to Tokens tab
4. Generate a new token

### 2. Environment Variables

Set the following environment variables:

```bash
export RAILWAY_TOKEN="your_railway_token"
export TF_VAR_db_password="your_secure_database_password"
```

### 3. GitHub Secrets (for CI/CD)

Add these secrets to your GitHub repository:
- `RAILWAY_TOKEN`: Your Railway API token
- `DEV_DB_PASSWORD`: Database password for development
- `PROD_DB_PASSWORD`: Database password for production

## Manual Deployment

### Deploy to Development

```bash
cd deploy/scripts
RAILWAY_TOKEN="your_token" TF_VAR_db_password="dev_password" ./deploy.sh dev
```

### Deploy to Production

```bash
cd deploy/scripts
RAILWAY_TOKEN="your_token" TF_VAR_db_password="prod_password" ./deploy.sh prod
```

### Destroy Environment

```bash
cd deploy/scripts
RAILWAY_TOKEN="your_token" TF_VAR_db_password="password" ./destroy.sh dev
```

## GitHub Actions CI/CD

### Workflows

1. **terraform-check.yml**: Validates Terraform on PRs
2. **deploy-dev.yml**: Deploys to dev environment on dev/develop branch pushes
3. **deploy-prod.yml**: Deploys to production on main branch pushes

### Branch Strategy

- `dev/develop` → Development environment
- `main` → Production environment

### Environment Protection

Set up GitHub environment protection rules:
1. Go to repository Settings → Environments
2. Create `development` and `production` environments
3. Add required reviewers for production
4. Configure environment secrets

## Services Deployed

### Backend Service
- **Image**: `ghcr.io/nathanaelcunningham/expenses/backend`
- **Environment Variables**: Database connection details
- **Dependencies**: Database service

### Frontend Service  
- **Image**: `ghcr.io/nathanaelcunningham/expenses/frontend`
- **Environment Variables**: Backend API URL
- **Dependencies**: Backend service

### Database Service
- **Image**: PostgreSQL 15 Alpine
- **Environment Variables**: Database credentials
- **Persistent Storage**: Railway-managed

## Customization

### Adding Environment Variables

1. Add variables to `deploy/terraform/variables.tf`
2. Update module configurations in `deploy/terraform/modules/*/main.tf`
3. Update environment files in `deploy/environments/`

### Adding New Services

1. Create new module in `deploy/terraform/modules/`
2. Add module call in `deploy/terraform/main.tf`
3. Update outputs in `deploy/terraform/outputs.tf`

## Troubleshooting

### Common Issues

1. **Railway Token Invalid**: Verify token is correct and active
2. **Image Not Found**: Ensure Docker images are built and pushed to GHCR
3. **Service Dependencies**: Check service startup order in modules

### Debugging

```bash
# Check Terraform plan
cd deploy/terraform
terraform plan -var-file="../environments/dev.tfvars"

# Check Terraform state
terraform show

# View service logs in Railway dashboard
```

## Security Best Practices

- Never commit secrets to git
- Use GitHub environment secrets for CI/CD
- Use strong, unique database passwords
- Regularly rotate API tokens
- Enable Railway's security features

## Support

For issues with:
- **Terraform Railway Provider**: [GitHub Issues](https://github.com/terraform-community-providers/terraform-provider-railway/issues)
- **Railway Platform**: [Railway Help](https://railway.com/help)
- **This Configuration**: Create an issue in this repository