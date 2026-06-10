# AuctionXI — Production Deployment Runbook

## Architecture Overview

```
Internet
   │
   ▼
AWS Route53 (optional)
   │ DNS A record → Elastic IP
   ▼
EC2 t3.medium (Ubuntu 22.04)
   │ Elastic IP — static public IP
   ├── Port 80  → k3s NGINX Ingress (HTTP → HTTPS redirect)
   ├── Port 443 → k3s NGINX Ingress (HTTPS/WSS)
   └── Port 6443 → k3s API Server (kubectl access)
         │
         ▼
   k3s Cluster (single-node)
   ├── Namespace: auctionxi
   │     ├── Deployment: frontend   (Next.js → :3000)
   │     ├── Deployment: backend    (Go/Gin → :8080)
   │     ├── StatefulSet: postgres  (PVC 10Gi)
   │     └── Deployment: redis      (PVC 2Gi)
   │
   ├── Namespace: argocd
   │     └── ArgoCD Server  ← watches GitHub infra/k8s/base
   │
   ├── Namespace: ingress-nginx
   │     └── NGINX Ingress Controller (DaemonSet, hostNetwork)
   │
   ├── Namespace: cert-manager
   │     └── cert-manager  (Let's Encrypt TLS)
   │
   └── Namespace: monitoring
         ├── Prometheus (7d retention, 5Gi PVC)
         ├── Grafana    (1Gi PVC)
         ├── Alertmanager
         └── Loki + Promtail

ECR (Elastic Container Registry)
   ├── auctionxi/backend:sha
   └── auctionxi/frontend:sha
```

## GitOps CI/CD Flow

```
Developer pushes to main
        │
        ▼
GitHub Actions — deploy-backend.yml / deploy-frontend.yml
        │
        ├── 1. Run tests / lint
        ├── 2. Build Docker image
        ├── 3. Push to ECR with git SHA tag
        ├── 4. Trivy security scan
        └── 5. Update image tag in infra/k8s/base/*.yaml
                        │
                        ▼
              ArgoCD detects Git change
                        │
                        ▼
              ArgoCD syncs to k3s cluster
                        │
                        ▼
              Rolling deployment (0 downtime)
```

---

## Pre-requisites

- AWS CLI configured (`aws configure`)
- Terraform >= 1.6 installed
- kubectl installed
- An AWS key pair created in ap-south-1
- Your GitHub repository cloned locally

---

## Step 1 — Bootstrap Terraform State Backend

Run once before anything else:

```bash
chmod +x infra/terraform/bootstrap/bootstrap.sh
./infra/terraform/bootstrap/bootstrap.sh
```

This creates:
- S3 bucket `auctionxi-terraform-state` (versioned, encrypted)
- DynamoDB table `auctionxi-terraform-locks`

---

## Step 2 — Configure Terraform Variables

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:

```hcl
ec2_key_name = "your-aws-keypair-name"   # The name of your key pair (no .pem)
domain_name  = ""                         # Optional: your domain
```

---

## Step 3 — Deploy Infrastructure

```bash
cd infra/terraform
terraform init
terraform plan
terraform apply
```

**Expected resources created:**
- 1× VPC, 1× Public Subnet, 1× IGW, 1× Route Table
- 1× EC2 t3.medium (Ubuntu 22.04) with Elastic IP
- 1× EBS 30 GiB root + 1× EBS 20 GiB data
- 2× ECR repositories (backend, frontend)
- IAM role + instance profile for EC2
- IAM user for GitHub Actions

**Note your outputs:**
```bash
terraform output ec2_public_ip      # Your server IP
terraform output backend_ecr_url    # ECR URL for backend
terraform output frontend_ecr_url   # ECR URL for frontend
terraform output ssh_command        # SSH command
```

**Wait for bootstrap to complete (~8-12 minutes):**
```bash
ssh -i ~/.ssh/your-key.pem ubuntu@<EC2_IP>
tail -f /var/log/auctionxi-bootstrap.log
```

---

## Step 4 — Get GitHub Actions AWS Credentials

```bash
cd infra/terraform
terraform output -raw github_actions_access_key_id
terraform output -raw github_actions_secret_access_key
```

Add these as GitHub repository secrets:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

Also add:
- `NEXT_PUBLIC_API_URL` → `http://<EC2_IP>/api/v1` (or HTTPS with domain)
- `NEXT_PUBLIC_WS_URL` → `ws://<EC2_IP>` (or WSS with domain)

