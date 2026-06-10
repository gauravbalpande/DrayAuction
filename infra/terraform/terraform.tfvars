# ─────────────────────────────────────────────────────────────────────────────
# terraform.tfvars — Production values (DO NOT commit secrets)
# Copy this to terraform.tfvars and fill in actual values.
# Secrets (JWT, DB passwords) are stored in AWS SSM Parameter Store, not here.
# ─────────────────────────────────────────────────────────────────────────────

aws_region             = "ap-south-1"
project                = "auctionxi"
environment            = "prod"
vpc_cidr               = "10.0.0.0/16"
ec2_instance_type      = "t3.medium"
ec2_key_name           = "auctionxi-prod" # Name of your AWS key pair (no .pem extension)
domain_name            = ""               # e.g. "auctionxi.yourdomain.com"
create_route53_records = false            # Set true if you have a Route53 hosted zone
