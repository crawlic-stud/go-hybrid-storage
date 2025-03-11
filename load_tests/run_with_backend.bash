#!/bin/bash

# Check if an argument is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <argument>"
    exit 1
fi

argument="$1"

# Function to run the backend
run_backend() {
    go build & ./hybrid-storage "$argument"
}

# Function to run the tests
run_tests() {
    echo "Running tests..."
    bash load_tests/run_test_scenario.bash "$argument"
    echo "Tests finished."
}

# Run the backend in the background
run_backend 2> /dev/null &

# Run the tests
seconds=5
echo "Waiting $seconds seconds for the backend to start..."
sleep $seconds
run_tests 


backend_pid=$(ps aux | grep "./hybrid-storage" | grep -v grep | awk '{print $2}')

ps aux | grep "./hybrid-storage" | grep -v grep
echo "Killing the backend..."
kill -9 "$backend_pid"

echo "Script finished."

exit 0
