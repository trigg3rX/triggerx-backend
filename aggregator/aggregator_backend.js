require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const bls = require('noble-bls12-381');
const taskManagerABI = require('../utils/abi/TaskManager.json');

const app = express();
app.use(express.json());

// Configure ethers.js provider and wallet for Holesky
const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

// Load the contract ABI and address
const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';

// BLS key pair (in production, use secure key management)
const blsSecretKey = process.env.BLS_SECRET_KEY;

// Convert hex string to Uint8Array
const blsSecretKeyBytes = Uint8Array.from(Buffer.from(blsSecretKey, 'hex'));
const blsPublicKey = bls.getPublicKey(blsSecretKeyBytes);

// console.log('BLS Public Key:', Buffer.from(blsPublicKey).toString('hex'));

// Mock data for demonstration (replace with actual data in production)
const quorumInfo = {
    quorumNumbers: [1], // Single quorum
    quorumThresholdPercentage: 66,
    signerPubkeys: [blsPublicKey], // Use the actual BLS public key
    signerStakes: [[1000]] // Single keeper's stake
};

// Initialize contract instance
const taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, wallet);

app.post('/receive-result', async (req, res) => {
    try {
        const { jobId, taskId, blockNumber, quorumNumbers, result } = req.body;
        console.log(`Received result for task ${taskId}:`, result);

        console.log("Job ID: ", jobId);
        console.log("Task ID:", taskId);
        console.log("Task Created Block: ", blockNumber);
        console.log("Quorum Numbers: ", quorumNumbers);
        console.log(`-------------------------------------------------------------------------`);

        // Create Task struct
        const task = {
            jobId: ethers.toBigInt(jobId),
            taskCreatedBlock: ethers.toBigInt(blockNumber),
            quorumNumbers: quorumNumbers
        };

        // Create TaskResponse struct
        const taskResponse = {
            referenceTaskIndex: ethers.toBigInt(taskId),
            operator: wallet.address,
            transactionHash: ethers.keccak256(ethers.toUtf8Bytes(JSON.stringify(result)))
        };

        // Generate BLS signature
        const messageToSign = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(
            ['tuple(uint32 referenceTaskIndex, address operator, bytes32 transactionHash)'],
            [taskResponse]
        ));
        const signature = await bls.sign(messageToSign, blsSecretKeyBytes);      
        
        // console.log('BLS Signature:', signature);
        // console.log('BLS PUBLIC KEY:', blsPublicKey);

        const signatureX = `0x${signature.slice(0, 64)}`;
        const signatureY = `0x${signature.slice(128, 192)}`;
        
        // Create NonSignerStakesAndSignature struct here
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

        // console.log('Task:', task);
        // console.log('TaskResponse:', taskResponse);
        // console.log('NonSignerStakesAndSignature:', nonSignerStakesAndSignature);

        // Write transaction path
        // console.log(JSON.stringify({
        //     task: {
        //         ...task,
        //         jobId: task.jobId.toString(),
        //         taskCreatedBlock: task.taskCreatedBlock.toString()
        //     },
        //     taskResponse: {
        //         ...taskResponse,
        //         referenceTaskIndex: taskResponse.referenceTaskIndex.toString()
        //     },
        //     nonSignerStakesAndSignature: {
        //         ...nonSignerStakesAndSignature,
        //         quorumApkIndices: nonSignerStakesAndSignature.quorumApkIndices.map(i => i.toString()),
        //         totalStakeIndices: nonSignerStakesAndSignature.totalStakeIndices.map(i => i.toString())
        //     }
        // }, null, 2));

        const tx = await taskManagerContract.respondToTask(
            task,
            taskResponse,
            nonSignerStakesAndSignature
        );

        const receipt = await tx.wait();

        // check if the transaction is successful
        if (receipt.status === 1) {
            console.log('Transaction successful');
            console.log('Transaction hash:', tx.hash);
            console.log(`-------------------------------------------------------------------------`);
        } else {
            console.error('Transaction failed');
            console.log(`-------------------------------------------------------------------------`);
        }
        
    } catch (error) {
        console.error('Error in aggregation and signing:', error);
        console.error('Error details:', error.stack);
        if (error.reason) console.error('Error reason:', error.reason);
        if (error.code) console.error('Error code:', error.code);
        if (error.method) console.error('Error method:', error.method);
        if (error.transaction) console.error('Error transaction:', error.transaction);
        
        res.status(500).json({ 
            error: 'Failed to aggregate and sign results', 
            details: error.message,
            reason: error.reason,
            code: error.code,
            method: error.method,
            transaction: error.transaction ? {
                ...error.transaction,
                jobId: error.transaction.jobId ? error.transaction.jobId.toString() : undefined,
                taskCreatedBlock: error.transaction.taskCreatedBlock ? error.transaction.taskCreatedBlock.toString() : undefined
            } : undefined
        });
    }
});


const aggregatorPort = 3006;
app.listen(aggregatorPort, () => {
    console.log(`Aggregator service listening at http://localhost:${aggregatorPort}`);
});
