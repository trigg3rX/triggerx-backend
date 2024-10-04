const express = require('express');
const bls = require('@noble/bls12-381');
const { ethers } = require('ethers');
const crypto = require('crypto');

const aggregatorApp = express();
aggregatorApp.use(express.json());

const jobResults = new Map();
let blsKeyPair;

async function initializeBLSKeyPair() {
    const privateKey = crypto.randomBytes(32);
    const publicKey = await bls.getPublicKey(privateKey);
    blsKeyPair = { privateKey, publicKey };
    console.log('BLS key pair initialized');
}

async function aggregateAndSign(jobId) {
    const results = jobResults.get(jobId) || [];
    if (results.length === 0) return null;

    const aggregatedResult = results.join(',');
    console.log('Aggregated result:', aggregatedResult);

    // Use ethers v6 utilities
    const messageHash = ethers.keccak256(ethers.toUtf8Bytes(aggregatedResult));
    console.log('Message hash:', messageHash);
    console.log('Private key:', Buffer.from(blsKeyPair.privateKey).toString('hex'));

    const signature = await bls.sign(blsKeyPair.privateKey, ethers.getBytes(messageHash));
    console.log('Signature:', Buffer.from(signature).toString('hex'));

    const signedResult = {
        jobId,
        aggregatedResult,
        signature: Buffer.from(signature).toString('hex'),
        publicKey: Buffer.from(blsKeyPair.publicKey).toString('hex')
    };

    console.log('Signed Result:', signedResult);

    return signedResult;
}

aggregatorApp.post('/receive-result', async (req, res) => {
    const { jobId, result } = req.body;
    
    console.log(`Aggregator received result for job ${jobId}:`, result);

    if (!jobResults.has(jobId)) {
        jobResults.set(jobId, []);
    }
    jobResults.get(jobId).push(result);
    console.log("JOB ID: ", jobId);

    if (jobResults.get(jobId).length >= 3) {
        console.log("Processing aggregation...");
        try {
            const aggregatedData = await aggregateAndSign(jobId);
            console.log('Aggregated and signed data:');
            console.log(JSON.stringify(aggregatedData, null, 2));

            jobResults.delete(jobId);

            res.status(200).json({
                message: `Results for job ${jobId} aggregated and signed.`,
                data: aggregatedData
            });
        } catch (error) {
            console.error('Error in aggregation and signing:', error);
            res.status(500).json({ error: 'Failed to aggregate and sign results', details: error.message });
        }
    } else {
        res.status(200).send(`Result for job ${jobId} received and stored. Current count: ${jobResults.get(jobId).length}`);
    }
});

const aggregatorPort = 3006;
aggregatorApp.listen(aggregatorPort, async () => {
    await initializeBLSKeyPair();
    console.log(`Aggregator service listening at http://localhost:${aggregatorPort}`);
});