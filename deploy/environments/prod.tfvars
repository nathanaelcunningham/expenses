environment = "prod"
railway_project_name = "expenses-app"

# Docker images (use :latest tag for production)
backend_image = "ghcr.io/nathanaelcunningham/expenses/backend:latest"
frontend_image = "ghcr.io/nathanaelcunningham/expenses/frontend:latest"

# Database configuration
db_user = "expenses"
db_name = "expenses"
# db_password should be set via TF_VAR_db_password environment variable