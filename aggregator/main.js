require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const bls = require('noble-bls12-381');
const taskManagerABI = require('../utils/abi/TaskManager.json');

const app = express();
app.use(express.json());

const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';
const blsSecretKey = process.env.BLS_SECRET_KEY;
const blsSecretKeyBytes = Uint8Array.from(Buffer.from(blsSecretKey, 'hex'));
const blsPublicKey = bls.getPublicKey(blsSecretKeyBytes);

const taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, wallet);

// Batch processing configuration
const BATCH_TIMEOUT = 20000; // 20 seconds
const MAX_BATCH_SIZE = 10;   // Maximum number of transactions in a batch
let pendingTasks = [];
let batchTimer = null;

async function processTaskBatch() {
    if (pendingTasks.length === 0) return;

    console.log(`Processing batch of ${pendingTasks.length} tasks`);
    const currentBatch = [...pendingTasks];
    pendingTasks = [];

    try {
        // Create multicall interface
        const multicallInterface = new ethers.Interface([
            'function aggregate((address target, bytes callData)[] calls) returns (uint256 blockNumber, bytes[] returnData)'
        ]);

        // Prepare batch calls
        const calls = currentBatch.map(taskData => ({
            target: taskManagerAddress,
            callData: taskManagerContract.interface.encodeFunctionData(
                'respondToTask',
                [taskData.task, taskData.taskResponse, taskData.nonSignerStakesAndSignature]
            )
        }));

        // Execute batch transaction
        const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11'; // Multicall3 address
        const multicall = new ethers.Contract(
            multicallAddress,
            multicallInterface,
            wallet
        );

        const tx = await multicall.aggregate(calls);
        const receipt = await tx.wait();

        if (receipt.status === 1) {
            console.log('Batch transaction successful');
            console.log('Transaction hash:', tx.hash);
            
            // Resolve all promises in the batch
            currentBatch.forEach(taskData => {
                taskData.resolve({ 
                    success: true, 
                    message: 'Task processed in batch', 
                    transactionHash: tx.hash 
                });
            });
        } else {
            throw new Error('Batch transaction failed');
        }

    } catch (error) {
        console.error('Error processing batch:', error);
        // Reject all promises in the batch
        currentBatch.forEach(taskData => {
            taskData.reject({
                success: false,
                message: 'Batch processing failed',
                error: error.message
            });
        });
    }
}

function scheduleBatchProcessing() {
    if (batchTimer) clearTimeout(batchTimer);
    
    batchTimer = setTimeout(() => {
        processTaskBatch();
    }, BATCH_TIMEOUT);
}

app.post('/receive-result', async (req, res) => {
    try {
        const { jobId, taskId, blockNumber, quorumNumbers, result: taskResult } = req.body;
        console.log(`Queuing result for task ${taskId}:`, taskResult);

        const task = {
            jobId: ethers.toBigInt(jobId),
            taskCreatedBlock: ethers.toBigInt(blockNumber),
            quorumNumbers: quorumNumbers
        };

        const taskResponse = {
            referenceTaskIndex: ethers.toBigInt(taskId),
            operator: wallet.address,
            transactionHash: ethers.keccak256(ethers.toUtf8Bytes(JSON.stringify(taskResult)))
        };

        const messageToSign = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(
            ['tuple(uint32 referenceTaskIndex, address operator, bytes32 transactionHash)'],
            [taskResponse]
        ));
        const signature = await bls.sign(messageToSign, blsSecretKeyBytes);      

        const signatureX = `0x${signature.slice(0, 64)}`;
        const signatureY = `0x${signature.slice(128, 192)}`;
        
        const nonSignerStakesAndSignature = {
            nonSignerQuorumBitmapIndices: [],
            nonSignerPubkeys: [],
            quorumApks: [{ 
                X: ethers.hexlify(Buffer.from(blsPublicKey.slice(0, 24))), 
                Y: ethers.hexlify(Buffer.from(blsPublicKey.slice(24))) 
            }],
            apkG2: {
                X: [ethers.ZeroHash, ethers.ZeroHash],
                Y: [ethers.ZeroHash, ethers.ZeroHash]
            },
            sigma: { 
                X: signatureX,
                Y: signatureY
            },
            quorumApkIndices: [ethers.toBigInt(0)],
            totalStakeIndices: [ethers.toBigInt(0)],
            nonSignerStakeIndices: []
        };

        // Create a promise for this task
        const taskPromise = new Promise((resolve, reject) => {
            pendingTasks.push({
                task,
                taskResponse,
                nonSignerStakesAndSignature,
                resolve,
                reject
            });
        });

        // Schedule batch processing if this is the first task
        if (pendingTasks.length === 1) {
            scheduleBatchProcessing();
        }

        // Process immediately if batch size is reached
        if (pendingTasks.length >= MAX_BATCH_SIZE) {
            clearTimeout(batchTimer);
            processTaskBatch();
        }

        // Wait for the task to be processed
        const processedResult = await taskPromise;
        res.status(200).json(processedResult);

    } catch (error) {
        console.error('Error in task processing:', error);
        res.status(500).json({ 
            success: false, 
            message: 'Error processing task', 
            error: error.message 
        });
    }
});

const aggregatorPort = 3006;
app.listen(aggregatorPort, () => {
    console.log(`Aggregator service listening at http://localhost:${aggregatorPort}`);
});