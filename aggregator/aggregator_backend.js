require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const bls = require('noble-bls12-381');

const app = express();
app.use(express.json());

// Configure ethers.js provider and wallet for Holesky
const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

// Load the contract ABI and address
const contractABI = require('./contractABI.json');
const contractAddress = '0xDaa3d01f71F638952db924c9FE4f1CDa847A23Ad';

// BLS key pair (in production, use secure key management)
const blsSecretKey = process.env.BLS_SECRET_KEY;
// Convert hex string to Uint8Array
const blsSecretKeyBytes = Uint8Array.from(Buffer.from(blsSecretKey, 'hex'));
const blsPublicKey = bls.getPublicKey(blsSecretKeyBytes);

console.log('BLS Public Key:', Buffer.from(blsPublicKey).toString('hex'));

// Mock data for demonstration (replace with actual data in production)
const quorumInfo = {
    quorumNumbers: [1], // Single quorum
    quorumThresholdPercentage: 66,
    signerPubkeys: [blsPublicKey], // Use the actual BLS public key
    signerStakes: [[1000]] // Single keeper's stake
};

// Initialize contract instance
const contract = new ethers.Contract(contractAddress, contractABI, wallet);

app.post('/receive-result', async (req, res) => {
    try {
        const { jobId, result } = req.body;
        console.log(`Received result for job ${jobId}:`, result);
        console.log(`-------------------------------------------------------------------------`);

        // Get the current block number
        const currentBlock = await provider.getBlockNumber();
        const taskCreatedBlock = currentBlock;

        console.log("JOB ID: ", jobId);
        console.log("TASK CREATED BLOCK: ", taskCreatedBlock);
        console.log("Quorum Numbers: ", quorumInfo.quorumNumbers);

        // Convert quorum numbers to bytes
        const quorumNumbersBytes = ethers.zeroPadValue(ethers.toBeHex(quorumInfo.quorumNumbers[0]), 32);
        console.log("Quorum Numbers Bytes: ", quorumNumbersBytes);

        // Create Task struct
        const task = {
            jobId: ethers.toBigInt(jobId),
            taskCreatedBlock: ethers.toBigInt(taskCreatedBlock),
            quorumNumbers: quorumNumbersBytes
        };

        // Create TaskResponse struct
        const taskResponse = {
            referenceTaskIndex: ethers.toBigInt(jobId),
            dataHash: ethers.keccak256(ethers.toUtf8Bytes(JSON.stringify(result))),
            operator: wallet.address
        };

        // Generate BLS signature
        const messageToSign = ethers.keccak256(ethers.AbiCoder.defaultAbiCoder().encode(['tuple(uint32 referenceTaskIndex, bytes32 dataHash, address operator)'], [taskResponse]));
        const signature = await bls.sign(messageToSign, blsSecretKeyBytes);

        // Create NonSignerStakesAndSignature struct
        const nonSignerStakesAndSignature = {
            nonSignerQuorumBitmapIndices: [],
            nonSignerPubkeys: [],
            quorumApks: [{
                X: ethers.hexlify(Buffer.from(blsPublicKey.slice(0, 48))),
                Y: ethers.hexlify(Buffer.from(blsPublicKey.slice(48)))
            }],
            apkG2: { X: [ethers.ZeroHash, ethers.ZeroHash], Y: [ethers.ZeroHash, ethers.ZeroHash] },
            sigma: {
                X: ethers.hexlify(Buffer.from(signature.slice(0, 48))),
                Y: ethers.hexlify(Buffer.from(signature.slice(48)))
            },
            quorumApkIndices: [ethers.toBigInt(0)],
            totalStakeIndices: [ethers.toBigInt(0)],
            nonSignerStakeIndices: [[]],
            transactionHash: ethers.keccak256(ethers.toUtf8Bytes('dummyTransactionHash')) // Add this line
        };


        console.log('Task:', task);
        console.log('TaskResponse:', taskResponse);
        console.log('NonSignerStakesAndSignature:', nonSignerStakesAndSignature);

        // Write transaction path
        console.log('Calling respondToTask with parameters:');
        console.log(JSON.stringify({
            task: {
                ...task,
                jobId: task.jobId.toString(),
                taskCreatedBlock: task.taskCreatedBlock.toString()
            },
            taskResponse: {
                ...taskResponse,
                referenceTaskIndex: taskResponse.referenceTaskIndex.toString()
            },
            nonSignerStakesAndSignature: {
                ...nonSignerStakesAndSignature,
                quorumApkIndices: nonSignerStakesAndSignature.quorumApkIndices.map(i => i.toString()),
                totalStakeIndices: nonSignerStakesAndSignature.totalStakeIndices.map(i => i.toString()),
                transactionHash: nonSignerStakesAndSignature.transactionHash // Add this line
            }
        }, null, 2));

        const tx = await contract.respondToTask(
            task,
            taskResponse,
            nonSignerStakesAndSignature
        );

        console.log('Transaction sent, tx:', tx);

        // Wait for the transaction to be mined
        try {
            const receipt = await tx.wait();
            console.log('Transaction mined, receipt:', receipt);
        } catch (error) {
            console.error('Error while waiting for transaction to be mined:', error);
        }

    } catch (error) {
        console.error('Error in aggregation and signing:', error);
        console.error('Error details:', error.stack);
        if (error.reason) console.error('Error reason:', error.reason);
        if (error.code) console.error('Error code:', error.code);
        if (error.method) console.error('Error method:', error.method);
        if (error.transaction) console.error('Error transaction:', error.transaction);
        
        // Convert BigInt fields to strings for the error response
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