---

## Step 5 — Configure Kubernetes Secrets

SSH into the EC2 instance:
```bash
ssh -i ~/.ssh/your-key.pem ubuntu@<EC2_IP>
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

Generate secrets:
```bash
# Generate strong passwords
DB_PASSWORD=$(openssl rand -hex 32)
JWT_ACCESS=$(openssl rand -hex 64)
JWT_REFRESH=$(openssl rand -hex 64)

# Create postgres secret
kubectl create secret generic postgres-secret \
  --from-literal=POSTGRES_USER=auctionxi \
  --from-literal=POSTGRES_PASSWORD="$DB_PASSWORD" \
  -n auctionxi

# Create backend secret
kubectl create secret generic backend-secret \
  --from-literal=DATABASE_URL="postgres://auctionxi:${DB_PASSWORD}@postgres:5432/auctionxi?sslmode=disable" \
  --from-literal=REDIS_URL="redis://redis:6379/0" \
  --from-literal=JWT_ACCESS_SECRET="$JWT_ACCESS" \
  --from-literal=JWT_REFRESH_SECRET="$JWT_REFRESH" \
  -n auctionxi
```

---

## Step 6 — Update ConfigMaps

Edit `infra/k8s/base/configmaps.yaml` and replace `YOUR_DOMAIN_OR_IP`:
```yaml
CORS_ORIGIN: "http://YOUR_EC2_IP"
NEXT_PUBLIC_API_URL: "http://YOUR_EC2_IP/api/v1"
NEXT_PUBLIC_WS_URL: "ws://YOUR_EC2_IP"
```

---

## Step 7 — Set Up ArgoCD

```bash
# On EC2 — get ArgoCD initial password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d && echo

# Port-forward to access ArgoCD UI locally
kubectl port-forward svc/argocd-server -n argocd 8080:80 &

# Or access via NodePort (already patched to 30080):
# http://<EC2_IP>:30080
```

Connect your GitHub repo to ArgoCD:
```bash
# Update argocd-app.yaml with your GitHub repo URL, then:
kubectl apply -f infra/k8s/argocd/argocd-app.yaml -n argocd
```

---

## Step 8 — Push Your First Images

On your local machine:

```bash
# Login to ECR
aws ecr get-login-password --region ap-south-1 | \
  docker login --username AWS --password-stdin <ECR_REGISTRY>

# Build and push backend
docker build -t <BACKEND_ECR_URL>:latest ./backend
docker push <BACKEND_ECR_URL>:latest

# Build and push frontend
docker build -t <FRONTEND_ECR_URL>:latest ./frontend \
  --build-arg NEXT_PUBLIC_API_URL=http://<EC2_IP>/api/v1
docker push <FRONTEND_ECR_URL>:latest
```

backend_ecr_url = "040342463442.dkr.ecr.ap-south-1.amazonaws.com/auctionxi/backend"
frontend_ecr_url = "040342463442.dkr.ecr.ap-south-1.amazonaws.com/auctionxi/frontend"


Update image tags in manifests:
```bash
# backend.yaml
sed -i '' "s|PLACEHOLDER_BACKEND_IMAGE|<BACKEND_ECR_URL>:latest|g" \
  infra/k8s/base/backend.yaml

# frontend.yaml
sed -i '' "s|PLACEHOLDER_FRONTEND_IMAGE|040342463442.dkr.ecr.ap-south-1.amazonaws.com/auctionxi/frontend:latest|g" \
  infra/k8s/base/frontend.yaml

git add infra/k8s/base/
git commit -m "chore: set initial ECR image URIs"
git push
```

ArgoCD will automatically sync and deploy everything.

---

## Step 9 — Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n auctionxi

# Check backend logs
kubectl logs -n auctionxi deploy/backend -f

# Check frontend logs
kubectl logs -n auctionxi deploy/frontend -f

# Check ingress
kubectl get ingress -n auctionxi

# Test the app
curl http://<EC2_IP>/health
curl http://<EC2_IP>/api/v1/

# Open in browser
open http://<EC2_IP>
```

---

## Step 10 — Configure TLS (Optional, Recommended)

If you have a domain pointed at the EC2 Elastic IP:

