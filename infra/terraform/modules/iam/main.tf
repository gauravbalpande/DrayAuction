# ─────────────────────────────────────────────────────────────────────────────
# Module: iam
# Creates IAM roles and instance profiles for EC2 (ECR pull, SSM access)
# ─────────────────────────────────────────────────────────────────────────────

variable "project" {}
variable "environment" {}
variable "aws_region" {}
variable "aws_account_id" {}

# ── EC2 → ECR + SSM role ─────────────────────────────────────────────────────

resource "aws_iam_role" "ec2" {
  name = "${var.project}-${var.environment}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

# Allow EC2 to pull from ECR
resource "aws_iam_role_policy" "ecr_pull" {
  name = "${var.project}-${var.environment}-ecr-pull"
  role = aws_iam_role.ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:DescribeRepositories",
          "ecr:ListImages"
        ]
        Resource = "*"
      }
    ]
  })
}

# Allow EC2 to read SSM parameters (for secrets)
resource "aws_iam_role_policy" "ssm_read" {
  name = "${var.project}-${var.environment}-ssm-read"
  role = aws_iam_role.ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:GetParametersByPath"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${var.aws_account_id}:parameter/${var.project}/*"
      }
    ]
  })
}

# Attach AWS managed SSM policy (for Session Manager / SSH alternative)
resource "aws_iam_role_policy_attachment" "ssm_managed" {
  role       = aws_iam_role.ec2.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

# ── Instance Profile ──────────────────────────────────────────────────────────

resource "aws_iam_instance_profile" "ec2" {
  name = "${var.project}-${var.environment}-ec2-profile"
  role = aws_iam_role.ec2.name
}

# ── GitHub Actions IAM User (for CI/CD ECR push) ──────────────────────────────

resource "aws_iam_user" "github_actions" {
  name = "${var.project}-github-actions"

  tags = {
    Purpose = "GitHub Actions CI/CD - ECR push"
  }
}

resource "aws_iam_user_policy" "github_actions_ecr" {
  name = "${var.project}-github-actions-ecr-push"
  user = aws_iam_user.github_actions.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage",
          "ecr:DescribeImages",
          "ecr:ListImages"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_access_key" "github_actions" {
  user = aws_iam_user.github_actions.name
}

# ── Outputs ───────────────────────────────────────────────────────────────────

output "ec2_instance_profile_name" {
  value = aws_iam_instance_profile.ec2.name
}

output "github_actions_access_key_id" {
  value     = aws_iam_access_key.github_actions.id
  sensitive = true
}

output "github_actions_secret_access_key" {
  value     = aws_iam_access_key.github_actions.secret
  sensitive = true
}
