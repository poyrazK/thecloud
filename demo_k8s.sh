#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
CLOUD_CLI="./bin/cloud"
VPC_NAME="demo-vpc"
CLUSTER_NAME="demo-cluster"
USER_EMAIL="demo@example.com"
USER_PASS="DemoPassword123!"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Kubernetes as a Service Demo ===${NC}"

# Check server
echo -e "\n${GREEN}[1] Checking Server Availability...${NC}"
if ! curl -s $API_URL/health > /dev/null; then
    echo "Server is not running at $API_URL."
    echo "Please run 'make run' in another terminal."
    exit 1
fi
echo "Server is UP."

# Register/Login
echo -e "\n${GREEN}[2] Authentication...${NC}"
# Register via API directly
echo "Registering user..."
curl -s -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$USER_EMAIL\", \"password\": \"$USER_PASS\", \"name\": \"Demo User\"}" > /dev/null

# Login to get API Key
echo "Logging in..."
LOGIN_RESP=$(curl -s -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$USER_EMAIL\", \"password\": \"$USER_PASS\"}")

# Extract API Key (simple filtering, assuming valid JSON response)
API_KEY=$(echo $LOGIN_RESP | grep -o '"api_key":"[^"]*"' | cut -d '"' -f 4)

if [ -z "$API_KEY" ]; then
    echo "Login failed. Response: $LOGIN_RESP"
    exit 1
fi

echo "API Key obtained: ${API_KEY:0:10}..."

# Configure CLI with the key
$CLOUD_CLI auth login $API_KEY

# Create VPC
echo -e "\n${GREEN}[3] Creating VPC...${NC}"
EXISTING_VPC=$($CLOUD_CLI vpc list --json | grep "\"name\": \"$VPC_NAME\"")
if [ -z "$EXISTING_VPC" ]; then
    # vpc create takes name as argument
    $CLOUD_CLI vpc create $VPC_NAME --cidr-block 10.100.0.0/16
    sleep 2 # wait for async creation if applicable
    VPC_ID=$($CLOUD_CLI vpc list --json | grep -B 1 $VPC_NAME | grep "id" | cut -d '"' -f 4)
else
    echo "VPC $VPC_NAME already exists."
    # Extract ID from list
    # Assumes JSON output format: { "id": "uuid", "name": "demo-vpc" ... }
    VPC_ID=$($CLOUD_CLI vpc list --json | grep -B 1 $VPC_NAME | grep "id" | cut -d '"' -f 4)
fi
echo "Using VPC ID: $VPC_ID"

# Create Kubernetes Cluster
echo -e "\n${GREEN}[4] Creating Kubernetes Cluster...${NC}"
EXISTING_CLUSTER=$($CLOUD_CLI k8s list --json | grep "\"name\": \"$CLUSTER_NAME\"")
if [ -z "$EXISTING_CLUSTER" ]; then
    # k8s create uses flags
    $CLOUD_CLI k8s create --name $CLUSTER_NAME --vpc $VPC_ID --workers 2
else
    echo "Cluster $CLUSTER_NAME already exists."
fi

# List Status
echo -e "\n${GREEN}[5] Cluster Status...${NC}"
$CLOUD_CLI k8s list

# Get Kubeconfig
echo -e "\n${GREEN}[6] Retrieving Kubeconfig...${NC}"
CLUSTER_ID=$($CLOUD_CLI k8s list --json | grep -B 1 $CLUSTER_NAME | grep "id" | cut -d '"' -f 4)
if [ -n "$CLUSTER_ID" ]; then
    $CLOUD_CLI k8s kubeconfig $CLUSTER_ID > demo-kubeconfig.yaml
    echo "Kubeconfig saved to demo-kubeconfig.yaml"
    
    echo -e "\n${BLUE}Demo Complete!${NC}"
    echo "You can view the cluster config: cat demo-kubeconfig.yaml"
    echo "Note: The IPs in kubeconfig are internal to the cloud platform (192.168...)."
else
    echo "Could not find cluster ID."
fi
