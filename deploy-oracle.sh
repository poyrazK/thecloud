#!/bin/bash

# Configuration
REGION="ams" # Change to your Oracle Cloud region
APP_NAME="thecloud"
IMAGE_NAME="YOUR_REGISTRY/thecloud-api:latest"

echo "🚀 Preparing deployment to Oracle Cloud (OKE)..."

# 1. Build Multi-Arch Image
echo "📦 Building multi-arch Docker image (AMD64 + ARM64)..."
docker buildx build --platform linux/amd64,linux/arm64 \
    -t $IMAGE_NAME \
    --push .

# 2. Setup Kubernetes context (Assumes user has configured oci-cli)
# echo "🔧 Switching to OKE context..."
# oci ce cluster create-kubeconfig --cluster-id <YOUR_CLUSTER_ID> --file $HOME/.kube/config --region $REGION --token-version 2.0.0

# 3. Apply Namespace
echo "📂 Applying namespace..."
kubectl apply -f k8s/oracle/namespace.yaml

# 4. Apply Config and Secrets
echo "🔐 Applying configuration and secrets..."
kubectl apply -f k8s/oracle/configmap.yaml
kubectl apply -f k8s/oracle/secrets.yaml

# 5. Deploy Database and Redis
echo "💾 Deploying storage layer (Postgres + PgBouncer)..."
kubectl apply -f k8s/oracle/postgres.yaml
# kubectl apply -f k8s/oracle/redis-cluster.yaml # If using cluster, otherwise use standalone
kubectl apply -f k8s/oracle/pgbouncer.yaml

# 6. Deploy Application
echo "🚢 Deploying application (API + Workers)..."
sed -i "s|YOUR_REGISTRY/thecloud-api:latest|$IMAGE_NAME|g" k8s/oracle/api-deployment.yaml
sed -i "s|YOUR_REGISTRY/thecloud-api:latest|$IMAGE_NAME|g" k8s/oracle/worker-deployment.yaml

kubectl apply -f k8s/oracle/api-deployment.yaml
kubectl apply -f k8s/oracle/worker-deployment.yaml
kubectl apply -f k8s/oracle/api-service.yaml
kubectl apply -f k8s/oracle/ingress.yaml

echo "✅ Deployment manifests applied!"
echo "📍 Run 'kubectl get svc -n thecloud' to find your LoadBalancer IP."
echo "🔗 Check Swagger docs at http://<LOADBALANCER_IP>/docs/index.html"
