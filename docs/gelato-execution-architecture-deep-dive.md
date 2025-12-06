# Gelato Web3 Functions Execution Architecture - Deep Dive

## Executive Summary

This document provides a detailed analysis of how Gelato's TypeScript Web3 Functions work, including code storage, execution environment, secrets management, and comparison with TriggerX's architecture.

**Key Finding:** Gelato uses a **serverless JavaScript runtime** (js-1.0) with **IPFS storage** for code, not GitHub repos in Docker. This is fundamentally different from TriggerX's Docker-based approach.

---

## Table of Contents

1. [How Gelato Stores & Executes Code](#1-how-gelato-stores--executes-code)
2. [Execution Environment Details](#2-execution-environment-details)
3. [Secrets Management](#3-secrets-management)
4. [Deployment Workflow](#4-deployment-workflow)
5. [TypeScript Function Structure](#5-typescript-function-structure)
6. [Comparison: Gelato vs TriggerX](#6-comparison-gelato-vs-triggerx)
7. [Recommendations for TriggerX](#7-recommendations-for-triggerx)

---

## 1. How Gelato Stores & Executes Code

### 1.1 Code Storage: IPFS (NOT GitHub Repos)

**Workflow:**

```
Developer writes TypeScript → Compiles locally → Uploads to IPFS → Gets CID
                                                        ↓
                                    Executor nodes fetch from IPFS using CID
```

**Key Points:**

- ❌ **NOT stored in GitHub repos** (only used for local development)
- ✅ **Stored on IPFS** (decentralized storage)
- ✅ **Immutable:** Once deployed, code can't change (new deployment = new CID)
- ✅ **Pinned by Gelato:** Gelato nodes pin the IPFS content for availability

**Example Deployment:**

```bash
# 1. Developer writes function in TypeScript
# File: web3-functions/my-function/index.ts

# 2. Test locally
npx w3f test web3-functions/my-function/index.ts --logs

# 3. Deploy to IPFS
npx w3f deploy web3-functions/my-function/index.ts

# Output:
✓ Web3Function deployed to ipfs.
✓ CID: QmVfDbGGN6qfPs5ocu2ZuzLdBsXpu7zdfPwh14LwFUHLnc
```

**What gets uploaded to IPFS:**

- Compiled JavaScript (transpiled from TypeScript)
- `schema.json` (runtime configuration)
- Dependencies bundled in (npm modules included in build)
- Total size: ~1.6MB typical

### 1.2 Execution: JavaScript Runtime (NOT Docker)

**Runtime Environment:**

| Property | Value | Notes |
|----------|-------|-------|
| **Runtime** | `js-1.0` | JavaScript execution environment |
| **Memory** | `128 MB` | Per execution |
| **Timeout** | `30 seconds` | Max execution time |
| **Isolation** | Stateless context | Fresh memory per execution |
| **Network** | Full HTTP access | Can call any API |

**Architecture:**

```
┌─────────────────────────────────────────────────────────┐
│  Gelato Executor Node                                    │
│                                                          │
│  ┌────────────────────────────────────────────────┐     │
│  │  JavaScript Runtime ("js-1.0")                 │     │
│  │                                                 │     │
│  │  ┌───────────────────────────────────────┐     │     │
│  │  │ Execution Context (Stateless)         │     │     │
│  │  │                                        │     │     │
│  │  │ - Fetch code from IPFS (CID)          │     │     │
│  │  │ - Load into fresh V8 isolate (likely) │     │     │
│  │  │ - Inject context:                     │     │     │
│  │  │   • userArgs                          │     │     │
│  │  │   • gelatoArgs                        │     │     │
│  │  │   • multiChainProvider (RPC)          │     │     │
│  │  │   • storage (key-value)               │     │     │
│  │  │   • secrets                           │     │     │
│  │  │ - Execute function                    │     │     │
│  │  │ - Return {canExec, callData}          │     │     │
│  │  │                                        │     │     │
│  │  └───────────────────────────────────────┘     │     │
│  │                                                 │     │
│  └────────────────────────────────────────────────┘     │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**What happens when a task executes:**

1. **Executor node receives task** (every N seconds, or on trigger)
2. **Fetches code from IPFS** using CID (cached after first fetch)
3. **Creates fresh execution context** (stateless, empty memory)
4. **Loads compiled JavaScript** into runtime
5. **Injects context objects**:
   - `userArgs`: User-provided parameters
   - `gelatoArgs`: Block number, timestamp, etc.
   - `multiChainProvider`: Ethers.js provider for RPC calls
   - `storage`: Persistent key-value store
   - `secrets`: Encrypted secrets
6. **Executes `Web3Function.onRun()`**
7. **Returns result**: `{canExec: boolean, callData?: [], message?: string}`
8. **Destroys context** (fresh state next execution)

**NOT using Docker:**

- ❌ No Docker containers per execution
- ❌ No container pooling
- ❌ No image building
- ✅ Likely using **V8 isolates** (like Cloudflare Workers, Deno Deploy)
- ✅ Much faster startup (<10ms vs Docker's ~500ms)

**Evidence it's NOT Docker:**

1. **30-second timeout** (Docker startup alone can take 200-500ms)
2. **128MB memory** (very small, optimized for isolates)
3. **"js-1.0" runtime** (specific JavaScript runtime, not OS-level container)
4. **Stateless fresh context** (isolates pattern, not container reuse)
5. **Similar to AWS Lambda / Cloudflare Workers** (which use Firecracker microVMs or V8 isolates)

---

## 2. Execution Environment Details

### 2.1 Runtime Configuration (`schema.json`)

Every Web3 Function requires a `schema.json`:

```json
{
  "web3FunctionVersion": "2.0.0",
  "runtime": "js-1.0",
  "memory": 128,
  "timeout": 30,
  "userArgs": {
    "currency": "string",
    "oracle": "string"
  }
}
```

**Fields:**

- `web3FunctionVersion`: SDK version (currently `2.0.0`)
- `runtime`: Execution environment (fixed: `js-1.0`)
- `memory`: RAM allocation in MB (fixed: `128`)
- `timeout`: Max execution time in seconds (fixed: `30`)
- `userArgs`: User-configurable parameters with types

**Constraints:**

- ❌ **Cannot change runtime** (only `js-1.0` supported)
- ❌ **Cannot increase memory** (fixed 128MB)
- ❌ **Cannot extend timeout** (max 30 seconds)
- ❌ **Limited arg types**: `string`, `number`, `boolean`, and arrays of these

### 2.2 Execution Context API

**Available in `Web3FunctionContext`:**

```typescript
import { Web3Function, Web3FunctionContext } from "@gelatonetwork/web3-functions-sdk";

Web3Function.onRun(async (context: Web3FunctionContext) => {
  // 1. User Arguments (from schema.json)
  const { currency, oracle } = context.userArgs;

  // 2. Gelato Arguments (automatic)
  const { blockTime, blockNumber, chainId } = context.gelatoArgs;

  // 3. Multi-Chain Provider (Ethers.js)
  const provider = context.multiChainProvider.default();
  const contract = new Contract(address, abi, provider);
  const price = await contract.getPrice();

  // 4. Storage (persistent key-value)
  const lastBlock = await context.storage.get("lastBlock") ?? "0";
  await context.storage.set("lastBlock", blockNumber.toString());

  // 5. Secrets (encrypted)
  const apiKey = await context.secrets.get("API_KEY");
  if (!apiKey) {
    return { canExec: false, message: "API_KEY not set" };
  }

  // 6. HTTP Requests (via ky or fetch)
  const response = await fetch(`https://api.coingecko.com/...?api_key=${apiKey}`);
  const data = await response.json();

  // 7. Return execution decision
  return {
    canExec: true,
    callData: [
      {
        to: targetContract,
        data: contractInterface.encodeFunctionData("updatePrice", [price])
      }
    ]
  };
});
```

### 2.3 Stateless Execution

**Key Characteristics:**

```
Execution 1: Fresh memory → Execute → Return → DESTROY context
               ↓
            Storage persists (key-value store)
               ↓
Execution 2: Fresh memory → Execute → Return → DESTROY context
               ↓
            Storage persists
               ↓
...
```

**What persists between executions:**

- ✅ `storage` key-value data
- ✅ On-chain state (via RPC calls)
- ✅ Task configuration

**What does NOT persist:**

- ❌ Variables
- ❌ File system
- ❌ Network connections
- ❌ Imported module state

**Storage API:**

```typescript
// All values MUST be strings
await storage.set("key", value.toString());
const value = await storage.get("key"); // Returns string | undefined

// Example: Track last execution block
const lastBlock = parseInt(await storage.get("lastBlock") ?? "0");
await storage.set("lastBlock", blockNumber.toString());
```

---

## 3. Secrets Management

### 3.1 How Secrets Work

**Architecture:**

```
┌──────────────────────────────────────────────────────────┐
│  Developer's Machine                                      │
│                                                           │
│  .env file (local development):                           │
│  API_KEY=sk_test_12345                                    │
│  PRIVATE_KEY=0xabc...                                     │
└──────────────────────────────────────────────────────────┘
                     │
                     ▼
         ┌────────────────────────┐
         │  Deploy to IPFS        │
         │  (code only, NO .env)  │
         └────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│  Gelato Platform (After Task Creation)                    │
│                                                           │
│  User uploads secrets via:                                │
│  1. Web UI (Task Secrets section)                         │
│  2. SDK: web3Function.secrets.set(secrets, taskId)        │
│                                                           │
│  Storage: Encrypted, task-specific                        │
└──────────────────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│  Gelato Executor Node                                     │
│                                                           │
│  Execution context:                                       │
│  - Secrets decrypted and injected                         │
│  - Accessible via context.secrets.get("KEY_NAME")         │
│  - NOT visible in IPFS code                               │
└──────────────────────────────────────────────────────────┘
```

### 3.2 Setting Secrets

**Method 1: Web UI**

1. Create task on https://app.gelato.network
2. Go to "Task Secrets" section
3. Add key-value pairs:
   - Key: `API_KEY`
   - Value: `sk_test_12345`

**Method 2: SDK (Programmatic)**

```typescript
import { AutomateSDK } from "@gelatonetwork/automate-sdk";

const automate = new AutomateSDK(chainId, signer);

// Create task
const { taskId } = await automate.createTask({
  name: "My Task",
  execAddress: targetContract,
  web3FunctionHash: cid, // From IPFS deployment
  web3FunctionArgs: { currency: "ETH" }
});

// Set secrets for this task
const secrets = {
  API_KEY: "sk_test_12345",
  PRIVATE_KEY: "0xabc..."
};

await automate.secrets.set(secrets, taskId);
```

### 3.3 Accessing Secrets in Code

```typescript
Web3Function.onRun(async (context) => {
  // Get secret (returns string | undefined)
  const apiKey = await context.secrets.get("API_KEY");

  // ALWAYS validate
  if (!apiKey) {
    return {
      canExec: false,
      message: "API_KEY not configured"
    };
  }

  // Use in API calls
  const response = await fetch(
    `https://api.example.com/data?key=${apiKey}`
  );

  // ...
});
```

### 3.4 Security Properties

| Property | Implementation |
|----------|----------------|
| **Storage** | Encrypted at rest on Gelato servers |
| **Transmission** | Encrypted in transit to executor nodes |
| **Access Control** | Task-specific (each task has own secrets) |
| **Visibility** | NOT visible in IPFS code |
| **Scope** | Only accessible during execution of that task |
| **Audit** | Gelato can technically access (trust required) |

**Threat Model:**

- ✅ **Protected from:** Public IPFS viewers, other task creators
- ⚠️ **NOT protected from:** Gelato team (centralized trust)
- ⚠️ **NOT protected from:** Malicious executor node operators (if they modify Gelato software)

**Comparison to TriggerX:**

TriggerX could improve this with:
- **Threshold encryption:** Secrets encrypted, require N/M executors to decrypt
- **On-chain secret management:** Use EIP-5564 stealth addresses or similar
- **TEE-based secrets:** Secrets only decrypted inside Trusted Execution Environment

---

## 4. Deployment Workflow

### 4.1 Complete Workflow

```
┌────────────────────────────────────────────────────────────┐
│  STEP 1: Local Development                                 │
│                                                             │
│  Developer writes:                                          │
│  - web3-functions/my-func/index.ts                          │
│  - web3-functions/my-func/schema.json                       │
│  - web3-functions/my-func/.env (for local testing)          │
└────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────────────┐
│  STEP 2: Local Testing                                     │
│                                                             │
│  $ npx w3f test web3-functions/my-func/index.ts --logs     │
│                                                             │
│  Output:                                                    │
│  ✓ Web3Function Build result: Success                      │
│  ✓ Web3Function Schema: Valid                              │
│  ✓ Web3Function User args: Valid                           │
│  ✓ Runtime stats:                                           │
│    - Duration: 1.2s                                         │
│    - Memory: 45MB / 128MB                                   │
│    - RPC calls: 3                                           │
│  ✓ Return value: { canExec: true, callData: [...] }        │
└────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────────────┐
│  STEP 3: Deploy to IPFS                                    │
│                                                             │
│  $ npx w3f deploy web3-functions/my-func/index.ts          │
│                                                             │
│  Process:                                                   │
│  1. Compile TypeScript → JavaScript                         │
│  2. Bundle npm dependencies                                 │
│  3. Include schema.json                                     │
│  4. Upload to IPFS                                          │
│  5. Pin on Gelato infrastructure                            │
│                                                             │
│  Output:                                                    │
│  ✓ Web3Function deployed to ipfs                           │
│  ✓ CID: QmVfDbGGN6qfPs5ocu2ZuzLdBsXpu7zdfPwh14LwFUHLnc      │
│                                                             │
│  File size: ~1.63 MB                                        │
└────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────────────┐
│  STEP 4: Create Task                                       │
│                                                             │
│  Method A: Web UI                                           │
│  - Go to https://app.gelato.network                         │
│  - Connect wallet                                           │
│  - "Create New Task"                                        │
│  - Paste CID                                                │
│  - Configure userArgs                                       │
│  - Set trigger (time-based, event-based, etc.)              │
│  - Set secrets                                              │
│                                                             │
│  Method B: SDK                                              │
│  const { taskId } = await automate.createBatchExecTask({   │
│    name: "My Automation",                                   │
│    web3FunctionHash: cid,                                   │
│    web3FunctionArgs: { currency: "ETH" },                   │
│    trigger: { type: "time", interval: 60 }                  │
│  });                                                        │
└────────────────────────────────────────────────────────────┘
                     │
                     ▼
┌────────────────────────────────────────────────────────────┐
│  STEP 5: Execution by Gelato Network                       │
│                                                             │
│  Every N seconds (or on trigger):                           │
│  1. Executor node checks if task should run                 │
│  2. Fetches code from IPFS (using CID)                      │
│  3. Creates fresh execution context                         │
│  4. Injects userArgs, gelatoArgs, secrets, storage          │
│  5. Runs Web3Function.onRun()                               │
│  6. If canExec: true, submits transaction on-chain          │
│  7. Charges fee from user's prepaid balance (1Balance)      │
└────────────────────────────────────────────────────────────┘
```

### 4.2 Code Immutability

**Once deployed to IPFS:**

- ✅ Code is **immutable** (CID is hash of content)
- ✅ **Cannot modify** existing deployment
- ✅ To update: Deploy new version → Get new CID → Update task

**Updating a task:**

```typescript
// Old task
const oldCID = "QmOldHash123...";
const { taskId } = await automate.createTask({
  web3FunctionHash: oldCID,
  // ...
});

// Later: Deploy updated function
$ npx w3f deploy web3-functions/my-func/index.ts
// Output: New CID: QmNewHash456...

// Update task to use new CID
await automate.updateTask(taskId, {
  web3FunctionHash: "QmNewHash456..."
});
```

---

## 5. TypeScript Function Structure

### 5.1 Minimal Example

```typescript
import {
  Web3Function,
  Web3FunctionContext,
} from "@gelatonetwork/web3-functions-sdk";

Web3Function.onRun(async (context: Web3FunctionContext) => {
  const { userArgs, gelatoArgs } = context;

  // Your logic here
  const shouldExecute = true; // Some condition

  if (!shouldExecute) {
    return { canExec: false };
  }

  // Encode transaction data
  const callData = {
    to: "0x1234...",
    data: "0xabcd..." // Encoded function call
  };

  return {
    canExec: true,
    callData: [callData]
  };
});
```

### 5.2 Complex Example: Price Oracle

```typescript
import {
  Web3Function,
  Web3FunctionContext,
} from "@gelatonetwork/web3-functions-sdk";
import { Contract } from "@ethersproject/contracts";
import ky from "ky"; // HTTP client

Web3Function.onRun(async (context: Web3FunctionContext) => {
  const { userArgs, storage, secrets, multiChainProvider } = context;

  // 1. Get configuration
  const oracleAddress = userArgs.oracle as string;
  const currency = userArgs.currency as string;

  // 2. Get API key from secrets
  const apiKey = await secrets.get("COINGECKO_API_KEY");
  if (!apiKey) {
    return { canExec: false, message: "Missing API key" };
  }

  // 3. Fetch price from API
  const api = ky.create({
    headers: { "X-API-Key": apiKey }
  });

  const priceData = await api.get(
    `https://api.coingecko.com/api/v3/simple/price?ids=${currency}&vs_currencies=usd`
  ).json<any>();

  const currentPrice = Math.floor(priceData[currency].usd * 1e8); // 8 decimals

  // 4. Get last updated price from storage
  const lastPriceStr = await storage.get("lastPrice");
  const lastPrice = lastPriceStr ? parseInt(lastPriceStr) : 0;

  // 5. Check if update needed (5% threshold)
  const priceDiff = Math.abs(currentPrice - lastPrice);
  const threshold = lastPrice * 0.05; // 5%

  if (priceDiff < threshold) {
    return {
      canExec: false,
      message: `Price change ${priceDiff} below threshold ${threshold}`
    };
  }

  // 6. Update storage
  await storage.set("lastPrice", currentPrice.toString());

  // 7. Encode transaction
  const provider = multiChainProvider.default();
  const oracle = new Contract(
    oracleAddress,
    ["function updatePrice(uint256 price) external"],
    provider
  );

  const callData = oracle.interface.encodeFunctionData("updatePrice", [
    currentPrice
  ]);

  return {
    canExec: true,
    callData: [
      {
        to: oracleAddress,
        data: callData
      }
    ],
    message: `Updating price from ${lastPrice} to ${currentPrice}`
  };
});
```

### 5.3 Available NPM Modules

**Officially supported:**

- ✅ `@ethersproject/*` (Ethers.js v5)
- ✅ `ky` (HTTP client)
- ✅ `@gelatonetwork/web3-functions-sdk`

**Others work but not guaranteed:**

- ⚠️ Any pure JavaScript module (no native bindings)
- ❌ Modules requiring file system access
- ❌ Modules requiring native dependencies (C++ addons)

**How dependencies are included:**

- Bundled during deployment (`npx w3f deploy`)
- Increases IPFS file size
- All dependencies loaded into 128MB memory limit

---

## 6. Comparison: Gelato vs TriggerX

### 6.1 Execution Environment

| Aspect | Gelato | TriggerX (Current) |
|--------|--------|-------------------|
| **Runtime** | JavaScript ("js-1.0") | Docker containers |
| **Languages** | TypeScript only | Go, Python, TypeScript, JavaScript |
| **Isolation** | V8 isolates (likely) | Docker sandbox |
| **Startup Time** | <10ms | 200-500ms (container) |
| **Memory** | 128MB fixed | 256MB (configurable) |
| **Timeout** | 30s fixed | Configurable |
| **Code Storage** | IPFS (CID) | URL (GitHub, IPFS, etc.) |
| **Deployment** | Compile → Upload IPFS → CID | User provides URL |
| **Dependencies** | Bundled in IPFS upload | Installed at runtime |

### 6.2 Script Execution Flow

**Gelato:**

```
CID stored in task → Executor fetches from IPFS → Load into js-1.0 runtime →
Execute → Return {canExec, callData} → Submit tx if canExec=true
```

**TriggerX:**

```
URL stored in task → Keeper downloads script → Execute in Docker container →
Parse output → Convert to ABI types → Pack calldata → Submit tx
```

### 6.3 Secrets Management

| Aspect | Gelato | TriggerX (Current) |
|--------|--------|--------------------|
| **Storage** | Gelato servers (encrypted) | ❌ Not implemented |
| **Access** | `context.secrets.get()` | ❌ N/A |
| **Scope** | Task-specific | ❌ N/A |
| **UI** | Web dashboard + SDK | ❌ N/A |
| **Security** | Centralized trust | ❌ N/A |

**TriggerX needs to implement:**

- Encrypted secrets storage (per job)
- API for setting/getting secrets
- Injection into Docker execution environment

### 6.4 State Persistence

| Aspect | Gelato | TriggerX |
|--------|--------|----------|
| **Persistent Storage** | ✅ Key-value store | ❌ Not implemented |
| **API** | `storage.get()` / `storage.set()` | ❌ N/A |
| **Data Type** | Strings only | ❌ N/A |
| **Use Case** | Track last execution, cache data | ❌ N/A |

**TriggerX could add:**

```go
// In Docker execution context
type ExecutionContext struct {
    JobID     uint256
    Storage   StorageAPI
    Secrets   SecretsAPI
    BlockNumber uint64
    Timestamp uint64
}

type StorageAPI interface {
    Get(key string) (string, error)
    Set(key string, value string) error
}
```

### 6.5 Code Deployment

**Gelato:**

```
Developer → Local dev → Test → Deploy to IPFS → Get CID → Create task with CID
                                      ↓
                        Code is immutable, versioned, decentralized
```

**TriggerX:**

```
Developer → Upload script to GitHub/IPFS/Server → Provide URL → Create job with URL
                                      ↓
                        Code can change if URL is mutable (e.g., GitHub main branch)
```

**Implications:**

| Aspect | Gelato (IPFS CID) | TriggerX (URL) |
|--------|------------------|----------------|
| **Immutability** | ✅ Guaranteed (CID = hash) | ⚠️ Depends on URL (GitHub main can change) |
| **Verification** | ✅ Easy (verify CID hash) | ⚠️ Hard (URL content can change) |
| **Censorship Resistance** | ✅ High (IPFS decentralized) | ⚠️ Depends (GitHub can ban) |
| **Versioning** | ✅ New version = new CID | ❌ URL unchanged, content changes |

**Recommendation for TriggerX:**

Support IPFS CIDs in addition to URLs:

```go
type ScriptSource struct {
    Type string // "url", "ipfs"
    Value string // URL or CID
}

// For IPFS
source := ScriptSource{
    Type: "ipfs",
    Value: "QmVfDbGGN6qfPs5ocu2ZuzLdBsXpu7zdfPwh14LwFUHLnc"
}

// Fetch script
if source.Type == "ipfs" {
    script = ipfsGateway.Get(source.Value)
} else {
    script = httpClient.Get(source.Value)
}
```

---

## 7. Recommendations for TriggerX

### 7.1 What to Adopt from Gelato

✅ **Recommended:**

1. **IPFS Support for Script Storage**
   - Allow users to deploy scripts to IPFS and provide CID
   - Immutable, verifiable, censorship-resistant
   - Can still support URLs for flexibility

2. **Secrets Management API**
   - Per-job encrypted secrets
   - Accessible via context in script execution
   - User-friendly UI/SDK for setting secrets

3. **Persistent Storage (Key-Value)**
   - Allow scripts to persist state between executions
   - Useful for tracking last block, caching data
   - Simple API: `get(key)`, `set(key, value)`

4. **Structured Return Format**
   - Instead of parsing arbitrary output, enforce structure:
     ```json
     {
       "shouldExecute": true,
       "params": [...],
       "message": "..."
     }
     ```

5. **Runtime Configuration (schema.json equivalent)**
   - Let users specify:
     - Max memory
     - Max timeout
     - Expected output format
     - Dependencies

### 7.2 What NOT to Adopt

❌ **Avoid:**

1. **Single Language Limitation**
   - Gelato: TypeScript only
   - TriggerX: Keep multi-language support (Go, Python, TS, JS)
   - **Advantage:** More flexible

2. **Fixed Resource Limits**
   - Gelato: 128MB, 30s (not configurable)
   - TriggerX: Make it configurable per task tier
   - **Advantage:** Support complex computations

3. **Centralized Secrets**
   - Gelato: Secrets on Gelato servers (trust required)
   - TriggerX: Use threshold encryption or TEE
   - **Advantage:** More decentralized security

4. **No Verification**
   - Gelato: Single executor, no consensus
   - TriggerX: Multi-validator verification
   - **Advantage:** True security (already covered in design)

### 7.3 Hybrid Architecture Proposal

**Best of both worlds:**

```
┌────────────────────────────────────────────────────────────┐
│  TriggerX Enhanced Architecture                            │
│                                                             │
│  Script Storage:                                            │
│  - Support IPFS CIDs (like Gelato) ✅                       │
│  - Support URLs (current) ✅                                │
│  - Support Arweave (permanent storage) ✅                   │
│                                                             │
│  Execution:                                                 │
│  - Docker containers (current) ✅                           │
│  - Multi-language: Go, Python, TS, JS ✅                    │
│  - Configurable resources (unlike Gelato) ✅                │
│                                                             │
│  Verification:                                              │
│  - Multi-validator re-execution ✅                          │
│  - Othentic consensus ✅                                    │
│  - Deterministic inputs ✅                                  │
│                                                             │
│  Features from Gelato:                                      │
│  - Secrets management ✅                                    │
│  - Persistent storage (key-value) ✅                        │
│  - Structured output format ✅                              │
│  - Runtime configuration ✅                                 │
│                                                             │
│  Security > Gelato:                                         │
│  - Multi-validator consensus (vs single executor) ✅        │
│  - EigenLayer security ✅                                   │
│  - Threshold secret encryption (optional) ✅                │
└────────────────────────────────────────────────────────────┘
```

### 7.4 Implementation Priority

**Phase 1 (MVP - 4-6 weeks):**

1. ✅ Structured output format enforcement
2. ✅ IPFS CID support for script storage
3. ⚠️ Basic secrets management (encrypted DB)

**Phase 2 (Production - 6-8 weeks):**

4. ✅ Persistent key-value storage for scripts
5. ✅ Runtime configuration (schema.json equivalent)
6. ✅ Secrets UI/SDK

**Phase 3 (Scale - 8-12 weeks):**

7. ✅ Threshold secret encryption
8. ✅ Arweave support (permanent storage)
9. ✅ Advanced storage (SQL-like queries)

---

## 8. Detailed Comparison Table

| Feature | Gelato | TriggerX (Current) | TriggerX (Proposed) |
|---------|--------|-------------------|---------------------|
| **Code Storage** | IPFS (CID) | URL | IPFS + URL + Arweave |
| **Languages** | TypeScript only | Go, Python, TS, JS | Go, Python, TS, JS |
| **Runtime** | js-1.0 (likely V8 isolates) | Docker containers | Docker containers |
| **Memory** | 128MB fixed | 256MB | Configurable (128-512MB) |
| **Timeout** | 30s fixed | Configurable | Configurable |
| **Startup** | <10ms | 200-500ms | 200-500ms (accept trade-off) |
| **Secrets** | ✅ Encrypted, task-specific | ❌ None | ✅ Threshold encrypted |
| **Storage** | ✅ Key-value | ❌ None | ✅ Key-value + SQL |
| **Output Format** | `{canExec, callData}` | Arbitrary (parsed) | Structured JSON |
| **Verification** | ❌ Single executor | ❌ Single keeper | ✅ Multi-validator |
| **Consensus** | ❌ None | ❌ None | ✅ Othentic BLS |
| **Cost** | Gas + 2% | Gas + TG | Gas + $0.004 |
| **Security** | Economic only | None (trust performer) | Cryptographic + Economic |
| **Decentralization** | ⚠️ Medium (IPFS good, secrets centralized) | ⚠️ Medium | ✅ High |

---

## 9. Example: How Gelato Function Maps to TriggerX

### Gelato TypeScript Function

```typescript
import { Web3Function, Web3FunctionContext } from "@gelatonetwork/web3-functions-sdk";
import { Contract } from "@ethersproject/contracts";

Web3Function.onRun(async (context: Web3FunctionContext) => {
  const { userArgs, storage, secrets, multiChainProvider } = context;

  // Get API key
  const apiKey = await secrets.get("API_KEY");
  if (!apiKey) return { canExec: false };

  // Fetch external data
  const response = await fetch(`https://api.example.com?key=${apiKey}`);
  const data = await response.json();

  // Check if execution needed
  const lastValue = await storage.get("lastValue") ?? "0";
  if (data.value === parseInt(lastValue)) {
    return { canExec: false };
  }

  // Update storage
  await storage.set("lastValue", data.value.toString());

  // Encode calldata
  const provider = multiChainProvider.default();
  const contract = new Contract(userArgs.target, ABI, provider);
  const callData = contract.interface.encodeFunctionData("update", [data.value]);

  return {
    canExec: true,
    callData: [{ to: userArgs.target, data: callData }]
  };
});
```

### TriggerX Equivalent (Proposed)

**Go Script:**

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strconv"
)

type Context struct {
    UserArgs    map[string]string
    Secrets     map[string]string
    Storage     StorageAPI
    BlockNumber uint64
    Timestamp   uint64
}

type Output struct {
    ShouldExecute bool          `json:"shouldExecute"`
    Params        []interface{} `json:"params"`
    Message       string        `json:"message"`
}

func main() {
    // Parse injected context
    context := getContext() // Injected by TriggerX

    // Get API key from secrets
    apiKey, exists := context.Secrets["API_KEY"]
    if !exists {
        output := Output{
            ShouldExecute: false,
            Message:       "API_KEY not set",
        }
        printJSON(output)
        return
    }

    // Fetch external data
    data := fetchAPI(apiKey)

    // Check storage
    lastValue := context.Storage.Get("lastValue")
    if lastValue == "" {
        lastValue = "0"
    }

    if data.Value == mustInt(lastValue) {
        output := Output{
            ShouldExecute: false,
            Message:       "No change in value",
        }
        printJSON(output)
        return
    }

    // Update storage
    context.Storage.Set("lastValue", strconv.Itoa(data.Value))

    // Return execution params
    output := Output{
        ShouldExecute: true,
        Params:        []interface{}{data.Value},
        Message:       fmt.Sprintf("Updating from %s to %d", lastValue, data.Value),
    }
    printJSON(output)
}

func printJSON(v interface{}) {
    json, _ := json.Marshal(v)
    fmt.Println(string(json))
}
```

**Python Script:**

```python
import json
import os
import requests

def main():
    # Get context (injected by TriggerX)
    context = get_context()

    # Get API key from secrets
    api_key = context['secrets'].get('API_KEY')
    if not api_key:
        return {
            'shouldExecute': False,
            'message': 'API_KEY not set'
        }

    # Fetch external data
    response = requests.get(f'https://api.example.com?key={api_key}')
    data = response.json()

    # Check storage
    last_value = context['storage'].get('lastValue', '0')

    if data['value'] == int(last_value):
        return {
            'shouldExecute': False,
            'message': 'No change in value'
        }

    # Update storage
    context['storage'].set('lastValue', str(data['value']))

    # Return execution params
    return {
        'shouldExecute': True,
        'params': [data['value']],
        'message': f"Updating from {last_value} to {data['value']}"
    }

if __name__ == '__main__':
    output = main()
    print(json.dumps(output))
```

---

## 10. Summary

### How Gelato Works (Simplified)

1. **Write TypeScript** function using Web3Functions SDK
2. **Deploy to IPFS** → Get CID (immutable hash)
3. **Create task** with CID on Gelato platform
4. **Set secrets** (API keys, etc.) via UI or SDK
5. **Executor nodes** fetch code from IPFS and run in `js-1.0` runtime
6. **If `canExec: true`** → Submit transaction on-chain
7. **Charge fee** from user's 1Balance (gas + 2%)

### Key Differences vs TriggerX

| Aspect | Gelato | TriggerX (Current) |
|--------|--------|-------------------|
| **Runtime** | JavaScript isolates | Docker containers |
| **Code storage** | IPFS (CID) | URL |
| **Verification** | ❌ None (single executor) | ❌ None (single keeper) |
| **Secrets** | ✅ Yes | ❌ No |
| **Storage** | ✅ Key-value | ❌ No |
| **Languages** | TypeScript only | Go, Python, TS, JS |

### TriggerX Should Add

1. ✅ **IPFS support** (immutable code)
2. ✅ **Secrets management** (per-job encrypted)
3. ✅ **Persistent storage** (key-value)
4. ✅ **Structured output** (JSON format)
5. ✅ **Multi-validator verification** (already in design)

### TriggerX Advantages

1. ✅ **Multi-language support** (Go, Python, TS, JS)
2. ✅ **Multi-validator consensus** (security)
3. ✅ **Configurable resources** (memory, timeout)
4. ✅ **EigenLayer security** (shared security)

---

**Document Version:** 1.0
**Date:** 2025-11-13
**Author:** TriggerX Research Team
**Status:** Final
