variable "project_id" {
  description = "Railway project ID"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "backend_image" {
  description = "Backend Docker image"
  type        = string
}

variable "db_host" {
  description = "Database host"
  type        = string
}

variable "db_port" {
  description = "Database port"
  type        = string
}

variable "db_user" {
  description = "Database username"
  type        = string
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "Database name"
  type        = string
}