#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# AuctionXI EC2 User Data Script
# Runs on first boot to provision the k3s node.
# This script: installs k3s, ArgoCD, NGINX Ingress, cert-manager, Prometheus stack
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

PROJECT="${project}"
ENVIRONMENT="${environment}"
AWS_REGION="${aws_region}"
DOMAIN_NAME="${domain_name}"
ECR_BACKEND_URL="${ecr_backend_url}"
ECR_FRONTEND_URL="${ecr_frontend_url}"

LOG_FILE="/var/log/auctionxi-bootstrap.log"
exec > >(tee -a "$LOG_FILE") 2>&1

echo "=========================================================="
echo " AuctionXI Bootstrap Starting: $(date)"
echo "=========================================================="

# ── 1. System Updates ─────────────────────────────────────────────────────────
apt-get update -y
apt-get install -y \
  curl wget git unzip jq \
  apt-transport-https ca-certificates gnupg \
  awscli

# ── 2. Install k3s (single-node) ─────────────────────────────────────────────
echo ">>> Installing k3s..."
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.29.4+k3s1" sh -s - \
  --write-kubeconfig-mode 644 \
  --disable traefik \
  --tls-san "$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)"

# Wait for k3s to be ready
until kubectl get nodes 2>/dev/null | grep -q "Ready"; do
  echo "Waiting for k3s to be ready..."
  sleep 5
done
echo "k3s is ready!"

# Set KUBECONFIG globally
echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" >> /etc/environment
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# ── 3. Install Helm ───────────────────────────────────────────────────────────
echo ">>> Installing Helm..."
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# ── 4. ECR Authentication Helper ─────────────────────────────────────────────
echo ">>> Setting up ECR auth..."
ECR_REGISTRY=$(echo "$ECR_BACKEND_URL" | cut -d'/' -f1)

# Create ECR credential refresh script
cat > /usr/local/bin/ecr-refresh.sh << 'ECRSCRIPT'
#!/bin/bash
set -e
AWS_REGION="${aws_region}"
ECR_REGISTRY="${ecr_registry}"
TOKEN=$(aws ecr get-login-password --region "$AWS_REGION")
kubectl create secret docker-registry ecr-secret \
  --docker-server="$ECR_REGISTRY" \
  --docker-username=AWS \
  --docker-password="$TOKEN" \
  --namespace=auctionxi \
  --dry-run=client -o yaml | kubectl apply -f -
ECRSCRIPT

sed -i "s|\${aws_region}|$AWS_REGION|g" /usr/local/bin/ecr-refresh.sh
sed -i "s|\${ecr_registry}|$ECR_REGISTRY|g" /usr/local/bin/ecr-refresh.sh
chmod +x /usr/local/bin/ecr-refresh.sh

# Refresh ECR token every 10 hours (tokens expire after 12h)
echo "0 */10 * * * root /usr/local/bin/ecr-refresh.sh >> /var/log/ecr-refresh.log 2>&1" \
  >> /etc/crontab

# ── 5. Install NGINX Ingress Controller ──────────────────────────────────────
echo ">>> Installing NGINX Ingress..."
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.hostNetwork=true \
  --set controller.kind=DaemonSet \
  --wait --timeout 5m
   # --set controller.service.type=NodePort \
  # --set controller.service.nodePorts.http=80 \
  # --set controller.service.nodePorts.https=443 \

# ── 6. Install cert-manager ───────────────────────────────────────────────────
echo ">>> Installing cert-manager..."
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --wait --timeout 5m

# ── 7. Install ArgoCD ─────────────────────────────────────────────────────────
echo ">>> Installing ArgoCD..."
kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -n argocd \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for ArgoCD to be ready
kubectl wait --for=condition=available deployment/argocd-server \
  -n argocd --timeout=300s

# Patch ArgoCD server to use NodePort (for kubectl port-forward access)
kubectl patch svc argocd-server -n argocd \
  -p '{"spec":{"type":"NodePort","ports":[{"port":80,"targetPort":8080,"nodePort":30080}]}}'

# Get initial admin password
ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d)
echo "ArgoCD initial admin password: $ARGOCD_PASSWORD" | tee -a /root/argocd-credentials.txt

# ── 8. Install kube-prometheus-stack (Prometheus + Grafana + Alertmanager) ────
echo ">>> Installing kube-prometheus-stack..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=changeme-set-in-argo \
  --set prometheus.prometheusSpec.retention=7d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=5Gi \
  --set grafana.persistence.enabled=true \
  --set grafana.persistence.size=1Gi \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=1Gi \
  --wait --timeout 10m

# ── 9. Install Loki + Promtail (log aggregation) ──────────────────────────────
echo ">>> Installing Loki stack..."
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm upgrade --install loki-stack grafana/loki-stack \
  --namespace monitoring \
  --set loki.persistence.enabled=true \
  --set loki.persistence.size=5Gi \
  --set promtail.enabled=true \
  --wait --timeout 10m

# ── 10. Create auctionxi namespace ───────────────────────────────────────────
echo ">>> Creating auctionxi namespace..."
kubectl create namespace auctionxi --dry-run=client -o yaml | kubectl apply -f -

# Populate ECR secret immediately
/usr/local/bin/ecr-refresh.sh || true

# ── 11. Store useful info ─────────────────────────────────────────────────────
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
cat > /root/cluster-info.txt << EOF
========================================================
AuctionXI Cluster Info
========================================================
Public IP:      $PUBLIC_IP
k3s kubeconfig: /etc/rancher/k3s/k3s.yaml
ArgoCD UI:      http://$PUBLIC_IP:30080  (or kubectl port-forward)
ArgoCD pass:    $ARGOCD_PASSWORD
Grafana UI:     kubectl port-forward svc/kube-prometheus-stack-grafana -n monitoring 3001:80
ECR Backend:    $ECR_BACKEND_URL
ECR Frontend:   $ECR_FRONTEND_URL
Domain:         $DOMAIN_NAME
========================================================
EOF

cat /root/cluster-info.txt

echo "=========================================================="
echo " Bootstrap Complete: $(date)"
echo "=========================================================="
