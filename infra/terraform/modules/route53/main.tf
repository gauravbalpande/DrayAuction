# ─────────────────────────────────────────────────────────────────────────────
# Module: route53
# Creates A record pointing domain to EC2 Elastic IP (optional)
# ─────────────────────────────────────────────────────────────────────────────

variable "domain_name"             { default = "" }
variable "ec2_public_ip"           {}
variable "create_route53_records"  { default = false }

data "aws_route53_zone" "main" {
  count = var.create_route53_records && var.domain_name != "" ? 1 : 0
  name  = var.domain_name
}

resource "aws_route53_record" "root" {
  count   = var.create_route53_records && var.domain_name != "" ? 1 : 0
  zone_id = data.aws_route53_zone.main[0].zone_id
  name    = var.domain_name
  type    = "A"
  ttl     = 300
  records = [var.ec2_public_ip]
}

resource "aws_route53_record" "www" {
  count   = var.create_route53_records && var.domain_name != "" ? 1 : 0
  zone_id = data.aws_route53_zone.main[0].zone_id
  name    = "www.${var.domain_name}"
  type    = "A"
  ttl     = 300
  records = [var.ec2_public_ip]
}

resource "aws_route53_record" "api" {
  count   = var.create_route53_records && var.domain_name != "" ? 1 : 0
  zone_id = data.aws_route53_zone.main[0].zone_id
  name    = "api.${var.domain_name}"
  type    = "A"
  ttl     = 300
  records = [var.ec2_public_ip]
}
