require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const bls = require('noble-bls12-381');

const app = express();
app.use(express.json());

// Configure ethers.js provider and wallet for Holesky
const provider = new ethers.providers.JsonRpcProvider('https://holesky.infura.io/v3/' + process.env.INFURA_PROJECT_ID);
const wallet = new ethers.Wallet(process.env.ETH_PRIVATE_KEY, provider);

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

        // Get the current block number
        const currentBlock = await provider.getBlockNumber();
        const taskCreatedBlock = currentBlock;

        // Create Task struct
        const task = {
            taskId: ethers.utils.hexlify(jobId),
            taskCreatedBlock: ethers.utils.hexlify(taskCreatedBlock),
            quorumNumbers: ethers.utils.hexlify(quorumInfo.quorumNumbers),
            quorumThresholdPercentage: ethers.utils.hexlify(quorumInfo.quorumThresholdPercentage)
        };

        // Create TaskResponse struct
        const taskResponse = {
            referenceTaskIndex: ethers.utils.hexlify(jobId),
            dataHash: ethers.utils.keccak256(ethers.utils.toUtf8Bytes(JSON.stringify(result)))
        };

        // Generate BLS signature
        const messageToSign = taskResponse.dataHash;
        const signature = await bls.sign(messageToSign, blsSecretKeyBytes);

        // Convert signature to hex string
        const signatureHex = Buffer.from(signature).toString('hex');

        // Create NonSignerStakesAndSignature struct
        const nonSignerStakesAndSignature = {
            nonSignerQuorumBitmapIndices: [],
            nonSignerPubkeys: [],
            quorumApks: [{
                X: ethers.utils.hexlify(Buffer.from(blsPublicKey.slice(0, 48))),
                Y: ethers.utils.hexlify(Buffer.from(blsPublicKey.slice(48)))
            }],
            apkG2: { X: [ethers.utils.hexlify(0), ethers.utils.hexlify(0)], Y: [ethers.utils.hexlify(0), ethers.utils.hexlify(0)] },
            sigma: {
                X: ethers.utils.hexlify(Buffer.from(signature.slice(0, 48))),
                Y: ethers.utils.hexlify(Buffer.from(signature.slice(48)))
            },
            quorumApkIndices: [ethers.utils.hexlify(0)],
            totalStakeIndices: [ethers.utils.hexlify(0)],
            nonSignerStakeIndices: [[]]
        };

        console.log('Task:', task);
        console.log('TaskResponse:', taskResponse);
        console.log('NonSignerStakesAndSignature:', nonSignerStakesAndSignature);

        // Call the respondToTask function on the Holesky contract
        const tx = await contract.respondToTask(
            task,
            taskResponse,
            nonSignerStakesAndSignature,
            { gasLimit: 1000000 }
        );

        console.log('Transaction sent, tx:', tx.hash);

        // Wait for the transaction to be mined
        const receipt = await tx.wait();
        console.log('Transaction receipt:', receipt);

        res.status(200).json({ 
            message: 'Result received, signed with BLS, and sent to Holesky contract', 
            transactionHash: tx.hash,
            receipt
        });
    } catch (error) {
        console.error('Error processing result:', error);
        res.status(500).json({ error: 'Error processing result', details: error.message });
    }
});

const port = process.env.AGGREGATOR_PORT || 3002;
app.listen(port, () => {
    console.log(`Aggregator backend listening at http://localhost:${port}`);
});
