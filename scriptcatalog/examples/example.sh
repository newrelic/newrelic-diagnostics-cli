#!/usr/bin/env bash

echo "This is just an example script! Command options passed in were $@"
echo "Creating 2 files to collect!"
timestamp=$(date +"%Y%m%d%H%M%S")
echo "Example example" > "example_$timestamp.log"
gzip -c "example_$timestamp.log" > "example_$timestamp.gz"