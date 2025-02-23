#!/bin/bash

set -e # Exit on first error

for file in */main.go; do
  echo "Checking $file..."

  # Run the Go file, suppress standard output but show errors
  if ! GOFLAGS="" go run "$file" >/dev/null; then
    echo "❌ Compilation failed: $file"
    exit 1
  fi
done

echo "✅ All exercises compiled successfully!"
