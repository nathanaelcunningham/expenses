terraform {
  required_version = ">= 1.0"
  
  required_providers {
    railway = {
      source  = "terraform-community-providers/railway"
      version = "0.5.1"
    }
  }
}

provider "railway" {
  # Railway API token should be set via RAILWAY_TOKEN environment variable
}
