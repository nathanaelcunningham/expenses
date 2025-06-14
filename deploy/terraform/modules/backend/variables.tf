variable "project_id" {
  description = "Railway project ID"
  type        = string
}

variable "environment_id" {
  description = "Railway environment ID"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository URL"
  type        = string
}

variable "github_branch" {
  description = "GitHub branch to deploy"
  type        = string
  default     = "main"
}

variable "database_url" {
  description = "Database connection URL"
  type        = string
  sensitive   = true
}

variable "root_directory" {
  description = "Root directory for the backend service"
  type        = string
  default     = "backend"
}

variable "config_path" {
  description = "Path to Railway config file"
  type        = string
  default     = "backend/railway.toml"
}

variable "port" {
  description = "Port for the backend service"
  type        = number
  default     = 8080
}


variable "additional_env_vars" {
  description = "Additional environment variables for the backend service"
  type        = map(string)
  default     = {}
}

