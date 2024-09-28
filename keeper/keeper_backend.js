require('dotenv').config(); 
const express = require('express');
const axios = require('axios');
const { ethers } = require('ethers');

const keeperApp = express();
keeperApp.use(express.json());

const aggregatorUrl = 'http://localhost:3002/receive-result'; 
const etherscanApiKey = 'U5X9SJAFNJY7FS3TZWMWTVYJZ7Q1K6QJKM'; 
const provider = new ethers.JsonRpcProvider('https://opt-sepolia.g.alchemy.com/v2/xd07TFzs6Ele-LHAffmzsBiuC8k32VZv'); 
const privateKey = process.env.PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey, provider);

// Middleware to fetch and standardize API data
async function fetchAndStandardizeData(req, res, next) {
    const { apiEndpoint, argType } = req.body;

    // Skip API check for 'static' or 'none' argument types
    if (argType === '0' || argType === '1' || argType === 'None' || argType === 'Static') {
        req.standardizedData = { data: { value: null } };
        return next();
    }

    if (!apiEndpoint) {
        return res.status(400).send('API endpoint is required for dynamic arguments.');
    }

    try {
        const response = await axios.get(apiEndpoint);
        
        // Validate the structure of the response
        if (!response.data || typeof response.data.data === 'undefined' || typeof response.data.data.value === 'undefined') {
            throw new Error('API response does not match the expected structure. Expected: { "data": { "value": <value> } }');
        }

        // Standardize the response format
        req.standardizedData = {
            data: {
                value: response.data.data.value
            }
        };
        next();
    } catch (error) {
        console.error(`Error fetching data from ${apiEndpoint}:`, error.message);
        return res.status(500).send(`Error fetching data from API: ${error.message}`);
    }
}

// Function to fetch ABI from Etherscan
async function fetchABI(contractAddress) {
    const url = `https://api-sepolia-optimistic.etherscan.io/api?module=contract&action=getabi&address=${contractAddress}&apikey=${etherscanApiKey}`;
    
    try {
        const response = await axios.get(url);
        const data = response.data;
        if (data.status === '1' && data.result) {
            return JSON.parse(data.result);
        } else {
            console.error('Error fetching ABI:', data.message || 'Unknown error');
            return null;
        }
    } catch (error) {
        console.error('Fetch error:', error.message);
        return null;
    }
}

// Function to handle task execution based on argument type
async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, standardizedData } = task;
    const dynamicData = standardizedData.data.value;

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);

    let args = [];
    let argTypeString;

    // Fetch the ABI dynamically
    const abi = await fetchABI(contractAddress);
    if (!abi) {
        throw new Error(`Failed to fetch ABI for contract ${contractAddress}`);
    }
    console.log('Fetched ABI:', JSON.stringify(abi, null, 2));

    const contract = new ethers.Contract(contractAddress, abi, wallet);

    // Determine arguments based on argType
    switch (argType) {
        case '0':
        case 'None':
            args = [];
            argTypeString = 'None';
            break;
        case '1':
        case 'Static':
            args = task.argumentInfo.arguments;
            argTypeString = 'Static';
            break;
        case '2':
        case 'Dynamic':
            args = [dynamicData];
            argTypeString = 'Dynamic';
            break;
        default:
            throw new Error(`Invalid argument type: ${argType}`);
    }

    try {
        console.log(`Calling function ${targetFunction} with arguments:`, args);
        
        if (typeof contract[targetFunction] !== 'function') {
            throw new Error(`Function ${targetFunction} does not exist in the contract ABI`);
        }

        // Check if the contract is deployed
        const code = await provider.getCode(contractAddress);
        if (code === '0x') {
            throw new Error(`No contract deployed at address ${contractAddress}`);
        }

        // Try calling as a read-only function first
        try {
            const result = await contract[targetFunction].staticCall(...args);
            console.log(`Function ${targetFunction} executed successfully as read-only. Result:`, result);
            return result;
        } catch (staticCallError) {
            console.log(`Static call failed, attempting as state-changing function:`, staticCallError.message);
            
            // If static call fails, try as a state-changing function
            const tx = await contract[targetFunction](...args);
            console.log('Transaction sent. Waiting for confirmation...');
            const receipt = await tx.wait();
            console.log(`Transaction mined. Block number:`, receipt.blockNumber);
            
            // Try to get the return value from the transaction receipt
            const returnValue = receipt.logs[0]?.data;
            console.log(`Function ${targetFunction} executed successfully as state-changing. Return value:`, returnValue);
            return returnValue;
        }
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
const keeperPort = 3001;
keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service listening at http://localhost:${keeperPort}`);
});