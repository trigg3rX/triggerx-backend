require('dotenv').config(); 
const express = require('express');
const axios = require('axios');
const { ethers } = require('ethers');
const TronWeb = require('tronweb');

const keeperApp = express();
keeperApp.use(express.json());

const aggregatorUrl = 'http://localhost:3006/receive-result'; 
const etherscanApiKey = 'U5X9SJAFNJY7FS3TZWMWTVYJZ7Q1K6QJKM'; 
const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc'); 
const privateKey = process.env.HOLESKY_PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey, provider);

const tronWeb = new TronWeb({
    fullHost: 'https://nile.trongrid.io', // Replace with the appropriate TRON network URL
    privateKey: process.env.TRON_PRIVATE_KEY // Make sure to set this in your .env file
});

// Middleware to fetch and standardize API data
async function fetchAndStandardizeData(req, res, next) {
    const { argType, apiEndpoint } = req.body;

    // For argType 0 or 1, skip API check and proceed
    if (argType === 0 || argType === 1 || argType === 'None' || argType === 'Static') {
        req.standardizedData = { data: { value: null } };
        return next();
    }

    if (!apiEndpoint || apiEndpoint === 'null') {
        console.warn('API endpoint is missing or null for a dynamic argument type.');
        req.standardizedData = { data: { value: null } };
        return next();
    }

    try {
        const response = await axios.get(apiEndpoint);
        // Rest of the function remains the same
    } catch (error) {
        console.error(`Error fetching data from ${apiEndpoint}:`, error.message);
        req.standardizedData = { data: { value: null } };
        next();
    }
}

// Function to fetch ABI from Etherscan
async function fetchABI(contractAddress) {
    try {
        const contract = await tronWeb.contract().at(contractAddress);
        return contract.abi;
    } catch (error) {
        console.error("Error fetching ABI from TRON contract:", error);
        return null;
    }
}

// Function to handle task execution based on argument type
// Function to handle task execution based on argument type
async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, standardizedData } = task;

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);

    // Prepare the function selector and parameters
    const functionSelector = targetFunction; // e.g., 'someFunction(bytes)'
    let args;
    switch (argType) {
        case 0:  // For argType 0 (None)
            args = [];
            break;
        case 1:  // For argType 1 (Static)
            args = task.argumentInfo.arguments;
            break;
        case 2:  // For argType 2 (Dynamic)
            args = [standardizedData.data.value];
            break;
        default:
            throw new Error(`Invalid argument type: ${argType}`);
    }

    // Convert the arguments to `bytes` if necessary
    const parameters = args.map(arg => ({
        type: 'bytes', // Now handling `bytes` type
        value: tronWeb.toHex(arg) // Convert the argument to hex if it's a dynamic value
    }));

    try {
        console.log(`Calling function ${functionSelector} on contract ${contractAddress} with arguments:`, parameters);

        // Call the function using triggerConstantContract
        const result = await tronWeb.transactionBuilder.triggerConstantContract(
            contractAddress,
            functionSelector,
            {},  // No specific options required in this case
            parameters
        );

        // Extract the result from the constant_result field
        const constantResult = result.constant_result[0];
        console.log('Raw constant result (hex):', constantResult);

        // Handle the result, converting from `bytes` (hex) to readable form if necessary
        const decodedResult = tronWeb.toUtf8(constantResult); // Decoding from hex to UTF-8 string if applicable
        console.log(`Function ${targetFunction} executed successfully. Decoded result:`, decodedResult);

        return decodedResult;
    } catch (error) {
        console.error(`Error executing function ${targetFunction}:`, error);
        throw error;
    }
}




// Endpoint to receive tasks
keeperApp.post('/execute-task', fetchAndStandardizeData, async (req, res) => {
    const task = req.body;
    task.standardizedData = req.standardizedData; // Attach standardized data to task

    console.log('Keeper received task:', task);

    try {
        const result = await executeTask(task);
        
        // Convert BigInt to string if necessary
        const serializedResult = typeof result === 'bigint' ? result.toString() : result;
        
        // Send the result to the aggregator
        await axios.post(aggregatorUrl, { 
            jobId: task.jobId, 
            result: serializedResult
        });
        console.log('Task result sent to aggregator:', serializedResult);

        res.status(200).send('Task executed and result sent to aggregator');
    } catch (error) {
        console.error('Error executing task:', error);
        console.error('Error details:', {
            message: error.message,
            stack: error.stack
        });
        res.status(500).json({
            error: 'Error executing task',
            details: error.message,
            stack: error.stack
        });
    }
});

// Start the keeper server
const keeperPort = 3005;
keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service listening at http://localhost:${keeperPort}`);
});