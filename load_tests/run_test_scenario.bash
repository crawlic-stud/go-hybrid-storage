#!/bin/bash

# Function to execute a command
execute() {
  local cmd="$@"
  echo "running command: $cmd"
  eval "$cmd"
  local result=$?
  if [ $result -ne 0 ]; then
    echo "command '$cmd' failed"
    return 1
  fi
  return 0
}

# Function to run a k6 test
run_test() {
  local test_name="$1"
  local backend="$2"
  # local cmd="k6 run load_tests/scripts/${test_name}.js --out json=load_tests/results/json/${test_name}_${backend}.json"
  local cmd="k6 run load_tests/scripts/${test_name}.js --out csv=load_tests/results/csv/${test_name}_${backend}.csv"
  execute "$cmd"
  return $?
}

# Function to run the test suite for a given backend
run_tests_suite() {
  local backend="$1"

  if ! [[ "$backend" =~ ^(sqlite|postgres|mongo|fs)$ ]]; then
    echo "Unknown backend: $backend"
    exit 1
  fi

  # Build and run project
  tests=("upload_one_chunk")
  for test in "${tests[@]}"; do
    run_test "$test" "$backend" || exit 1
  done

  # Cleanup if backend == fs
  if [ "$backend" == "fs" ]; then
    execute "rm -rf files"
    mkdir files
  fi
}

# Main script
backend=("$1")

echo "Running tests for backend: $backend"
run_tests_suite "$backend"
echo "================ backend $backend done! ================"
