# ─────────────────────────────────────────────────────────────────────────────
# Module: networking
# Creates VPC, public subnet, internet gateway, route table, security groups
# ─────────────────────────────────────────────────────────────────────────────

variable "project"     {}
variable "environment" {}
variable "aws_region"  {}
variable "vpc_cidr"    {}

# ── VPC ──────────────────────────────────────────────────────────────────────

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project}-${var.environment}-vpc"
  }
}

# ── Public Subnet ─────────────────────────────────────────────────────────────

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, 1)
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project}-${var.environment}-public-subnet"
  }
}

# ── Internet Gateway ──────────────────────────────────────────────────────────

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project}-${var.environment}-igw"
  }
}

# ── Route Table ───────────────────────────────────────────────────────────────

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project}-${var.environment}-public-rt"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# ── Security Group — EC2 / k3s node ──────────────────────────────────────────

resource "aws_security_group" "ec2" {
  name        = "${var.project}-${var.environment}-ec2-sg"
  description = "Security group for AuctionXI k3s node"
  vpc_id      = aws_vpc.main.id

  # SSH — restrict to your IP in production!
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Tighten to your IP: ["X.X.X.X/32"]
  }

  # HTTP (NGINX Ingress → redirects to HTTPS)
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # k3s API server (kubectl access)
  ingress {
    description = "k3s API"
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Tighten to your IP: ["X.X.X.X/32"]
  }

  # WebSocket (pass-through via NGINX Ingress on 443)
  # No separate port needed — WS upgrades handled by NGINX

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project}-${var.environment}-ec2-sg"
  }
}

# ── Outputs ───────────────────────────────────────────────────────────────────

output "vpc_id" {
  value = aws_vpc.main.id
}

output "public_subnet_id" {
  value = aws_subnet.public.id
}

output "ec2_security_group_id" {
  value = aws_security_group.ec2.id
}
