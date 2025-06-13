# Railway Deployment with Terraform & Config-as-Code

This directory contains Terraform configuration and GitHub Actions workflows for deploying the expenses application to Railway using Railway's native build system and managed PostgreSQL.

## 🏗️ Architecture

Our modernized deployment leverages Railway's native features:
- **Source-based deployments** from GitHub (no Docker building required)
- **Railway Config-as-Code** via `railway.json` files
- **Managed PostgreSQL** using Railway's template
- **Automatic SSL** and environment variable management

## Structure

```
deploy/
├── terraform/           # Terraform configuration
│   ├── main.tf         # Main configuration with Railway project/environment
│   ├── variables.tf    # Input variables (GitHub repo, branches, etc.)
│   ├── outputs.tf      # Output values (service URLs)
│   ├── terraform.tf    # Provider requirements
│   └── modules/        # Reusable modules
│       ├── backend/    # Backend service (source-based deployment)
│       ├── frontend/   # Frontend service (source-based deployment)
│       └── database/   # PostgreSQL service (Railway template)
├── environments/       # Environment-specific variables
│   ├── dev.tfvars     # Development (dev branch)
│   └── prod.tfvars    # Production (main branch)
├── scripts/           # Manual deployment scripts
│   ├── deploy.sh      # Manual deployment script
│   └── destroy.sh     # Cleanup script
└── README.md          # This file

# Railway Config Files (in service directories)
backend/railway.json    # Backend build/deploy configuration
frontend/railway.json   # Frontend build/deploy configuration
```

## Prerequisites

1. **Railway Account**: Sign up at [railway.app](https://railway.app)
2. **Railway API Token**: Get your token from Railway dashboard
3. **Terraform**: Install Terraform >= 1.0
4. **GitHub Repository**: Your code must be in a GitHub repository

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

### 3. GitHub Repository Configuration

Ensure your repository URL is correctly set in the Terraform variables:
- Update `github_repo` in `deploy/environments/dev.tfvars` and `prod.tfvars`
- Verify branch names match your repository structure

### 4. GitHub Secrets (for CI/CD)

Add these secrets to your GitHub repository:
- `RAILWAY_TOKEN`: Your Railway API token
- `DEV_DB_PASSWORD`: Database password for development
- `PROD_DB_PASSWORD`: Database password for production

## Railway Config Files

### Backend Configuration (`backend/railway.json`)

```json
{
  "$schema": "https://railway.com/railway.schema.json",
  "build": {
    "builder": "nixpacks",
    "buildCommand": "go build -o main cmd/server/main.go",
    "watchPatterns": ["cmd/**", "internal/**", "pkg/**", "go.mod", "go.sum"]
  },
  "deploy": {
    "startCommand": "./main",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 30,
    "restartPolicyType": "on_failure"
  }
}
```

### Frontend Configuration (`frontend/railway.json`)

```json
{
  "$schema": "https://railway.com/railway.schema.json",
  "build": {
    "builder": "nixpacks",
    "buildCommand": "npm run build",
    "watchPatterns": ["src/**", "public/**", "package.json", "vite.config.js"]
  },
  "deploy": {
    "startCommand": "npx serve -s build -l 3000",
    "healthcheckPath": "/",
    "healthcheckTimeout": 30,
    "restartPolicyType": "on_failure"
  }
}
```

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

- `dev/develop` → Development environment (Railway builds from dev branch)
- `main` → Production environment (Railway builds from main branch)

### Simplified CI/CD Flow

1. **Code Push** → GitHub triggers workflow
2. **Terraform Deployment** → Creates/updates Railway services
3. **Railway Auto-Build** → Railway automatically builds from GitHub source
4. **Auto-Deploy** → Railway deploys built applications

## Services Deployed

### Backend Service
- **Source**: GitHub repository `backend/` directory
- **Builder**: Nixpacks (auto-detects Go)
- **Build Command**: `go build -o main cmd/server/main.go`
- **Start Command**: `./main`
- **Environment Variables**: Database connection details via Railway

### Frontend Service  
- **Source**: GitHub repository `frontend/` directory
- **Builder**: Nixpacks (auto-detects Node.js)
- **Build Command**: `npm run build`
- **Start Command**: `npx serve -s build -l 3000`
- **Environment Variables**: Backend API URL

### Database Service
- **Template**: Railway's managed PostgreSQL
- **Features**: Auto-SSL, automatic backups, DATABASE_URL provision
- **Management**: Built-in Railway database UI

## Key Benefits

### 🚀 Performance
- **Faster Builds**: Railway's optimized build environment
- **Better Caching**: Native Nixpacks layer caching
- **CDN Integration**: Automatic static asset optimization

### 🔧 Maintenance
- **No Docker Management**: Railway handles all containerization
- **Automatic SSL**: Built-in certificate management
- **Managed Database**: No PostgreSQL container maintenance required

### 📝 Developer Experience
- **Config-as-Code**: Build/deploy settings tracked in git
- **Auto-detection**: Nixpacks automatically detects frameworks
- **Hot Reloads**: Watch patterns trigger rebuilds on file changes

## Customization

### Adding Environment Variables

1. Add variables to respective module `main.tf` files
2. Update environment files in `deploy/environments/`
3. Railway will automatically inject DATABASE_URL for database connections

### Adding New Services

1. Create new module in `deploy/terraform/modules/`
2. Add `railway.json` config file in service directory
3. Update main.tf with module call

### Modifying Build Process

Edit `railway.json` files in service directories:
- `buildCommand`: Custom build commands
- `startCommand`: Custom start commands  
- `watchPatterns`: Files that trigger rebuilds
- `healthcheckPath`: Health check endpoints

## Troubleshooting

### Common Issues

1. **GitHub Repository Access**: Ensure Railway can access your repository
2. **Build Failures**: Check `railway.json` build commands match your setup
3. **Service Dependencies**: Verify database service starts before apps

### Debugging

```bash
# Check Terraform plan
cd deploy/terraform
terraform plan -var-file="../environments/dev.tfvars"

# Check service logs in Railway dashboard
# Railway provides real-time build and runtime logs

# Validate railway.json files
# Use Railway CLI: railway login && railway status
```

## Migration from Docker-based Deployment

If migrating from the previous Docker-based setup:

1. **Remove Docker files**: Dockerfiles are no longer needed
2. **Update CI/CD**: GitHub Actions no longer build/push images
3. **Configure railway.json**: Define build/deploy steps
4. **Update variables**: Switch from image tags to branch names

## Security Best Practices

- Never commit secrets to git
- Use GitHub environment secrets for CI/CD
- Use strong, unique database passwords
- Regularly rotate API tokens
- Enable Railway's security features and monitoring

## Support

For issues with:
- **Terraform Railway Provider**: [GitHub Issues](https://github.com/terraform-community-providers/terraform-provider-railway/issues)
- **Railway Platform**: [Railway Help](https://railway.com/help)
- **Railway Config-as-Code**: [Railway Docs](https://docs.railway.com/guides/config-as-code)
- **This Configuration**: Create an issue in this repository