#!/bin/bash

# Navigate to the web/admin directory and run npm build
echo "Building the admin frontend..."
cd web/admin || { echo "web/admin directory not found"; exit 1; }
npm run build || { echo "npm build failed"; exit 1; }

# Navigate back to the root directory and build the Go app
echo "Building the Go application..."
cd ../../ || { echo "Failed to navigate to root directory"; exit 1; }
go build -o floe-cms || { echo "Go build failed"; exit 1; }

# Run the built application
echo "Running the application..."
./floe-cms || { echo "Failed to run the application"; exit 1; }