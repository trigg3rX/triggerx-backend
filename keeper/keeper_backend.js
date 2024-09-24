const express = require('express');
const axios = require('axios');
const { ethers } = require('ethers');

const keeperApp = express();
keeperApp.use(express.json());

const aggregatorUrl = 'http://localhost:3002/receive-result'; 
const etherscanApiKey = 'V332PZUHEE7V97ZP2V7YV3YPA6RXNR4XEV'; 
const provider = new ethers.JsonRpcProvider('https://eth-sepolia.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc'); // Connect to Ethereum node

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
    const { jobId, jobType, contractAddress, targetFunction, argType, argumentInfo } = task;
    
    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);
    
    let args;
    // Handle numeric argTypes (0 for None, 1 for Static, 2 for Dynamic)
    let argTypeString;
    switch (argType) {
        case '0':
        case 'None':
            argTypeString = 'None';
            args = [];
            console.log('Executing without arguments.');

            // Fetch the ABI dynamically
            const abi = await fetchABI(contractAddress);
            const contract = new ethers.Contract(contractAddress, abi, provider);

            // Call the target function without arguments
            try {
                const result = await contract[targetFunction](); // Execute function with no arguments
                console.log(`Function ${targetFunction} executed successfully:`, result);
                return result;
            } catch (error) {
                console.error(`Error executing function ${targetFunction}:`, error.message);
                throw error;
            }

        case '1':
        case 'Static':
            argTypeString = 'Static';
            args = argumentInfo.details; // Static argument details provided
            console.log(`Executing with static arguments: ${args}`);
            break;

        case '2':
        case 'Dynamic':
            argTypeString = 'Dynamic';
            // For dynamic arguments, fetch data from an external source
            const dynamicData = await fetchDynamicData();
            args = [dynamicData];
            console.log(`Executing with dynamic arguments: ${dynamicData}`);
            break;

        default:
            console.error('Unknown argument type!');
            throw new Error('Invalid argument type');
    }

    // Simulate executing the target function on the contract (in reality, you'd interact with the blockchain)
    const result = await simulateContractInteraction(contractAddress, targetFunction, args);
    
    return result;
}


// Simulated contract interaction (replace this with actual contract interaction if needed)
async function simulateContractInteraction(contractAddress, targetFunction, args) {
    console.log(`Simulated contract call to ${contractAddress} -> ${targetFunction} with args:`, args);
    return { success: true, message: 'Task executed successfully', args };
}

// Fetch dynamic data (e.g., price from CoinGecko)
async function fetchDynamicData() {
    try {
        const response = await axios.get('https://api.coingecko.com/api/v3/simple/price?ids=usd-coin&vs_currencies=usd');
        return response.data['usd-coin'].usd;
    } catch (error) {
        console.error('Error fetching dynamic data:', error.message);
        throw error;
    }
}

// Endpoint to receive tasks
keeperApp.post('/execute-task', async (req, res) => {
    const task = req.body;
    console.log('Keeper received task:', task);

    try {
        const result = await executeTask(task);
        
        // Send the result to the aggregator
        await axios.post(aggregatorUrl, { jobId: task.jobId, result });
        console.log('Task result sent to aggregator:', result);

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
