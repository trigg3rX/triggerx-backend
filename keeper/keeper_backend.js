require('dotenv').config(); 
const express = require('express');
const axios = require('axios');
const { ethers } = require('ethers');

const keeperApp = express();
keeperApp.use(express.json());

const aggregatorUrl = 'http://localhost:3002/receive-result'; 
const etherscanApiKey = 'V332PZUHEE7V97ZP2V7YV3YPA6RXNR4XEV'; 
const provider = new ethers.JsonRpcProvider('https://eth-sepolia.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc'); 
const privateKey = process.env.PRIVATE_KEY;
const wallet = new ethers.Wallet(privateKey, provider);

// Middleware to fetch and standardize API data
async function fetchAndStandardizeData(req, res, next) {
    const { apiEndpoint } = req.body;

    if (!apiEndpoint) {
        return res.status(400).send('API endpoint is required.');
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
                value: response.data.data.value // This extracts the value
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
    const url = `https://api-sepolia.etherscan.io/api?module=contract&action=getabi&address=${contractAddress}&apikey=${etherscanApiKey}`;
    try {
        const response = await axios.get(url);
        const data = response.data;

        if (data.status === '1') {
            console.log('ABI fetched successfully');
            return JSON.parse(data.result); // ABI is returned as a JSON string
        } else {
            console.error(`Failed to fetch ABI: ${data.message}`);
            if (data.message === 'NOTOK') {
                console.error(`Contract address ${contractAddress} might not be verified on Etherscan or there is another issue.`);
            }
            throw new Error(`Failed to fetch ABI: ${data.message}`);
        }
    } catch (error) {
        console.error('Error fetching ABI:', error.message);
        throw error;
    }
}

// Function to handle task execution based on argument type
async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, standardizedData } = task;
    const dynamicData = standardizedData.data.value; // Get standardized data

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);

    let args;
    let argTypeString;

    // Fetch the ABI dynamically
    const abi = await fetchABI(contractAddress);
    const contract = new ethers.Contract(contractAddress, abi, wallet);

    switch (argType) {
        case '0':
        case 'None':
            argTypeString = 'None';
            args = [];
            console.log('Executing without arguments.');
            break;

        case '1':
        case 'Static':
            argTypeString = 'Static';
            args = task.argumentInfo.arguments || [];
            console.log(`Executing with static arguments:`, args);
            break;

        case '2':
        case 'Dynamic':
            console.log(`Executing with dynamic arguments: ${standardizedData.data.value}`);
            args = [standardizedData.data.value]; // Use the value directly without conversion
            break;

        default:
            console.error('Unknown argument type!');
            throw new Error('Invalid argument type');
    }

    try {
        // Call the target function with the prepared arguments
        const result = await contract[targetFunction](...args);
        console.log(`Function ${targetFunction} executed successfully. Result:`, result);
        
        // If the result is a transaction response, wait for it to be mined
        // if (result.wait && typeof result.wait === 'function') {
        //     const receipt = await result.wait();
        //     console.log(`Transaction mined. Block number:`, receipt.blockNumber);
            
        //     // For state-changing functions, we might need to call a getter function to get the updated state
        //     // This is just an example, adjust according to your contract's structure
        //     const updatedState = await contract[`get${targetFunction.charAt(0).toUpperCase() + targetFunction.slice(1)}`]();
        //     return updatedState;
        // }
        
        // For non-state-changing functions, return the result directly
        return result;
    } catch (error) {
        console.error(`Error executing function ${targetFunction}:`, error.message);
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
        console.error('Error executing task:', error.message);
        res.status(500).send('Error executing task');
    }
});

// Start the keeper server
const keeperPort = 3001;
keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service listening at http://localhost:${keeperPort}`);
});