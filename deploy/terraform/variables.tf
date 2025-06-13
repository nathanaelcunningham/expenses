variable "railway_project_name" {
  description = "Name of the Railway project"
  type        = string
  default     = "expenses-app"
}

variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "Environment must be either 'dev' or 'prod'."
  }
}

variable "github_repo" {
  description = "GitHub repository URL"
  type        = string
  default     = "https://github.com/nathanaelcunningham/expenses"
}

variable "github_branch" {
  description = "GitHub branch to deploy from"
  type        = string
  default     = "main"
}

variable "db_user" {
  description = "Database username"
  type        = string
  default     = "expenses"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "expenses"
}