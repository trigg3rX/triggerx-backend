package scripts

import "github.com/trigg3rX/triggerx-backend/pkg/docker/types"

// GetInitializationScript returns the script to initialize a container.
func GetInitializationScript(language types.Language) string {
	switch language {
	case types.LanguageGo:
		return goInitializationScript
	case types.LanguagePy:
		return pythonInitializationScript
	case types.LanguageJS, types.LanguageNode:
		return javascriptInitializationScript
	case types.LanguageTS:
		return typescriptInitializationScript
	default:
		return ""
	}
}

// GetSetupScript returns the script to prepare a container for execution.
func GetSetupScript(language types.Language) string {
	switch language {
	case types.LanguageGo:
		return goSetupScript
	case types.LanguagePy:
		return pythonSetupScript
	case types.LanguageJS, types.LanguageNode:
		return javascriptSetupScript
	case types.LanguageTS:
		return typescriptSetupScript
	default:
		return ""
	}
}

// GetExecutionScript returns the script that runs the user's code.
func GetExecutionScript(language types.Language) string {
	switch language {
	case types.LanguageGo:
		return goExecutionScript
	case types.LanguagePy:
		return pythonExecutionScript
	case types.LanguageJS, types.LanguageNode:
		return javascriptExecutionScript
	case types.LanguageTS:
		return typescriptExecutionScript
	default:
		return ""
	}
}

// GetCleanupScript returns the script to clean a container after execution.
// It is now language-specific to handle different artifacts.
func GetCleanupScript(language types.Language) string {
	switch language {
	case types.LanguageGo:
		return goCleanupScript
	case types.LanguagePy:
		return pythonCleanupScript
	case types.LanguageJS, types.LanguageNode:
		return javascriptCleanupScript
	case types.LanguageTS:
		return typescriptCleanupScript
	default:
		return ""
	}
}

// --- INITIALIZATION SCRIPTS ---
const goInitializationScript = `#!/bin/sh

set -e
mkdir -p /code
cd /code
# A minimal hello world for a valid initial state.
echo 'package main; import "fmt"; func main() { fmt.Println("init") }' > code.go
go mod init code
echo "Go container initialized successfully"
`

const pythonInitializationScript = `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo 'print("init")' > code.py
echo "Python container initialized successfully"
`

const javascriptInitializationScript = `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo 'console.log("init");' > code.js
echo "JavaScript container initialized successfully"
`

const typescriptInitializationScript = `#!/bin/sh
set -e
mkdir -p /code
cd /code
# Install TypeScript globally in the container during init.
npm install -g typescript
echo 'console.log("init");' > code.ts
echo "TypeScript container initialized successfully"
`

// --- SETUP SCRIPTS (Warming up, dependency installation) ---

const goSetupScript = `#!/bin/sh
set -e
cd /code
# One-time warm-up of Go build cache.
if [ ! -f /code/.warm ]; then
    echo 'package main; func main(){}' > warm.go
    GOFLAGS='-buildvcs=false -trimpath' go build -o /tmp/warm warm.go
    rm warm.go /tmp/warm
    touch /code/.warm
fi
# This is a critical step for Go Modules. It downloads dependencies.
go mod tidy
`

const pythonSetupScript = `#!/bin/sh
set -e
cd /code
# One-time warm-up of Python bytecode cache.
if [ ! -f /code/.warm ]; then
    cat > warm.py << 'EOF'
import json, os, sys, time, datetime, requests, web3
EOF
    python -m py_compile warm.py
    python -c "import warm"
    rm warm.py __pycache__/warm.cpython-*.pyc
    touch /code/.warm
fi
# Install dependencies from requirements.txt if it exists.
if [ -f requirements.txt ]; then
    pip install -r requirements.txt
fi
`

const javascriptSetupScript = `#!/bin/sh
set -e
cd /code
# One-time warm-up of V8 engine.
if [ ! -f /code/.warm ]; then
    cat > warm.js << 'EOF'
require('fs'); require('path'); require('http'); require('https'); require('crypto');
EOF
    node warm.js || true
    rm warm.js
    touch /code/.warm
fi
# Install dependencies from package.json if it exists.
if [ -f package.json ]; then
    npm install
fi
`

const typescriptSetupScript = `#!/bin/sh
set -e
cd /code
# One-time warm-up of TypeScript compiler.
if [ ! -f /code/.warm ]; then
    echo 'const a: string = "warm";' > warm.ts
    tsc warm.ts
    rm warm.ts warm.js
    touch /code/.warm
fi
# Install dependencies from package.json if it exists.
if [ -f package.json ]; then
    npm install
fi
`

// --- EXECUTION SCRIPTS (Running the code) ---

const goExecutionScript = `#!/bin/sh
set -e
cd /code
# Execute the code. Logs go to stdout/stderr. The result is written to result.json.
GOFLAGS='-buildvcs=false -trimpath' go run code.go > result.json 2>&1
# Create a completion marker file
echo "done" > execution_complete.flag
`

const pythonExecutionScript = `#!/bin/sh
set -e
cd /code
# Execute with unbuffered output and redirect to result.json
python -u -B code.py > result.json 2>&1
# Create a completion marker file
echo "done" > execution_complete.flag
`

const javascriptExecutionScript = `#!/bin/sh
set -e
cd /code
V8_MEMORY_LIMIT=${V8_MEMORY_LIMIT:-256}
NODE_OPTIONS="--no-warnings --max-old-space-size=${V8_MEMORY_LIMIT}"
# Execute and redirect output to result.json
node code.js > result.json 2>&1
# Create a completion marker file
echo "done" > execution_complete.flag
`

const typescriptExecutionScript = `#!/bin/sh
set -e
cd /code
V8_MEMORY_LIMIT=${V8_MEMORY_LIMIT:-256}
# First, compile the user's code.
tsc code.ts --target ES2020 --module commonjs --esModuleInterop --skipLibCheck
# Then, execute the compiled JavaScript and redirect output to result.json
NODE_OPTIONS="--no-warnings --max-old-space-size=${V8_MEMORY_LIMIT}" node code.js > result.json 2>&1
# Create a completion marker file
echo "done" > execution_complete.flag
`

// --- CLEANUP SCRIPTS (Resetting the container state) ---

const goCleanupScript = `#!/bin/sh
cd /code
rm -f code.go
rm -f result.json
rm -f execution_complete.flag
rm -rf /tmp/go-build*
`

const pythonCleanupScript = `#!/bin/sh
cd /code
rm -f code.py
rm -f result.json
rm -f execution_complete.flag
rm -f requirements.txt
rm -rf __pycache__
`

const javascriptCleanupScript = `#!/bin/sh
cd /code
rm -f code.js
rm -f result.json
rm -f execution_complete.flag
rm -f package.json
rm -f package-lock.json
rm -rf node_modules
`

const typescriptCleanupScript = `#!/bin/sh
cd /code
rm -f code.ts
rm -f code.js
rm -f result.json
rm -f execution_complete.flag
rm -f package.json
rm -f package-lock.json
rm -f tsconfig.json
rm -rf node_modules dist
`
