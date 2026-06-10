# ─────────────────────────────────────────────────────────────────────────────
# Module: ec2
# Creates the single-node EC2 instance running k3s
# ─────────────────────────────────────────────────────────────────────────────

variable "project"               {}
variable "environment"           {}
variable "aws_region"            {}
variable "vpc_id"                {}
variable "subnet_id"             {}
variable "security_group_id"     {}
variable "instance_type"         {}
variable "key_name"              {}
variable "instance_profile_name" {}
variable "ecr_backend_url"       {}
variable "ecr_frontend_url"      {}
variable "domain_name"           { default = "" }

# ── Ubuntu 22.04 LTS AMI (latest) ────────────────────────────────────────────

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# ── Elastic IP — static public IP ─────────────────────────────────────────────

resource "aws_eip" "ec2" {
  domain = "vpc"

  tags = {
    Name = "${var.project}-${var.environment}-eip"
  }
}

resource "aws_eip_association" "ec2" {
  instance_id   = aws_instance.main.id
  allocation_id = aws_eip.ec2.id
}

# ── EBS Volume for persistent data ───────────────────────────────────────────

resource "aws_ebs_volume" "data" {
  availability_zone = "${var.aws_region}a"
  size              = 20  # 20 GiB for PostgreSQL + Redis data
  type              = "gp3"
  encrypted         = true

  tags = {
    Name = "${var.project}-${var.environment}-data-volume"
  }
}

resource "aws_volume_attachment" "data" {
  device_name  = "/dev/sdf"
  volume_id    = aws_ebs_volume.data.id
  instance_id  = aws_instance.main.id
  force_detach = false
}

# ── User Data Script ──────────────────────────────────────────────────────────

locals {
  userdata = templatefile("${path.module}/userdata.sh", {
    project         = var.project
    environment     = var.environment
    aws_region      = var.aws_region
    domain_name     = var.domain_name
    ecr_backend_url = var.ecr_backend_url
    ecr_frontend_url = var.ecr_frontend_url
    // first change(1)
    ecr_registry = split("/", var.ecr_backend_url)[0]
  })
}

# ── EC2 Instance ──────────────────────────────────────────────────────────────

resource "aws_instance" "main" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = var.key_name
  subnet_id              = var.subnet_id
  vpc_security_group_ids = [var.security_group_id]
  iam_instance_profile   = var.instance_profile_name

  user_data = local.userdata

  root_block_device {
    volume_size           = 30   # 30 GiB root — k3s images, OS
    volume_type           = "gp3"
    encrypted             = true
    delete_on_termination = true

    tags = {
      Name = "${var.project}-${var.environment}-root"
    }
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"  # IMDSv2 (security best practice)
    http_put_response_hop_limit = 1
  }

  tags = {
    Name = "${var.project}-${var.environment}-k3s-node"
    Role = "k3s-control-plane"
  }

  lifecycle {
    ignore_changes = [user_data, ami]  # Don't recreate on AMI updates
  }
}

# ── Outputs ───────────────────────────────────────────────────────────────────

output "instance_id" {
  value = aws_instance.main.id
}

output "public_ip" {
  value = aws_eip.ec2.public_ip
}

output "private_ip" {
  value = aws_instance.main.private_ip
}

