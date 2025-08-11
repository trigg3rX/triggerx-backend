#!/bin/bash

set -e

# update the script directory to be the project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# update the project root directory to be the grandparent of the script directory
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Change to project root directory
cd "$PROJECT_ROOT"

# Define all Go folders to test
GO_FOLDERS=(
    "./cmd/..."
    "./internal/..."
    "./pkg/..."
    "./cli/..."
)

# echo "Running tests for coverage in: $(pwd)"
# echo "Testing folders: ${GO_FOLDERS[*]}"

# Check if coverage.out already exists, if not generate it
# if [ ! -f "coverage.out" ]; then
    go test -coverprofile=coverage.out "${GO_FOLDERS[@]}" || {
    # go test --coverprofile=coverage.out ./pkg/... || {
        echo "Some tests failed, but continuing with available coverage data..."
    }
# fi

go tool cover -html=coverage.out -o coverage.html
# Get the total coverage percentage from the coverage.out file
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')

echo "Total coverage: ${TOTAL_COVERAGE}%"

# Create a simple coverage header that preserves original styling
cat > temp_coverage_header.html << EOF
<span>Total Test Coverage: ${TOTAL_COVERAGE}% | Generated on $(date)</span>
EOF

# Find the line number where the body content starts
BODY_START=$(grep -n ">covered</span>" coverage.html | cut -d: -f1)

# Insert our header right after the opening body tag
{
    # Copy everything up to and including the opening body tag
    head -n $BODY_START coverage.html
    # Add our coverage header
    cat temp_coverage_header.html
    # Copy the rest of the file starting from the line after the opening body tag
    tail -n +$((BODY_START + 1)) coverage.html
} > coverage_with_percentage.html

# Replace the original file
mv coverage_with_percentage.html coverage.html

# Clean up temporary file
rm temp_coverage_header.html
rm coverage.out

mv coverage.html "$PROJECT_ROOT"/scripts/tests/coverage.html
echo "Coverage report generated: coverage.html"
