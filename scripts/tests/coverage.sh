#!/bin/bash

set -e

if [ -z "$1" ]; then
    echo "No folder provided, running all tests"
    gocov test ./... > coverage.json || {
        echo "Some tests failed, but continuing with available coverage data..."
    }
else
    gocov test "$1" > coverage.json || {
        echo "Some tests failed, but continuing with available coverage data..."
    }
fi

gocov-html -r -t kit -cmax 90 coverage.json > scripts/tests/coverage-detailed.html 
gocov-html coverage.json > scripts/tests/coverage.html
rm coverage.json
