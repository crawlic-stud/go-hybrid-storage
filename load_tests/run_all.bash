#!/bin/bash

for backend in "$@"; do
    bash load_tests/run_with_backend.bash "$backend"
done

echo "All tests finished."
echo "Plotting results..."

python load_tests/plot_results.py 
echo "Script finished."