1. Update `cert-issuer.yaml` with your email
2. Apply cert-manager issuers:
   ```bash
   kubectl apply -f infra/k8s/base/cert-issuer.yaml
   ```
3. Update `ingress.yaml` — uncomment TLS section and set your domain
4. Update `configmaps.yaml` — change URLs to `https://` and `wss://`
5. Push changes — ArgoCD will sync and cert-manager will auto-issue the certificate

---

## Monitoring & Observability

### Grafana

```bash
# Port-forward to local machine:
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3001:80

# Open: http://localhost:3001
# Username: admin
# Password: (set during bootstrap, or reset with kubectl)
```

**Pre-built dashboards:**
- Kubernetes Cluster Overview (ID: 315)
- NGINX Ingress (ID: 9614)
- Go application metrics (ID: 10826)
- Node Exporter Full (ID: 1860)

### Prometheus

```bash
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090
# Open: http://localhost:9090
```

### Loki (Logs)

In Grafana → Explore → select Loki data source:
```logql
{namespace="auctionxi", app="backend"} |= "auction"
{namespace="auctionxi"} |= "error"
```

### ArgoCD UI

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:80
# Open: http://localhost:8080
# Or: http://<EC2_IP>:30080 (NodePort)
```

---

## Cost Breakdown (~$20/month)

| Resource | Spec | Cost/month |
|----------|------|------------|
| EC2 t3.medium | 2 vCPU, 4 GiB RAM | ~$30 (with credits: ~$0) |
| EBS Root 30 GiB gp3 | OS + k3s images | ~$2.40 |
| EBS Data 20 GiB gp3 | Postgres + Redis | ~$1.60 |
| ECR Storage | ~2 GB | ~$0.20 |
| Data Transfer | ~5 GB out | ~$0.45 |
| Elastic IP (active) | Static IP | ~$0 |
| Route53 (optional) | Hosted zone | ~$0.50 |
| **Total** | | **~$5.15/month** |

> With AWS credits, total cost is effectively $0. When credits run out, total ~$35–40/month for t3.medium. Switch to t3.small (~$15/month) for production once credits deplete.

---

## Maintenance

### Upgrade an Image Manually

```bash
# On EC2:
kubectl set image deployment/backend backend=<NEW_IMAGE> -n auctionxi
kubectl set image deployment/frontend frontend=<NEW_IMAGE> -n auctionxi
kubectl rollout status deployment/backend -n auctionxi
```

### Rollback

```bash
kubectl rollout undo deployment/backend -n auctionxi
kubectl rollout history deployment/backend -n auctionxi
```

### Database Backup

```bash
# On EC2:
kubectl exec -n auctionxi statefulset/postgres -- \
  pg_dump -U auctionxi auctionxi > /backup/auctionxi-$(date +%Y%m%d).sql
```

### Scale Up

```bash
kubectl scale deployment/backend --replicas=2 -n auctionxi
kubectl scale deployment/frontend --replicas=2 -n auctionxi
```

### View All Resources

```bash
kubectl get all -n auctionxi
kubectl get all -n monitoring
kubectl top pods -n auctionxi
kubectl top nodes
```

---

## Troubleshooting

### Pod stuck in ImagePullBackOff

```bash
kubectl describe pod <pod-name> -n auctionxi
# Usually means ECR token expired — refresh it:
/usr/local/bin/ecr-refresh.sh
```

### Backend CrashLoopBackOff

```bash
kubectl logs -n auctionxi deploy/backend --previous
kubectl describe pod -n auctionxi -l app=backend
```

### Ingress not routing

```bash
kubectl get ingress -n auctionxi
kubectl describe ingress auctionxi-ingress -n auctionxi
kubectl logs -n ingress-nginx deploy/ingress-nginx-controller
```

### k3s node not Ready

```bash
kubectl get nodes
systemctl status k3s
journalctl -u k3s -f
```

---

## Multi-Stage Roadmap

| Stage | When | Change |
|-------|------|--------|
| Stage 1 (Now) | Launch | Single EC2 + k3s + in-cluster PG + Redis |
| Stage 2 | >1000 users | Add RDS PostgreSQL (Multi-AZ), ElastiCache Redis |
| Stage 3 | >10k users | Add second EC2 node to k3s, ALB, CloudFront CDN |
| Stage 4 | >100k users | Migrate to EKS, RDS Aurora, horizontal scaling |
