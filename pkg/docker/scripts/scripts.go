package scripts

import (
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
)


// getInitializationScript returns the script to initialize a container
func GetInitializationScript(language types.Language) string {
	switch language {
	case types.LanguageGo:
		return `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo 'package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}' > code.go
go mod init code
go mod download
echo "Container initialized successfully"
`
	case types.LanguagePy:
		return `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo 'print("Hello, World!")' > code.py
echo "Container initialized successfully"
`
	case types.LanguageJS, types.LanguageNode:
		return `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo 'console.log("Hello, World!");' > code.js
echo "Container initialized successfully"
`
	case types.LanguageTS:
		return `#!/bin/sh
set -e
mkdir -p /code
cd /code
npm install -g typescript
echo 'console.log("Hello, World!");' > code.ts
echo "Container initialized successfully"
`
	default:
		return `#!/bin/sh
set -e
mkdir -p /code
cd /code
echo "Container initialized successfully"
`
	}
}

func GetGoSetupScript() string {
	return `#!/bin/sh
set -e
cd /code

# One-time warm-up of Go build cache (if not already done)
if [ ! -f /code/.warm ]; then
    # Create a minimal Go program to warm up the compiler
    echo 'package main; func main(){}' > warm.go
    
    # Build with optimized flags to populate cache
    GOFLAGS='-buildvcs=false -trimpath' go build -o /tmp/warm warm.go
    
    # Clean up warm-up artifacts
    rm warm.go /tmp/warm
    touch /code/.warm
fi

echo "START_EXECUTION"
# Run with optimized flags
GOFLAGS='-buildvcs=false -trimpath' go run code.go 2>&1 || {
    echo "Error executing Go program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func GetPythonSetupScript() string {
	return `#!/bin/sh
set -e
cd /code

# One-time warm-up of Python bytecode cache
if [ ! -f /code/.warm ]; then
    # Create a minimal Python program to warm up common imports
    cat > warm.py << 'EOF'
import json
import os
import sys
import time
import datetime
import requests
import web3
EOF
    
    # Pre-compile to bytecode
    python -m py_compile warm.py
    # Run once to warm up imports
    python -c "import warm"
    rm warm.py warm.pyc
    touch /code/.warm
fi

echo "START_EXECUTION"
# Run with bytecode compilation enabled (default)
python -B code.py 2>&1 || {
    echo "Error executing Python program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func GetJavaScriptSetupScript() string {
	return `#!/bin/sh
set -e
cd /code

# One-time warm-up of V8 engine and common modules
if [ ! -f /code/.warm ]; then
    # Create a minimal JS program to warm up common modules
    cat > warm.js << 'EOF'
const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');
const crypto = require('crypto');
const { Web3 } = require('web3');
const { ethers } = require('ethers');
EOF
    
    # Run once to warm up V8 and modules
    NODE_OPTIONS='--no-warnings' node warm.js || true
    rm warm.js
    touch /code/.warm
fi

echo "START_EXECUTION"
# Run with optimized Node options
NODE_OPTIONS='--no-warnings --max-old-space-size=256' node code.js 2>&1 || {
    echo "Error executing JavaScript program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func GetTypeScriptSetupScript() string {
	return `#!/bin/sh
set -e
cd /code

# One-time warm-up of TypeScript compiler and V8 engine
if [ ! -f /code/.warm ]; then
    # Create a minimal TS program to warm up compiler
    cat > warm.ts << 'EOF'
import * as fs from 'fs';
import * as path from 'path';
import * as http from 'http';
import * as https from 'https';
import * as crypto from 'crypto';
import { Web3 } from 'web3';
import { ethers } from 'ethers';

interface WarmupTest {
    id: number;
    name: string;
}

const test: WarmupTest = { id: 1, name: "test" };
console.log(test);
EOF
    
    # Create tsconfig for faster compilation
    cat > tsconfig.json << 'EOF'
{
    "compilerOptions": {
        "target": "ES2020",
        "module": "commonjs",
        "strict": true,
        "esModuleInterop": true,
        "skipLibCheck": true,
        "forceConsistentCasingInFileNames": true,
        "outDir": "./dist",
        "incremental": true
    }
}
EOF
    
    # Warm up TypeScript compiler and V8
    NODE_OPTIONS='--no-warnings' tsc warm.ts && node warm.js || true
    rm warm.ts warm.js tsconfig.json
    touch /code/.warm
fi

echo "START_EXECUTION"
# Run with optimized options
NODE_OPTIONS='--no-warnings --max-old-space-size=256' tsc code.ts --incremental && node code.js 2>&1 || {
    echo "Error executing TypeScript program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}

func GetNodeSetupScript() string {
	return `#!/bin/sh
set -e
cd /code

# One-time warm-up of V8 engine and common modules
if [ ! -f /code/.warm ]; then
    # Create a minimal Node program to warm up common modules
    cat > warm.js << 'EOF'
const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');
const crypto = require('crypto');
const { Web3 } = require('web3');
const { ethers } = require('ethers');
EOF
    
    # Run once to warm up V8 and modules
    NODE_OPTIONS='--no-warnings' node warm.js || true
    rm warm.js
    touch /code/.warm
fi

echo "START_EXECUTION"
# Run with optimized Node options
NODE_OPTIONS='--no-warnings --max-old-space-size=256' node code.js 2>&1 || {
    echo "Error executing Node.js program. Exit code: $?"
    exit 1
}
echo "END_EXECUTION"
`
}