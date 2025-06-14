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

variable "backend_url" {
  description = "Backend service URL for API calls"
  type        = string
}

variable "root_directory" {
  description = "Root directory for the frontend service"
  type        = string
  default     = "frontend"
}

variable "additional_env_vars" {
  description = "Additional environment variables for the frontend service"
  type        = map(string)
  default     = {}
}