terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Remote state in S3 — bucket created manually once via bootstrap/
  backend "s3" {
    bucket         = "auctionxi-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "ap-south-1"
    encrypt        = true
    dynamodb_table = "auctionxi-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "AuctionXI"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# ─────────────────────────────────────────────
# Modules
# ─────────────────────────────────────────────

module "networking" {
  source      = "./modules/networking"
  project     = var.project
  environment = var.environment
  aws_region  = var.aws_region
  vpc_cidr    = var.vpc_cidr
}

module "ecr" {
  source      = "./modules/ecr"
  project     = var.project
  environment = var.environment
}

module "iam" {
  source         = "./modules/iam"
  project        = var.project
  environment    = var.environment
  aws_region     = var.aws_region
  aws_account_id = data.aws_caller_identity.current.account_id
}

module "ec2" {
  source                = "./modules/ec2"
  project               = var.project
  environment           = var.environment
  aws_region            = var.aws_region
  vpc_id                = module.networking.vpc_id
  subnet_id             = module.networking.public_subnet_id
  security_group_id     = module.networking.ec2_security_group_id
  instance_type         = var.ec2_instance_type
  key_name              = var.ec2_key_name
  instance_profile_name = module.iam.ec2_instance_profile_name
  ecr_backend_url       = module.ecr.backend_repository_url
  ecr_frontend_url      = module.ecr.frontend_repository_url
  domain_name           = var.domain_name
}

module "route53" {
  source                 = "./modules/route53"
  domain_name            = var.domain_name
  ec2_public_ip          = module.ec2.public_ip
  create_route53_records = var.create_route53_records
}

# ─────────────────────────────────────────────
# Data Sources
# ─────────────────────────────────────────────

data "aws_caller_identity" "current" {}

# ─────────────────────────────────────────────
# Outputs
# ─────────────────────────────────────────────

output "ec2_public_ip" {
  description = "Public IP address of the AuctionXI server"
  value       = module.ec2.public_ip
}

output "ec2_instance_id" {
  description = "EC2 instance ID"
  value       = module.ec2.instance_id
}

output "backend_ecr_url" {
  description = "ECR URL for backend image"
  value       = module.ecr.backend_repository_url
}

output "frontend_ecr_url" {
  description = "ECR URL for frontend image"
  value       = module.ecr.frontend_repository_url
}

output "ssh_command" {
  description = "SSH command to connect to the server"
  value       = "ssh -i ~/.ssh/${var.ec2_key_name}.pem ubuntu@${module.ec2.public_ip}"
}

output "github_actions_access_key_id" {
  description = "GitHub Actions Access Key ID"
  value       = module.iam.github_actions_access_key_id
  sensitive   = true
}

output "github_actions_secret_access_key" {
  description = "GitHub Actions Secret Access Key"
  value       = module.iam.github_actions_secret_access_key
  sensitive   = true
}
