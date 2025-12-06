# Per-Block Script Execution Feature - Design Document

## Executive Summary

This document outlines the design for TriggerX's per-block script execution feature, where user-provided scripts run on every block to determine:
1. **Whether** a contract call should execute
2. **What parameters** should be passed if execution is required

This feature enables dynamic, condition-based automation with verifiable execution across a decentralized keeper network.

---

## Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Proposed Architecture](#proposed-architecture)
3. [Verification Mechanism](#verification-mechanism)
4. [Performance & Latency Optimization](#performance--latency-optimization)
5. [Cost Model & Pricing](#cost-model--pricing)
6. [Security Considerations](#security-considerations)
7. [Implementation Roadmap](#implementation-roadmap)

---

## 1. Current Architecture Analysis

### Existing Capabilities

**✅ What TriggerX Already Has:**
- Docker-based sandboxed script execution (Go, Python, TypeScript, JavaScript)
- Dynamic argument parsing from script output
- ABI-based parameter conversion and validation
- Othentic AVS integration with performer/validator roles
- EigenLayer shared security
- TaskExecutionHub for L2 execution
- Complexity-based script cost calculation

**❌ Current Limitations:**
- Scripts execute **only when triggered** (time/event/condition)
- No per-block execution model
- No script result verification mechanism
- Single performer execution without validation
- No replay protection for script results

### Task Definition Types

| ID | Type | Arguments | Current Behavior |
|----|------|-----------|------------------|
| 1 | Time-based | Static | Execute at scheduled time |
| 2 | Time-based | Dynamic (script) | Execute at scheduled time, run script for args |
| 3 | Event-based | Static | Execute when event detected |
| 4 | Event-based | Dynamic (script) | Execute when event detected, run script for args |
| 5 | Condition-based | Static | Execute when condition met |
| 6 | Condition-based | Dynamic (script) | Execute when condition met, run script for args |

**New Task Types Needed:**

| ID | Type | Arguments | Proposed Behavior |
|----|------|-----------|-------------------|
| 7 | Block-based | Dynamic (script) | **Run script every block**, execute if script returns true |
| 8 | Block-based with API | Dynamic (script) | **Run script every block** with external API calls |

---

## 2. Proposed Architecture

### 2.1 High-Level Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         EVERY NEW BLOCK                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  PERFORMER KEEPER (Randomly Selected via Othentic)              │
│                                                                   │
│  1. Detects new block via RPC subscription                       │
│  2. Fetches all active block-based tasks                         │
│  3. For each task:                                                │
│     a. Execute script in Docker sandbox                          │
│     b. Parse script output:                                       │
│        - shouldExecute: bool                                      │
│        - params: []interface{}                                    │
│        - blockNumber: uint256                                     │
│        - timestamp: uint256                                       │
│     c. Generate execution proof:                                  │
│        - scriptHash: keccak256(scriptCode)                       │
│        - inputHash: keccak256(blockNumber, timestamp, state)     │
│        - outputHash: keccak256(shouldExecute, params)            │
│        - signature: sign(inputHash, outputHash)                  │
│  4. If shouldExecute == true:                                     │
│     - Submit transaction to TaskExecutionHub                      │
│     - Include execution proof in calldata                         │
│  5. Broadcast execution data to Othentic network                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  VALIDATOR KEEPERS (Multiple, via Othentic threshold)           │
│                                                                   │
│  1. Receive execution data from Othentic                         │
│  2. For each execution:                                           │
│     a. Re-execute script with SAME inputs:                       │
│        - Same blockNumber                                         │
│        - Same timestamp                                           │
│        - Same blockchain state (via archive node)                │
│     b. Compare outputs:                                           │
│        - Does shouldExecute match?                                │
│        - Do params match?                                         │
│        - Does outputHash match?                                   │
│     c. Verify on-chain transaction (if executed):                │
│        - Confirm tx included in block                             │
│        - Verify correct parameters used                           │
│        - Check timing within tolerance                            │
│  3. Submit attestation to Othentic:                               │
│     - APPROVE: Script output matches, tx valid                    │
│     - REJECT: Script output differs OR tx invalid                 │
│  4. If threshold rejects → slash performer                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  OTHENTIC CONSENSUS & SLASHING                                   │
│                                                                   │
│  - Aggregates validator attestations                              │
│  - If > threshold approve → performer rewarded                    │
│  - If > threshold reject → performer slashed                      │
│  - Custom slashing criteria:                                      │
│    • Wrong script output (determinism failure)                    │
│    • Missing execution (liveness failure)                         │
│    • Invalid transaction parameters                               │
│    • Outside block window                                          │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Script Output Format

Scripts **MUST** return JSON with this structure:

```json
{
  "shouldExecute": true,
  "params": [
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    1000000000000000000,
    "0xabcdef..."
  ],
  "metadata": {
    "blockNumber": 12345678,
    "timestamp": 1709876543,
    "reason": "Price threshold exceeded",
    "apiCalls": [
      {
        "url": "https://api.coingecko.com/...",
        "response": {...}
      }
    ]
  }
}
```

### 2.3 New Smart Contract Components

#### A. ScriptRegistry.sol (L2)

```solidity
contract ScriptRegistry {
    struct Script {
        bytes32 scriptHash;        // keccak256(scriptCode)
        string scriptUrl;          // IPFS/Arweave URL
        address owner;
        uint256 maxGasPerExecution;
        uint256 maxExecutionsPerBlock;
        bool isActive;
        uint256 createdAt;
    }

    mapping(uint256 => Script) public scripts;
    mapping(bytes32 => bool) public isScriptVerified;

    // Users register scripts here
    function registerScript(
        uint256 jobId,
        string calldata scriptUrl,
        bytes32 scriptHash,
        uint256 maxGasPerExecution
    ) external;

    // Returns script for block-based execution
    function getScriptForJob(uint256 jobId) external view returns (Script memory);
}
```

#### B. ExecutionProofVerifier.sol (L2)

```solidity
contract ExecutionProofVerifier {
    struct ExecutionProof {
        uint256 blockNumber;
        uint256 timestamp;
        bytes32 scriptHash;
        bytes32 inputHash;      // keccak256(blockNumber, timestamp, stateRoot)
        bytes32 outputHash;     // keccak256(shouldExecute, params)
        address performer;
        bytes signature;
    }

    // Verifies that execution proof is valid
    function verifyExecutionProof(
        ExecutionProof calldata proof,
        bool shouldExecute,
        bytes calldata params
    ) external view returns (bool);

    // Checks if proof has been used (replay protection)
    mapping(bytes32 => bool) public usedProofs;
}
```

#### C. Updated TaskExecutionHub.sol

Add new function:

```solidity
function executeBlockBasedFunction(
    uint256 jobId,
    uint256 tgAmount,
    address target,
    bytes calldata data,
    ExecutionProof calldata proof  // NEW
) external payable onlyKeeper nonReentrant {
    // 1. Verify proof hasn't been used
    require(!proofVerifier.usedProofs(proof.outputHash), "Proof already used");

    // 2. Verify proof is valid
    require(proofVerifier.verifyExecutionProof(proof, true, data), "Invalid proof");

    // 3. Verify block window (must be within X blocks)
    require(block.number - proof.blockNumber <= MAX_BLOCK_DELAY, "Proof too old");

    // 4. Execute as usual
    (uint256 chainId, ) = jobRegistry.unpackJobId(jobId);
    require(chainId == block.chainid, "Job is from a different chain");

    address jobOwner = jobRegistry.getJobOwner(jobId);
    require(jobOwner != address(0), "Job not found");

    triggerGasRegistry.deductTGBalance(jobOwner, tgAmount);

    _executeFunction(target, data);

    // 5. Mark proof as used
    proofVerifier.markProofUsed(proof.outputHash);
}
```

---

## 3. Verification Mechanism

### 3.1 Deterministic Execution Requirements

For validation to work, script execution **MUST be deterministic**:

**✅ Allowed:**
- Reading blockchain state (via RPC at specific block)
- Mathematical calculations
- String/data parsing
- Deterministic API calls (if pinned to specific data)

**❌ Not Allowed:**
- `time.Now()` / `Date.now()` (use block timestamp instead)
- Random number generation (without seed)
- Non-deterministic API calls (live price feeds without proof)
- File system access
- Network calls without verifiable responses

### 3.2 Script Validation Strategy

**Approach: Deterministic Re-Execution with State Pinning**

#### Performer Phase:
1. Execute script at block N
2. Capture inputs:
   - `blockNumber`: N
   - `timestamp`: Block N timestamp
   - `stateRoot`: State root at block N (optional, for complex state)
3. Capture outputs:
   - `shouldExecute`: bool
   - `params`: []interface{}
4. Sign: `signature = sign(keccak256(inputHash, outputHash))`
5. Broadcast to Othentic

#### Validator Phase:
1. Receive execution data from Othentic
2. Use **archive node** to query state at block N
3. Re-execute script with **exact same inputs**
4. Compare outputs:
   ```go
   performerOutput := proof.OutputHash
   validatorOutput := keccak256(shouldExecute, params)

   if performerOutput != validatorOutput {
       return AttestationReject
   }
   ```
5. Submit attestation to Othentic

### 3.3 Handling Non-Determinism: API Calls

**Problem:** External API calls return different data over time.

**Solution 1: TLS Proof (Recommended)**

Use **TLSNotary** or similar to create verifiable proofs of API responses:

```go
// Performer executes
apiResponse, tlsProof := tlsnotary.FetchWithProof("https://api.coingecko.com/...")

// Include in execution data
executionData := {
    "shouldExecute": true,
    "params": [...],
    "tlsProofs": [tlsProof]  // Validators can verify without re-calling API
}
```

**Validators** verify the TLS proof instead of re-calling the API.

**Solution 2: Oracle Integration**

- Use Chainlink, Pyth, or Chronicle oracles
- Script reads on-chain oracle data at specific block
- Deterministic because blockchain state is pinned

**Solution 3: Snapshot + IPFS**

- Performer uploads API response to IPFS
- Includes IPFS CID in execution proof
- Validators fetch from IPFS and verify hash

### 3.4 Comparison: TEE vs Re-Execution

| Aspect | TEE (Intel SGX, AWS Nitro) | Deterministic Re-Execution |
|--------|----------------------------|---------------------------|
| **Cost** | High ($$$) | Low ($) |
| **Complexity** | High | Medium |
| **Trust Assumption** | Trust TEE hardware | Trust consensus (Othentic) |
| **Verification Speed** | Fast (just verify attestation) | Slower (re-run script) |
| **Flexibility** | Limited (hardware constraints) | High (any script) |
| **Decentralization** | Low (limited TEE providers) | High (any keeper can validate) |

**Recommendation: Start with Deterministic Re-Execution**

- Lower cost and complexity
- Better fits Othentic's consensus model
- Can add TEE later as optional "fast verification" mode

### 3.5 EigenCompute Integration (Future)

**When to consider:**
- After proving the product-market fit
- If verification costs become prohibitive
- If users need <1s execution guarantees

**How it works:**
1. Upload script to EigenCompute as Docker image
2. Script runs in TEE with cryptographic proof
3. Proof is much smaller than re-execution
4. Validators just verify proof signature

**Cost:** Significantly higher (~10-100x)

---

## 4. Performance & Latency Optimization

### 4.1 Challenge: Per-Block Execution at Scale

**Ethereum L1:** ~12s block time → 5 executions/minute
**Arbitrum/Optimism:** ~0.25s block time → 240 executions/minute
**Polygon PoS:** ~2s block time → 30 executions/minute

**Problem:** If 1000 users have block-based tasks, keepers must execute 1000 scripts per block.

### 4.2 Optimization Strategies

#### Strategy 1: Task Batching

Group tasks by script similarity:

```go
type TaskBatch struct {
    ScriptHash bytes32
    Tasks      []Task
}

// Execute script once, reuse result for similar tasks
func (k *Keeper) ExecuteBatch(batch TaskBatch) {
    result := k.dockerExecutor.Execute(batch.ScriptHash)

    for _, task := range batch.Tasks {
        // Use same script output for all tasks with identical script
        k.processResult(task, result)
    }
}
```

#### Strategy 2: Parallel Execution

Use container pooling (already implemented):

```go
// Pre-warm N containers per language
pool := NewContainerPool(
    "go":     20 containers,
    "python": 20 containers,
    "js":     20 containers,
)

// Execute M scripts in parallel
var wg sync.WaitGroup
for _, task := range tasks {
    wg.Add(1)
    go func(t Task) {
        defer wg.Done()
        container := pool.Get(t.Language)
        result := container.Execute(t.Script)
        pool.Return(container)
    }(task)
}
wg.Wait()
```

#### Strategy 3: Lazy Evaluation

Don't execute every task every block:

```go
type Task struct {
    JobID           uint256
    Script          Script
    ExecutionCadence uint64  // Execute every N blocks
    LastExecutedBlock uint64
}

func (k *Keeper) ShouldExecuteThisBlock(task Task, currentBlock uint64) bool {
    return (currentBlock - task.LastExecutedBlock) >= task.ExecutionCadence
}
```

**Example:**
- User specifies: "Run every 10 blocks"
- Keeper only executes on blocks: 100, 110, 120, ...

#### Strategy 4: Priority Queue

Execute high-value tasks first:

```go
type PriorityTask struct {
    Task Task
    Priority float64  // Based on tgAmount, user tier, historical success
}

tasks := k.GetTasksForBlock(blockNumber)
sort.Slice(tasks, func(i, j int) bool {
    return tasks[i].Priority > tasks[j].Priority
})

// Execute top N tasks within block time
deadline := time.Now().Add(blockTime * 0.8)  // Use 80% of block time
for _, task := range tasks {
    if time.Now().After(deadline) {
        break  // Out of time, skip remaining
    }
    k.Execute(task)
}
```

#### Strategy 5: Execution Sharding

Divide tasks among keepers:

```go
// Each keeper handles tasks where: jobID % numKeepers == keeperIndex
func (k *Keeper) IsMyTask(jobID uint256) bool {
    keeperIndex := k.GetMyIndexInActiveSet()  // From Othentic
    numKeepers := k.GetActiveKeeperCount()
    return (jobID.Uint64() % numKeepers) == uint64(keeperIndex)
}
```

**Result:** If 10 keepers, each handles 10% of tasks.

### 4.3 Latency Targets

| Chain | Block Time | Target Execution Time | Max Tasks per Keeper |
|-------|-----------|---------------------|---------------------|
| Ethereum L1 | 12s | 10s | 100 |
| Arbitrum | 0.25s | 0.2s | 5 |
| Optimism | 2s | 1.5s | 25 |
| Polygon PoS | 2s | 1.5s | 25 |
| Base | 2s | 1.5s | 25 |

**Execution Time Budget per Script:** 50-100ms

**How to achieve:**
- Pre-warmed containers
- Script complexity limits
- Timeout enforcement (already have: `DockerExecutor.Execute(..., timeout)`)

### 4.4 Archive Node Requirements

Validators need archive nodes to re-execute scripts at historical blocks.

**Options:**
1. **QuickNode / Alchemy / Infura**: Archive node access (~$500-1000/month)
2. **Self-hosted**: Run own archive node (~$2000/month infra)
3. **Shared pool**: Keepers share archive node infrastructure

**Cost allocation:** Pass to users via pricing model (see Section 5).

---

## 5. Cost Model & Pricing

### 5.1 Cost Components

| Component | Description | Cost per Execution |
|-----------|-------------|-------------------|
| **Script Execution** | Docker container CPU/memory | $0.0001 - 0.001 |
| **Validation** | 3-5 validators re-execute | $0.0003 - 0.005 |
| **Archive Node** | Historical state queries | $0.0001 |
| **Gas** | On-chain transaction | Variable ($0.10 - 10.00) |
| **Keeper Reward** | Performer + Validators | 5-10% of total cost |
| **Protocol Fee** | TriggerX margin | 2-5% of total cost |

### 5.2 Complexity-Based Pricing

Already implemented in `file/validator.go`:

```go
complexity = sizeKB*0.05 +
             numLines*0.02 +
             functionCount*0.8 +
             importCount*0.3 +
             controlFlowCount*0.4 +
             nestingDepth*0.5
```

**Enhancement for Block-Based Tasks:**

```go
type BlockTaskPricing struct {
    BaseExecutionCost     float64  // Base cost per execution
    ComplexityMultiplier  float64  // From existing formula
    ValidationMultiplier  float64  // Cost of N validators
    FrequencyMultiplier   float64  // Based on blocks/execution
    GasCost              float64  // Estimated gas for tx
}

func CalculateExecutionCost(task BlockTask) float64 {
    baseCost := 0.001  // $0.001 per execution

    complexity := CalculateComplexity(task.Script)
    complexityMultiplier := 1 + (complexity / 100)

    // More validators = higher cost
    validationMultiplier := 1 + (task.ValidatorCount * 0.2)

    // Executing every block vs every 100 blocks
    frequencyMultiplier := 1.0
    if task.ExecutionCadence < 10 {
        frequencyMultiplier = 2.0  // Premium for high frequency
    }

    gasCost := EstimateGasCost(task.TargetChain, task.TargetFunction)

    totalCost := (baseCost * complexityMultiplier * validationMultiplier * frequencyMultiplier) + gasCost

    // Add keeper reward (7%) + protocol fee (3%)
    totalCost *= 1.10

    return totalCost
}
```

### 5.3 Pricing Tiers

#### Tier 1: Basic (Free Trial)
- Max 10 executions/day
- Max script complexity: 50
- 3 validators
- Execution cadence: Every 100 blocks minimum
- **Price:** Free (subsidized)

#### Tier 2: Standard
- Up to 1000 executions/day
- Max script complexity: 100
- 3 validators
- Execution cadence: Every 10 blocks minimum
- **Price:** $0.001 per execution + gas

#### Tier 3: Professional
- Up to 10,000 executions/day
- Max script complexity: 200
- 5 validators
- Execution cadence: Every block
- **Price:** $0.002 per execution + gas

#### Tier 4: Enterprise
- Unlimited executions
- Unlimited complexity
- 7+ validators
- Every block + priority execution
- **Price:** Custom (volume discounts)

### 5.4 Competitor Comparison

| Provider | Cost Model | Verification | Notes |
|----------|-----------|--------------|-------|
| **Gelato** | Gas + 2% fee | Single executor | No multi-validator |
| **Chainlink Automation** | Gas + flat fee (~$0.10) | Decentralized oracle network | Higher base cost |
| **TriggerX (Proposed)** | Gas + $0.001-0.002 + 10% | Othentic consensus (3-7 validators) | True verification |

**Competitive advantage:** Lower cost than Chainlink, more secure than Gelato.

### 5.5 TriggerGas Point (TG) Calculation

Users prepay with TG points. Need to calculate TG deduction per execution:

```solidity
function calculateTGCost(
    uint256 scriptComplexity,
    uint256 validatorCount,
    uint256 executionCadence,
    uint256 estimatedGas
) public pure returns (uint256 tgAmount) {
    // Base cost: 1 TG = $0.001
    uint256 baseCost = 1000; // 1000 TG = $1

    // Complexity multiplier (0-200 scale)
    uint256 complexityFactor = (scriptComplexity * baseCost) / 100;

    // Validator cost (each validator adds 20%)
    uint256 validatorFactor = validatorCount * 200; // 200 TG per validator

    // Frequency premium
    uint256 frequencyFactor = 0;
    if (executionCadence < 10) {
        frequencyFactor = baseCost; // Double cost for <10 blocks
    }

    // Gas cost in TG (convert estimated gas to TG)
    uint256 gasTG = (estimatedGas * tx.gasprice * ethPriceInUSD) / 1e18;

    tgAmount = baseCost + complexityFactor + validatorFactor + frequencyFactor + gasTG;

    // Add 10% margin (7% keeper reward + 3% protocol)
    tgAmount = (tgAmount * 110) / 100;
}
```

---

## 6. Security Considerations

### 6.1 Threat Model

| Threat | Mitigation |
|--------|-----------|
| **Malicious performer submits fake results** | Validators re-execute and slash if mismatch |
| **Non-deterministic scripts** | Reject scripts with random/time functions |
| **Replay attacks** | Track used execution proofs on-chain |
| **Timing attacks** | Enforce block window (MAX_BLOCK_DELAY) |
| **Resource exhaustion** | Docker limits, timeouts, complexity caps |
| **Collusion between performer + validators** | Othentic's BLS threshold signatures prevent <67% collusion |
| **Malicious scripts (DoS, exploits)** | Sandboxing, seccomp profiles, no network access |

### 6.2 Slashing Criteria

Extend Othentic's slashing to include:

```solidity
enum SlashingReason {
    DOUBLE_SIGN,           // Existing: Sign two different blocks
    FALSE_VOTE,            // Existing: Vote against consensus
    WRONG_SCRIPT_OUTPUT,   // NEW: Output doesn't match validators
    MISSING_EXECUTION,     // NEW: Fail to execute assigned task
    INVALID_PARAMS,        // NEW: Wrong params sent to contract
    LATE_EXECUTION         // NEW: Execute outside block window
}
```

**Slashing amounts:**
- WRONG_SCRIPT_OUTPUT: 10% of stake
- MISSING_EXECUTION: 5% of stake
- INVALID_PARAMS: 10% of stake
- LATE_EXECUTION: 2% of stake

### 6.3 Script Sandboxing Enhancements

Current sandboxing (from `container/manager.go`):
```go
Resources: &container.Resources{
    Memory:     256 * 1024 * 1024,  // 256MB
    NanoCPUs:   1000000000,         // 1 CPU
    PidsLimit:  ptr.Int64(100),
}
```

**Enhancements for block-based tasks:**

```go
// Stricter limits for per-block execution
Resources: &container.Resources{
    Memory:     128 * 1024 * 1024,  // 128MB (reduced)
    NanoCPUs:   500000000,          // 0.5 CPU (reduced)
    PidsLimit:  ptr.Int64(50),      // Fewer processes

    // NEW: Network rate limiting
    BlkioWeight: 10,                // Low I/O priority
}

// Disable network for most scripts (unless TLS proof enabled)
NetworkMode: "none"

// Add seccomp profile
SecurityOpt: []string{
    "seccomp=/path/to/restrictive-profile.json",
}
```

### 6.4 Rate Limiting

Prevent spam / DoS:

```go
type RateLimiter struct {
    MaxExecutionsPerBlock map[address]uint64
    MaxTotalExecutionsPerDay map[address]uint64
}

// Before executing
func (r *RateLimiter) CheckLimit(user address, jobID uint256) error {
    execsThisBlock := r.GetExecutionsThisBlock(user)
    if execsThisBlock >= r.MaxExecutionsPerBlock[user] {
        return errors.New("rate limit exceeded for this block")
    }

    execsToday := r.GetExecutionsToday(user)
    if execsToday >= r.MaxTotalExecutionsPerDay[user] {
        return errors.New("daily rate limit exceeded")
    }

    return nil
}
```

---

## 7. Implementation Roadmap

### Phase 1: MVP (4-6 weeks)

**Week 1-2: Core Infrastructure**
- [ ] Add Task Definition Type 7 (block-based dynamic)
- [ ] Implement block subscription in keeper (`internal/keeper/`)
- [ ] Update `TaskExecutor` to handle per-block execution
- [ ] Add script output format validation (JSON with `shouldExecute`, `params`)

**Week 3: Smart Contracts**
- [ ] Deploy `ScriptRegistry.sol` on Base Sepolia testnet
- [ ] Deploy `ExecutionProofVerifier.sol` on Base Sepolia testnet
- [ ] Update `TaskExecutionHub.sol` with `executeBlockBasedFunction`
- [ ] Add execution proof tracking

**Week 4: Validation**
- [ ] Implement deterministic re-execution in validators
- [ ] Add archive node integration (QuickNode/Alchemy)
- [ ] Create attestation submission to Othentic
- [ ] Test validation flow end-to-end

**Week 5-6: Testing & Optimization**
- [ ] Load testing: 100 concurrent block-based tasks
- [ ] Optimize container pooling for per-block execution
- [ ] Add monitoring dashboards (Grafana)
- [ ] Security audit (internal)

**Success Criteria:**
- ✅ Perform 10 block-based tasks on testnet with 3-validator consensus
- ✅ Validation accuracy >99%
- ✅ Execution latency <1s per task

### Phase 2: Production Beta (6-8 weeks)

**Week 7-8: Mainnet Preparation**
- [ ] Deploy contracts to Base, OP, Arbitrum mainnet
- [ ] Set up mainnet keeper infrastructure (3-5 performers, 10 validators)
- [ ] Implement TG point cost calculation
- [ ] Create user dashboard for block-based tasks

**Week 9-10: Advanced Features**
- [ ] Add TLS proof support for API calls (TLSNotary integration)
- [ ] Implement task batching optimization
- [ ] Add execution sharding among keepers
- [ ] Priority queue for high-value tasks

**Week 11-12: User Onboarding**
- [ ] Create documentation & tutorials
- [ ] Build script templates (price feeds, liquidation checks, etc.)
- [ ] Launch beta program (10 selected users)
- [ ] Collect feedback & iterate

**Success Criteria:**
- ✅ 10 beta users with 100+ block-based tasks
- ✅ 99.9% uptime
- ✅ No slashing incidents due to false positives

### Phase 3: Scale & Optimize (8-12 weeks)

**Week 13-16: Performance**
- [ ] Implement parallel execution (20+ containers per language)
- [ ] Add lazy evaluation (execute every N blocks)
- [ ] Optimize archive node caching
- [ ] Reduce validation latency to <500ms

**Week 17-20: Economic Security**
- [ ] Launch slashing mechanism for script output mismatches
- [ ] Implement dynamic pricing based on demand
- [ ] Add volume discounts for enterprise users
- [ ] Create insurance fund for user refunds

**Week 21-24: Advanced Verification**
- [ ] Evaluate EigenCompute integration
- [ ] Pilot TEE verification (AWS Nitro Enclaves)
- [ ] Add ZK proofs for script execution (future research)
- [ ] Multi-chain state queries (cross-chain scripts)

**Success Criteria:**
- ✅ 100+ active users
- ✅ 1000+ block-based tasks
- ✅ <$0.001 cost per execution
- ✅ 10ms p99 latency for script execution

### Phase 4: Enterprise & Ecosystem (Ongoing)

- [ ] Chainlink oracle integration
- [ ] Pyth price feed integration
- [ ] API marketplace (users share scripts)
- [ ] DAO governance for protocol parameters
- [ ] Cross-chain execution (execute on chain A based on chain B state)

---

## 8. Open Questions & Decisions Needed

### 8.1 Architecture Decisions

| Question | Options | Recommendation |
|----------|---------|----------------|
| **Verification method** | (A) Deterministic re-execution<br>(B) TEE<br>(C) ZK proofs | **A** for MVP, evaluate B in Phase 3 |
| **Archive node strategy** | (A) Require validators to run own<br>(B) Shared pool<br>(C) Use paid services | **C** for MVP (QuickNode), migrate to B in Phase 2 |
| **Execution sharding** | (A) Random assignment<br>(B) Hash-based<br>(C) Round-robin | **B** (deterministic, no coordination) |
| **Pricing model** | (A) Fixed per execution<br>(B) Complexity-based<br>(C) Auction | **B** (already have complexity formula) |

### 8.2 Business Decisions

1. **Free tier limits:** How many free executions to offer?
   - **Recommendation:** 10/day to allow testing, prevent abuse

2. **Validator count:** How many validators per task?
   - **Recommendation:** 3 for standard, 5 for professional, 7+ for enterprise

3. **Slashing severity:** How much to slash for wrong outputs?
   - **Recommendation:** 10% for malicious, 2% for timing errors

4. **TG pricing:** What's the USD value of 1 TG?
   - **Recommendation:** 1 TG = $0.001 (1000 TG = $1)

### 8.3 Technical Decisions

1. **Script determinism enforcement:**
   - **Option A:** Static analysis (detect `time.Now()`, `random()`, etc.)
   - **Option B:** Runtime sandbox restrictions
   - **Option C:** Manual review + community flagging
   - **Recommendation:** A + B (automated + sandbox)

2. **Non-deterministic API call handling:**
   - **Option A:** Require TLS proofs (TLSNotary)
   - **Option B:** Use on-chain oracles only
   - **Option C:** Allow but mark as "unverifiable"
   - **Recommendation:** A + B (TLS proofs or oracles)

3. **Container pooling strategy:**
   - **Current:** 5 containers per language
   - **Needed:** 20+ containers for per-block execution
   - **Question:** Pool size vs cost?
   - **Recommendation:** Start with 10, scale based on load

---

## 9. Competitive Analysis Summary

### Gelato Network

**How they verify:**
- Single executor runs `checker()` off-chain
- If returns `canExec == true`, executor submits tx
- **No multi-party verification**
- Security relies on executor reputation + slashing

**Pricing:**
- Gas + 2% fee
- Computation-based fee model (upcoming)

**Limitations:**
- No consensus on script output
- Centralized execution decision

### Chainlink Automation

**How they verify:**
- Decentralized oracle network (DON)
- Multiple nodes check conditions
- Threshold consensus before execution

**Pricing:**
- Gas + flat fee (~$0.10 per execution)
- More expensive than Gelato

**Advantages:**
- True decentralization
- Battle-tested (Chainlink reputation)

**Limitations:**
- Higher cost
- Less flexible scripts

### TriggerX Advantage

| Feature | Gelato | Chainlink | TriggerX |
|---------|--------|-----------|----------|
| **Multi-party verification** | ❌ Single executor | ✅ DON consensus | ✅ Othentic consensus |
| **Cost** | ✅ Low (gas + 2%) | ❌ High (gas + $0.10) | ✅ Low (gas + $0.001) |
| **Flexibility** | ✅ Any script | ⚠️ Limited | ✅ Docker sandbox (any language) |
| **Security** | ⚠️ Reputation-based | ✅ Decentralized | ✅ EigenLayer + Othentic |
| **Verifiability** | ❌ No consensus | ✅ Oracle network | ✅ Re-execution consensus |

**Key differentiator:** TriggerX offers Chainlink-level security at Gelato-level pricing.

---

## 10. Summary & Recommendation

### Recommended Approach

**Phase 1 (MVP): Deterministic Re-Execution**
- ✅ Lower complexity & cost
- ✅ Fits Othentic consensus model
- ✅ Proven approach (used by optimistic rollups)
- ✅ Can ship in 4-6 weeks

**Phase 3 (Future): Optional TEE Mode**
- For users who need <1s execution guarantees
- Premium pricing tier
- Reduces validator cost

### Cost Comparison

| Approach | Performer Cost | Validator Cost | Total Cost per Execution |
|----------|----------------|----------------|------------------------|
| **Re-execution** | $0.001 | $0.003 (3 validators) | $0.004 + gas |
| **TEE (EigenCompute)** | $0.01 | $0.0001 (just verify signature) | $0.0101 + gas |

**For 1000 executions/day:**
- Re-execution: $4/day
- TEE: $10.10/day

**Recommendation:** Re-execution for MVP, add TEE as "fast lane" option later.

### Key Success Metrics

| Metric | Target | Rationale |
|--------|--------|-----------|
| **Validation accuracy** | >99.9% | False positives hurt user trust |
| **Execution latency** | <1s per script | Stay within block time |
| **Cost per execution** | <$0.005 | Competitive with Gelato |
| **Slashing rate** | <1% | Ensure honest behavior |
| **User satisfaction** | >4.5/5 | Product-market fit |

---

## 11. Next Steps

1. **Review this document** with the core team
2. **Decide on open questions** (Section 8)
3. **Allocate engineering resources** (suggest 2 engineers for 6 weeks)
4. **Set up testnet environment** (Base Sepolia, OP Sepolia)
5. **Begin Phase 1 implementation** (follow roadmap in Section 7)

---

## Appendix A: Example Block-Based Scripts

### Example 1: Price-Based Liquidation

```javascript
// Script: Check if collateral ratio drops below threshold
const Web3 = require('web3');

async function main() {
    // Read on-chain data at current block
    const web3 = new Web3(process.env.RPC_URL);
    const blockNumber = parseInt(process.env.BLOCK_NUMBER);

    const lendingContract = new web3.eth.Contract(ABI, CONTRACT_ADDRESS);

    // Get user position
    const position = await lendingContract.methods.getPosition(USER_ADDRESS).call({}, blockNumber);
    const collateralValue = position.collateral * await getPrice('ETH', blockNumber);
    const debtValue = position.debt * await getPrice('USDC', blockNumber);

    const collateralRatio = collateralValue / debtValue;

    if (collateralRatio < 1.2) {
        // Trigger liquidation
        console.log(JSON.stringify({
            shouldExecute: true,
            params: [USER_ADDRESS, position.debt],
            metadata: {
                blockNumber: blockNumber,
                collateralRatio: collateralRatio,
                reason: "Collateral ratio below threshold"
            }
        }));
    } else {
        console.log(JSON.stringify({
            shouldExecute: false,
            params: [],
            metadata: {
                blockNumber: blockNumber,
                collateralRatio: collateralRatio
            }
        }));
    }
}

main();
```

### Example 2: Cross-Chain Bridge Monitor

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

type Output struct {
    ShouldExecute bool          `json:"shouldExecute"`
    Params        []interface{} `json:"params"`
    Metadata      Metadata      `json:"metadata"`
}

type Metadata struct {
    BlockNumber uint64 `json:"blockNumber"`
    Reason      string `json:"reason"`
}

func main() {
    blockNumber := getEnvUint64("BLOCK_NUMBER")

    // Check if bridge message is ready to relay
    bridgeContract := getContract("BRIDGE_ADDRESS")
    message := bridgeContract.Call("getPendingMessage", blockNumber, 0)

    if message.IsReady && message.Age > 100 {  // 100 blocks = ~20 min
        output := Output{
            ShouldExecute: true,
            Params: []interface{}{
                message.ID,
                message.Payload,
                message.Proof,
            },
            Metadata: Metadata{
                BlockNumber: blockNumber,
                Reason:      "Bridge message ready for relay",
            },
        }

        jsonOutput, _ := json.Marshal(output)
        fmt.Println(string(jsonOutput))
    } else {
        output := Output{
            ShouldExecute: false,
            Params:        []interface{}{},
            Metadata: Metadata{
                BlockNumber: blockNumber,
                Reason:      "No messages ready",
            },
        }
        jsonOutput, _ := json.Marshal(output)
        fmt.Println(string(jsonOutput))
    }
}
```

### Example 3: Governance Proposal Execution

```python
import json
import os
from web3 import Web3

def main():
    web3 = Web3(Web3.HTTPProvider(os.environ['RPC_URL']))
    block_number = int(os.environ['BLOCK_NUMBER'])

    gov_contract = web3.eth.contract(
        address=os.environ['GOV_CONTRACT'],
        abi=GOV_ABI
    )

    # Check if proposal has passed and is ready to execute
    proposal_id = int(os.environ['PROPOSAL_ID'])
    proposal = gov_contract.functions.proposals(proposal_id).call(block_identifier=block_number)

    current_timestamp = web3.eth.get_block(block_number)['timestamp']

    if proposal['state'] == 4 and proposal['eta'] <= current_timestamp:  # State 4 = Queued
        output = {
            'shouldExecute': True,
            'params': [proposal_id],
            'metadata': {
                'blockNumber': block_number,
                'timestamp': current_timestamp,
                'reason': 'Proposal ready to execute'
            }
        }
    else:
        output = {
            'shouldExecute': False,
            'params': [],
            'metadata': {
                'blockNumber': block_number,
                'state': proposal['state'],
                'eta': proposal['eta']
            }
        }

    print(json.dumps(output))

if __name__ == '__main__':
    main()
```

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **Performer** | Keeper selected to execute the task (via Othentic) |
| **Validator** | Keeper who verifies the performer's execution |
| **Attestation** | Validator's vote (approve/reject) on execution |
| **Slashing** | Penalty applied to dishonest keeper |
| **TG (TriggerGas)** | Prepaid point system for task execution |
| **Deterministic** | Script produces same output given same input |
| **Archive Node** | Node that stores historical blockchain state |
| **TLS Proof** | Cryptographic proof of HTTPS response |
| **TEE** | Trusted Execution Environment (e.g., Intel SGX) |
| **Othentic** | AVS framework for consensus and slashing |

---

**Document Version:** 1.0
**Last Updated:** 2025-11-12
**Author:** TriggerX Core Team
**Status:** Draft for Review
