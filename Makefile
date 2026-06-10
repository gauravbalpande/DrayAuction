.PHONY: help dev up down build test migrate backend-test frontend-dev backend-dev clean \
        tf-init tf-plan tf-apply tf-destroy tf-output \
        k-status k-pods k-logs k-exec k-secrets \
        argocd-pw grafana argocd-ui \
        ecr-login push-backend push-frontend bootstrap-state

# ─────────────────────────────────────────────────────────────────────────────
# Variables
# ─────────────────────────────────────────────────────────────────────────────
AWS_REGION    ?= ap-south-1
NAMESPACE     ?= auctionxi
KUBECONFIG_PATH ?= /etc/rancher/k3s/k3s.yaml

help:
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║            AuctionXI — Available Commands                   ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║ LOCAL DEV                                                    ║"
	@echo "║   make dev           Start all services via Docker Compose  ║"
	@echo "║   make down          Stop all services                      ║"
	@echo "║   make build         Build Docker images                    ║"
	@echo "║   make backend-test  Run Go unit tests                      ║"
	@echo "║   make backend-dev   Run Go API locally                     ║"
	@echo "║   make frontend-dev  Run Next.js dev server                 ║"
	@echo "║                                                              ║"
	@echo "║ TERRAFORM                                                    ║"
	@echo "║   make bootstrap-state   Create S3 + DynamoDB backend       ║"
	@echo "║   make tf-init           terraform init                     ║"
	@echo "║   make tf-plan           terraform plan                     ║"
	@echo "║   make tf-apply          terraform apply                    ║"
	@echo "║   make tf-output         Show Terraform outputs             ║"
	@echo "║   make tf-destroy        Destroy all infra (CAREFUL!)       ║"
	@echo "║                                                              ║"
	@echo "║ KUBERNETES                                                   ║"
	@echo "║   make k-status      Show all pods/services/ingress         ║"
	@echo "║   make k-pods        Watch pods                             ║"
	@echo "║   make k-logs        Stream backend logs                    ║"
	@echo "║   make k-secrets     Create k8s secrets interactively       ║"
	@echo "║                                                              ║"
	@echo "║ MONITORING / ARGOCD                                         ║"
	@echo "║   make argocd-pw     Get ArgoCD admin password              ║"
	@echo "║   make argocd-ui     Port-forward ArgoCD UI → localhost:8080║"
	@echo "║   make grafana       Port-forward Grafana → localhost:3001  ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"
	@echo ""

# ─────────────────────────────────────────────────────────────────────────────
# LOCAL DEV
# ─────────────────────────────────────────────────────────────────────────────

dev up:
	docker compose up --build

down:
	docker compose down

build:
	docker compose build

backend-test:
	cd backend && go test ./...

backend-dev:
	cd backend && go run ./cmd/api

frontend-dev:
	cd frontend && npm run dev

migrate:
	cd backend && goose -dir migrations postgres "$${DATABASE_URL:-postgres://auctionxi:auctionxi@localhost:5432/auctionxi?sslmode=disable}" up

clean:
	rm -rf backend/bin frontend/.next frontend/node_modules

# ─────────────────────────────────────────────────────────────────────────────
# TERRAFORM
# ─────────────────────────────────────────────────────────────────────────────

bootstrap-state:
	chmod +x infra/terraform/bootstrap/bootstrap.sh
	./infra/terraform/bootstrap/bootstrap.sh

tf-init:
	cd infra/terraform && terraform init

tf-plan:
	cd infra/terraform && terraform plan

tf-apply:
	cd infra/terraform && terraform apply

tf-output:
	cd infra/terraform && terraform output

tf-destroy:
	@echo "⚠️  WARNING: This will destroy ALL infrastructure!"
	@echo "Press Ctrl-C to cancel, or wait 10 seconds..."
	@sleep 10
	cd infra/terraform && terraform destroy

# ─────────────────────────────────────────────────────────────────────────────
# KUBERNETES
# ─────────────────────────────────────────────────────────────────────────────

k-status:
	kubectl get pods,svc,ingress,pvc -n $(NAMESPACE)

k-pods:
	kubectl get pods -n $(NAMESPACE) -w

k-logs:
	kubectl logs -n $(NAMESPACE) deploy/backend -f

k-logs-frontend:
	kubectl logs -n $(NAMESPACE) deploy/frontend -f

k-exec:
	kubectl exec -it -n $(NAMESPACE) deploy/backend -- sh

k-secrets:
	@echo "Creating Kubernetes secrets interactively..."
	@read -p "Enter DB password: " DB_PASS && \
	read -p "Enter JWT Access Secret (or press enter to generate): " JWT_ACCESS && \
	read -p "Enter JWT Refresh Secret (or press enter to generate): " JWT_REFRESH && \
	JWT_ACCESS=$${JWT_ACCESS:-$$(openssl rand -hex 64)} && \
	JWT_REFRESH=$${JWT_REFRESH:-$$(openssl rand -hex 64)} && \
	kubectl create secret generic postgres-secret \
		--from-literal=POSTGRES_USER=auctionxi \
		--from-literal=POSTGRES_PASSWORD="$$DB_PASS" \
		-n $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f - && \
	kubectl create secret generic backend-secret \
		--from-literal=DATABASE_URL="postgres://auctionxi:$$DB_PASS@postgres:5432/auctionxi?sslmode=disable" \
		--from-literal=REDIS_URL="redis://redis:6379/0" \
		--from-literal=JWT_ACCESS_SECRET="$$JWT_ACCESS" \
		--from-literal=JWT_REFRESH_SECRET="$$JWT_REFRESH" \
		-n $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

# ─────────────────────────────────────────────────────────────────────────────
# MONITORING & ARGOCD
# ─────────────────────────────────────────────────────────────────────────────

argocd-pw:
	kubectl -n argocd get secret argocd-initial-admin-secret \
		-o jsonpath="{.data.password}" | base64 -d && echo

argocd-ui:
	@echo "Opening ArgoCD at http://localhost:8080 (admin / see above)"
	kubectl port-forward svc/argocd-server -n argocd 8080:80

grafana:
	@echo "Opening Grafana at http://localhost:3001 (admin / changeme)"
	kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3001:80

prometheus:
	@echo "Opening Prometheus at http://localhost:9090"
	kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090

# ─────────────────────────────────────────────────────────────────────────────
# ECR
# ─────────────────────────────────────────────────────────────────────────────

ecr-login:
	aws ecr get-login-password --region $(AWS_REGION) | \
		docker login --username AWS \
		--password-stdin $$(aws ecr describe-repositories \
			--query 'repositories[0].repositoryUri' \
			--output text --region $(AWS_REGION) | cut -d'/' -f1)
