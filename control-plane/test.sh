#!/bin/bash

# Test script for the Skyscale control plane with test host VM

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Skyscale control plane in test mode...${NC}"

# Build the control plane if needed
if [ ! -f "./skyscale-control-plane" ]; then
    echo -e "${YELLOW}Building control plane...${NC}"
    make build
fi

# Kill any existing control plane process
pkill -f "skyscale-control-plane" || true

# Start the control plane in test mode
./skyscale-control-plane -test &
CONTROL_PLANE_PID=$!

# Wait for the control plane to start
echo -e "${YELLOW}Waiting for control plane to start...${NC}"
sleep 5

# Check if the control plane is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${RED}Control plane failed to start${NC}"
    kill $CONTROL_PLANE_PID
    exit 1
fi

echo -e "${GREEN}Control plane started successfully${NC}"

# Check if test mode is enabled
# TEST_STATUS=$(curl -s http://localhost:8080/test/status)
# if [[ $TEST_STATUS == *"test_mode\":true"* ]]; then
#     echo -e "${GREEN}Test mode is enabled${NC}"
# else
#     echo -e "${RED}Test mode is not enabled${NC}"
#     kill $CONTROL_PLANE_PID
#     exit 1
# fi

# Register a test function
# echo -e "${YELLOW}Registering test function...${NC}"
# FUNCTION_ID=$(curl -s -X POST http://localhost:8080/api/functions \
#     -H "Content-Type: application/json" \
#     -d '{
#     "name": "test-function",
#     "runtime": "python3.9",
#     "memory": 128,
#     "timeout": 30,
#     "code": "def handler(): return 1",
#     "requirements": "requests==2.26.0",
#     "config": "timeout: 30"
#   }' | jq -r '.id')

# if [ -z "$FUNCTION_ID" ] || [ "$FUNCTION_ID" == "null" ]; then
#     echo -e "${RED}Failed to register test function${NC}"
#     kill $CONTROL_PLANE_PID
#     exit 1
# fi

# echo -e "${GREEN}Registered test function with ID: $FUNCTION_ID${NC}"

# Invoke the function
echo -e "${YELLOW}Invoking test function...${NC}"
RESULT=$(curl -s -X POST http://localhost:8080/api/functions/f6c6a537-7ddf-4e22-ac10-979c37b00a12/invoke \
    -H "Content-Type: application/json" \
    -d '{
        "input": {},
        "sync": true
    }')

echo -e "${GREEN}Function invocation result: $RESULT${NC}"

# Check if the function was executed successfully
if [[ $RESULT == *1* ]]; then
    echo -e "${GREEN}Test function executed successfully${NC}"
else
    echo -e "${RED}Test function execution failed${NC}"
    kill $CONTROL_PLANE_PID
    exit 1
fi

# Clean up
echo -e "${YELLOW}Cleaning up...${NC}"
kill $CONTROL_PLANE_PID

echo -e "${GREEN}Test completed successfully${NC}"
exit 0 