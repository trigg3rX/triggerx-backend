/**
 * Simple Safe Transaction Test - TaskDefinitionID 7
 * Executes a simple ETH transfer through a Gnosis Safe
 */

interface TriggerXOutput {
  shouldExecute: boolean;
  targetContract?: string;
  calldata?: string;
  storageUpdates?: Record<string, string>;
  metadata: {
    timestamp: number;
    reason: string;
    gasEstimate?: number;
  };
}

async function main(): Promise<TriggerXOutput> {
  const timestamp = Date.now();

  // Fixed Safe address
  const safeAddress = "0x87EB883e8ae00120EF2c6Fd49b1F8A149E2172f4";

  // ETH transfer parameters
  const actionTarget = "0xa76Cacba495CafeaBb628491733EB86f1db006dF";
  const actionValue = "10000000000000"; // 0.00001 ETH in wei
  const actionData = "0x"; // empty calldata for plain ETH transfer
  const operation = 0; // 0 = CALL

  // Simple calldata encoding
  // For testing, we'll create a simple function call
  // transfer(address,uint256) or similar
  const functionSelector = "a9059cbb"; // transfer function selector

  // Encode recipient address (remove 0x, pad to 32 bytes)
  const recipient = actionTarget.replace('0x', '').toLowerCase().padStart(64, '0');

  // Encode amount (convert to hex, pad to 32 bytes)
  const amount = BigInt(actionValue).toString(16).padStart(64, '0');

  const calldata = "0x" + functionSelector + recipient + amount;

  return {
    shouldExecute: true,
    targetContract: "0xa0bC1477cfc452C05786262c377DE51FB8bc4669",
    calldata: calldata,
    storageUpdates: {
      "lastExecutionTime": timestamp.toString(),
      "lastActionTarget": actionTarget,
      "lastActionValue": actionValue,
      "lastSafeAddress": safeAddress,
      "executionCount": "1",
      "status": "executed"
    },
    metadata: {
      timestamp: timestamp,
      reason: `Executing Safe transaction: ${actionValue} wei to ${actionTarget}`,
      gasEstimate: 150000
    }
  };
}

// Execute and output JSON to stdout
main().then(output => {
  console.log(JSON.stringify(output));
}).catch(error => {
  const errorOutput: TriggerXOutput = {
    shouldExecute: false,
    storageUpdates: {
      "lastError": error.message,
      "errorTime": Date.now().toString()
    },
    metadata: {
      timestamp: Date.now(),
      reason: `Error: ${error.message}`
    }
  };
  console.log(JSON.stringify(errorOutput));
  process.exit(1);
});
