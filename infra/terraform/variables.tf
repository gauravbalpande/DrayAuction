variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "ap-south-1"
}

variable "project" {
  description = "Project name prefix for all resources"
  type        = string
  default     = "auctionxi"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "prod"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "ec2_instance_type" {
  description = "EC2 instance type — t3.medium recommended for k3s"
  type        = string
  default     = "t3.medium"
}

variable "ec2_key_name" {
  description = "Name of the AWS key pair for SSH access"
  type        = string
}

variable "domain_name" {
  description = "Root domain name (e.g., auctionxi.com)"
  type        = string
  default     = ""
}

variable "create_route53_records" {
  description = "Whether to create Route53 DNS records (requires domain registered in Route53)"
  type        = bool
  default     = false
}
