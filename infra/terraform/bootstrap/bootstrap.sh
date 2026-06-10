#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# Bootstrap: Create Terraform State Backend (S3 + DynamoDB)
# Run this ONCE before running terraform init.
# Usage: chmod +x bootstrap.sh && ./bootstrap.sh
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

BUCKET_NAME="auctionxi-terraform-state"
TABLE_NAME="auctionxi-terraform-locks"
REGION="ap-south-1"

echo ">>> Creating S3 bucket for Terraform state..."
aws s3api create-bucket \
  --bucket "$BUCKET_NAME" \
  --region "$REGION" \
  --create-bucket-configuration LocationConstraint="$REGION" 2>/dev/null || echo "Bucket already exists"

echo ">>> Enabling versioning on S3 bucket..."
aws s3api put-bucket-versioning \
  --bucket "$BUCKET_NAME" \
  --versioning-configuration Status=Enabled

echo ">>> Enabling encryption on S3 bucket..."
aws s3api put-bucket-encryption \
  --bucket "$BUCKET_NAME" \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

echo ">>> Blocking public access on S3 bucket..."
aws s3api put-public-access-block \
  --bucket "$BUCKET_NAME" \
  --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

echo ">>> Creating DynamoDB table for state locking..."
aws dynamodb create-table \
  --table-name "$TABLE_NAME" \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region "$REGION" 2>/dev/null || echo "DynamoDB table already exists"

echo ""
echo "✅ Terraform state backend ready!"
echo "   S3 bucket: $BUCKET_NAME"
echo "   DynamoDB:  $TABLE_NAME"
echo ""
echo "Now run: cd infra/terraform && terraform init"
