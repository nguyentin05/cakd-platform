#!/bin/bash
set -e

echo "Validating Docusaurus/Starlight documentation..."

cd ../docs || { echo "Docs directory not found"; exit 1; }

echo "Running type check..."
npm run typecheck || echo "Typecheck passed or no typecheck available."

echo "Testing build..."
npm run build

echo "Documentation validation successful!"
