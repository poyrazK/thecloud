#!/bin/bash
set -e

# Configuration
API_PORT=8081
NODE_START_PORT=9101
DATA_DIR="./thecloud-data/storage-test"

# Cleanup function
cleanup() {
    echo "Stopping services..."
    pkill -P $$ || true
    kill $API_PID || true
    kill $NODE1_PID || true
    kill $NODE2_PID || true
    kill $NODE3_PID || true
    wait 2>/dev/null
}
trap cleanup EXIT

# 1. Build Binaries
echo "Building binaries..."
make build

# 2. Ensure Infrastructure
echo "Ensuring Database is up..."
docker compose up -d postgres redis
echo "Waiting for DB to be ready..."
sleep 5 # Simple wait, or check port

echo "Running migrations..."
export DATABASE_URL="postgres://cloud:cloud@localhost:5433/thecloud"
./bin/api --migrate-only

# 3. Start Storage Nodes
echo "Starting Storage Nodes..."
mkdir -p "$DATA_DIR/node-1" "$DATA_DIR/node-2" "$DATA_DIR/node-3"

./bin/storage-node --id node-1 --port $NODE_START_PORT --data-dir "$DATA_DIR/node-1" --peers "localhost:9102,localhost:9103" &
NODE1_PID=$!

./bin/storage-node --id node-2 --port $(($NODE_START_PORT+1)) --data-dir "$DATA_DIR/node-2" --peers "localhost:9101,localhost:9103" &
NODE2_PID=$!

./bin/storage-node --id node-3 --port $(($NODE_START_PORT+2)) --data-dir "$DATA_DIR/node-3" --peers "localhost:9101,localhost:9102" &
NODE3_PID=$!

sleep 2 # Wait for nodes to start

# 3. Start API Server
echo "Starting API Server..."
export PORT=$API_PORT
export OBJECT_STORAGE_MODE=distributed
export OBJECT_STORAGE_NODES="localhost:$NODE_START_PORT,localhost:$(($NODE_START_PORT+1)),localhost:$(($NODE_START_PORT+2))"
export APP_ENV=test
# Assuming DB connection envs are already present in .env or environment
# We load .env if it exists
if [ -f .env ]; then
    export $(cat .env | xargs)
fi
# Re-export overrides
export OBJECT_STORAGE_MODE=distributed
export OBJECT_STORAGE_NODES="localhost:$NODE_START_PORT,localhost:$(($NODE_START_PORT+1)),localhost:$(($NODE_START_PORT+2))"

./bin/api &
API_PID=$!

sleep 5 # Wait for API to start

# 4. Generate Auth Key
echo "Generating API Key..."
# We use the CLI to generate a key (requires API to be up? No, create-demo generates locally but likely saves to DB? 
# Wait, create-demo calls client.CreateKey which calls API. Yes.)
./bin/cloud auth create-demo test-user --api-url http://localhost:8081 > auth_output.txt
API_KEY=$(grep "Generated Key" auth_output.txt | awk '{print $NF}')
export CLOUD_API_KEY=$API_KEY
echo "Using API Key: $API_KEY"

# 5. Create Test File
echo "Creating test file..."
echo "Hello Distributed Storage!" > testfile.txt

# 6. Upload File
echo "Uploading file..."
./bin/cloud storage upload test-bucket testfile.txt --key my-object --api-url http://localhost:8081

# 7. Verification Loop
echo "Verifying data distribution..."
FOUND_COUNT=0
for i in 1 2 3; do
    if [ -f "$DATA_DIR/node-$i/test-bucket/my-object" ]; then
        echo "Found object on Node $i"
        FOUND_COUNT=$((FOUND_COUNT+1))
    fi
done

if [ $FOUND_COUNT -ge 2 ]; then
    echo "SUCCESS: Object replicated to $FOUND_COUNT nodes (Quorum met)"
else
    echo "FAILURE: Object found on only $FOUND_COUNT nodes"
    exit 1
fi

# 8. Download Verification
echo "Downloading file..."
./bin/cloud storage download test-bucket my-object downloaded.txt --api-url http://localhost:8081
if diff testfile.txt downloaded.txt; then
    echo "SUCCESS: Downloaded file matches original"
else
    echo "FAILURE: Downloaded file differs"
    exit 1
fi

echo "Test Complete!"
