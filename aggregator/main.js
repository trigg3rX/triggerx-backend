require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const bls = require('noble-bls12-381');
const taskManagerABI = require('../utils/abi/TaskManager.json');
const serviceManagerABI = require('../utils/abi/ServiceManager.json');
const multiCallABI = require('../utils/abi/MultiCall.json');
const app = express();
app.use(express.json());

const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';
const blsSecretKey = process.env.BLS_SECRET_KEY;
const blsSecretKeyBytes = Uint8Array.from(Buffer.from(blsSecretKey, 'hex'));
const blsPublicKey = bls.getPublicKey(blsSecretKeyBytes);

const taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, wallet);

let serviceManagerContract;

// Batch processing configuration
const BATCH_TIMEOUT = 20000; // 20 seconds
const MAX_BATCH_SIZE = 10;   // Maximum number of transactions in a batch
let pendingTasks = [];
let batchTimer = null;

async function initializeServiceManager() {
    const serviceManagerAddress = await taskManagerContract.serviceManager();
    serviceManagerContract = new ethers.Contract(serviceManagerAddress, serviceManagerABI, wallet);
}

async function validateTask(taskData) {
    try {
        // Initialize ServiceManager if not already initialized
        if (!serviceManagerContract) {
            await initializeServiceManager();
        }

        // Check if task exists first
        const taskHash = await taskManagerContract.allTaskHashes(
            taskData.taskResponse.referenceTaskIndex
        );
        if (taskHash === ethers.ZeroHash) {
            throw new Error('Task does not exist');
        }

        // Check if task already has a response
        const taskResponse = await taskManagerContract.allTaskResponses(
            taskData.taskResponse.referenceTaskIndex
        );
        if (taskResponse !== ethers.ZeroHash) {
            throw new Error('Task already has a response');
        }

        // Check if operator is blacklisted using the correct public view function
        const isBlacklisted = await serviceManagerContract.isOperatorBlacklisted(
            taskData.taskResponse.operator
        );
        if (isBlacklisted) {
            throw new Error('Operator is blacklisted');
        }

        // Verify task data integrity
        if (!taskData.task || !taskData.taskResponse || !taskData.nonSignerStakesAndSignature) {
            throw new Error('Missing required task data');
        }

        // Calculate and verify task hash matches
        const encodedTask = ethers.AbiCoder.defaultAbiCoder().encode(
            ['tuple(uint32 jobId, uint32 taskCreatedBlock, bytes quorumNumbers)'],
            [taskData.task]
        );
        const calculatedTaskHash = ethers.keccak256(encodedTask);
        
        if (calculatedTaskHash !== taskHash) {
            throw new Error('Task hash mismatch');
        }

        // Check if we're still within the response window
        const currentBlock = await provider.getBlockNumber();
        const responseWindow = await taskManagerContract.getTaskResponseWindowBlock();
        
        if (currentBlock > taskData.task.taskCreatedBlock + responseWindow) {
            throw new Error('Response window expired');
        }

        // Simulate the call with explicit gas limit
        await taskManagerContract.respondToTask.staticCall(
            taskData.task,
            taskData.taskResponse,
            taskData.nonSignerStakesAndSignature,
            { gasLimit: 500000 }
        );
        
        return true;
    } catch (error) {
        console.error(`Task validation failed for task ${taskData.taskResponse.referenceTaskIndex}:`, error.message);
        return false;
    }
}

async function processSingleTask(taskData) {
    try {
        const tx = await taskManagerContract.respondToTask(
            taskData.task,
            taskData.taskResponse,
            taskData.nonSignerStakesAndSignature,
            { gasLimit: 500000 }
        );
        const receipt = await tx.wait();
        return { success: true, txHash: receipt.hash };
    } catch (error) {
        console.error('Single task processing failed:', error);
        throw error;
    }
}

