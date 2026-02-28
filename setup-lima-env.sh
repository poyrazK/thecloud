#!/bin/bash
# setup-lima-env.sh - Run this INSIDE the Lima VM
set -e

# Change to the project directory
cd /Users/poyrazk/dev/Cloud/thecloud

echo "Starting background services via Docker Compose..."
docker compose up -d postgres redis jaeger powerdns

echo "Waiting for Postgres to be ready..."
# We use the service name from docker-compose
until docker compose exec postgres pg_isready -U cloud -d cloud; do
  echo "Still waiting for postgres..."
  sleep 2
done

echo "Running migrations..."
export DATABASE_URL=postgres://cloud:cloud@localhost:5433/cloud?sslmode=disable
go run cmd/api/main.go --migrate-only

echo "--------------------------------------------------------"
echo "Ready! To run The Cloud with the Libvirt backend:"
echo "--------------------------------------------------------"
echo "export COMPUTE_BACKEND=libvirt"
echo "export NETWORK_BACKEND=ovs"
echo "export DATABASE_URL=postgres://cloud:cloud@localhost:5433/cloud?sslmode=disable"
echo "export REDIS_URL=localhost:6379"
echo "export POWERDNS_API_URL=http://localhost:8081"
echo "go run cmd/api/main.go"
echo "--------------------------------------------------------"
