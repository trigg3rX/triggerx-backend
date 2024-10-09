require('dotenv').config(); 
const express = require('express');
const axios = require('axios');
const { ethers } = require('ethers');
const TronWeb = require('tronweb');

const keeperApp = express();
keeperApp.use(express.json());

const aggregatorUrl = 'http://localhost:3002/receive-result'; 
const etherscanApiKey = 'U5X9SJAFNJY7FS3TZWMWTVYJZ7Q1K6QJKM'; 
const provider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc'); 
const privateKey = process.env.HOLESKY_PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey, provider);
const tronPrivateKey = process.env.TRON_PRIVATE_KEY;
if (!tronPrivateKey) {
    throw new Error('TRON_PRIVATE_KEY is not set in the environment variables');
}

const tronWeb = new TronWeb({
    fullHost: 'https://nile.trongrid.io', // Replace with the appropriate TRON network URL
    privateKey: tronPrivateKey // Make sure to set this in your .env file
});
tronWeb.setAddress(tronWeb.address.fromPrivateKey(tronPrivateKey));
console.log('TronWeb initialized with address:', tronWeb.defaultAddress.base58);

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
        const abi = JSON.stringify(contract.abi, null, 2)
        return abi;
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

    // Fetch the ABI dynamically
    const abi = await fetchABI(contractAddress);
    if (!abi) {
        throw new Error(`Failed to fetch ABI for contract ${contractAddress}`);
    }

    // Initialize the contract instance
    let contract;
    const baseAddress = tronWeb.address.fromHex(contractAddress);
    try {
        contract = await tronWeb.contract().at(baseAddress);
        console.log(`Contract initialized at address ${baseAddress}`);
    } catch (error) {
        console.error(`Error initializing contract at address ${baseAddress}:`, error);
        throw error;
    }

    // Determine arguments based on argType
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

    try {
        console.log(`Calling function ${targetFunction} with arguments:`, args);

        // Ensure the function is available in the contract
        if (typeof contract[targetFunction] !== 'function') {
            throw new Error(`Function ${targetFunction} does not exist in the contract ABI`);
        }

        // Explicitly set the owner_address (from address)
        const method = contract[targetFunction](...args);

        // Set the caller's address (owner_address)
        const callerAddress = tronWeb.defaultAddress.base58;
        if (!callerAddress) {
            throw new Error('TronWeb default address is not set');
        }

        // Call the function
        let result = await method.send({
            from: callerAddress,
            callValue: 1
        });

        const serializedResult = result.toString();

        console.log(`Function call ${targetFunction} executed successfully.`);
        console.log(`-------------------------------------------------------------------------`);
        return serializedResult;
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
        
        console.log(`Result: `, result);
        console.log(`-------------------------------------------------------------------------`);
        
        // Send the result to the aggregator
        await axios.post(aggregatorUrl, { 
            jobId: task.jobId, 
            result: result
        });
        console.log('Task result sent to aggregator');

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
const keeperPort = 3001;
keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service listening at http://localhost:${keeperPort}`);
});