async function processTaskBatch() {
    if (pendingTasks.length === 0) return;

    console.log(`Processing batch of ${pendingTasks.length} tasks`);
    
    // Validate tasks first
    const validTasks = [];
    const invalidTasks = [];
    
    for (const task of pendingTasks) {
        const isValid = await validateTask(task);
        if (isValid) {
            validTasks.push(task);
        } else {
            invalidTasks.push(task);
        }
    }

    // Clear pending tasks
    pendingTasks = [];

    if (validTasks.length === 0) {
        console.log('No valid tasks to process');
        return;
    }

    // If only one task, process it directly
    if (validTasks.length === 1) {
        try {
            const result = await processSingleTask(validTasks[0]);
            validTasks[0].resolve(result);
            return;
        } catch (error) {
            validTasks[0].reject({
                success: false,
                message: 'Task processing failed',
                error: error.message
            });
            return;
        }
    }

    try {
        const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11';
        const multicall = new ethers.Contract( multicallAddress, multiCallABI, wallet);

        // Use tryAggregate instead of aggregate3
        const calls = validTasks.map(taskData => ({
            target: taskManagerAddress,
            callData: taskManagerContract.interface.encodeFunctionData(
                'respondToTask',
                [taskData.task, taskData.taskResponse, taskData.nonSignerStakesAndSignature]
            )
        }));

        // First try with requireSuccess = true
        //const gasEstimate = await multicall.aggregate.estimateGas(true, calls);
        
        // Add 30% buffer to the gas estimate
        //const gasLimit = (gasEstimate * BigInt(17)) / BigInt(10);

        // Execute the batch with the estimated gas limit
        const tx = await multicall.aggregate(calls);
        
        const receipt = await tx.wait();

        if (receipt.status === 1) {
            console.log('Batch transaction successful:', tx.hash);
            
            // Parse the results to check individual call success
            const events = receipt.logs.map(log => {
                try {
                    return taskManagerContract.interface.parseLog(log);
                } catch (e) {
                    return null;
                }
            }).filter(Boolean);

            validTasks.forEach((task, index) => {
                task.resolve({ 
                    success: true, 
                    message: 'Task processed', 
                    txHash: tx.hash,
                    events: events.filter(event => 
                        event.args?.referenceTaskIndex === task.taskResponse.referenceTaskIndex
                    )
                });
            });
        } else {
            throw new Error('Transaction failed');
        }

    } catch (error) {
        console.error('Batch processing error:', error);
        
        // If batch fails, try processing in smaller batches
        console.log('Attempting to process tasks in smaller batches...');
        
        // Split tasks into smaller batches of 2
        const batchSize = 2;
        for (let i = 0; i < validTasks.length; i += batchSize) {
            const batch = validTasks.slice(i, i + batchSize);
            
            try {
                // Add delay between batches to prevent nonce issues
                if (i > 0) {
                    await new Promise(resolve => setTimeout(resolve, 2000));
                }
                
                // Process smaller batch
                const calls = batch.map(taskData => ({
                    target: taskManagerAddress,
                    callData: taskManagerContract.interface.encodeFunctionData(
                        'respondToTask',
                        [taskData.task, taskData.taskResponse, taskData.nonSignerStakesAndSignature]
                    )
                }));

                const tx = await multicall.tryAggregate(true, calls, {
                    gasLimit: 1000000 // Use fixed higher gas limit for smaller batches
                });
                
                const receipt = await tx.wait();
                
                if (receipt.status === 1) {
                    batch.forEach((task) => {
                        task.resolve({
                            success: true,
                            message: 'Task processed in smaller batch',
                            txHash: receipt.hash
                        });
                    });
                }
            } catch (batchError) {
                console.error('Small batch processing failed:', batchError);
                
                // If small batch fails, process tasks individually
                for (const task of batch) {
                    try {
                        await new Promise(resolve => setTimeout(resolve, 1000));
                        const result = await processSingleTask(task);
                        task.resolve(result);
                    } catch (taskError) {
                        task.reject({
                            success: false,
                            message: 'Individual task processing failed',
                            error: taskError.message
                        });
                    }
                }
            }
        }
    }

    // Handle invalid tasks
    invalidTasks.forEach(task => {
        task.reject({
            success: false,
            message: 'Task validation failed'
        });
    });
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