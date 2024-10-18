const axios = require('axios');
const { ethers } = require('ethers');
//const { TronWeb } = require('tronweb');

// Initialize TronWeb (this should be moved to a config file in a real-world scenario)
// const tronWeb = new TronWeb({
//     fullHost: process.env.TRON_FULL_HOST,
//     privateKey: process.env.TRON_PRIVATE_KEY
// });

const provider = new ethers.JsonRpcProvider(process.env.ETHEREUM_RPC_URL);
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY);

async function fetchABI(contractAddress) {
    console.log("API Key: ", process.env.ETHERSCAN_API_KEY);
    console.log("Contract Address: ", contractAddress);
    const url = `https://api-sepolia-optimism.etherscan.io/api?module=contract&action=getabi&address=${contractAddress}&apikey=${process.env.ETHERSCAN_API_KEY}`;
    
    try {
        const response = await fetch(url);
        const data = await response.json();
        if (data.status === '1') {
            const abi = JSON.parse(data.result);
            console.log('ABI:', abi);
            return abi;
        } else {
            throw new Error(`Error fetching ABI: ${data.message}`);
        }
    } catch (error) {
        console.error('Fetch error:', error);
        throw error;
    }
}

async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, standardizedData } = task;
    const dynamicData = standardizedData.data.value;

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);

    let args;
    let argTypeString;

    // Fetch the ABI dynamically
    const abi = await fetchABI(contractAddress);
    console.log('Fetched ABI:', JSON.stringify(abi, null, 2));

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
            args = task.argumentInfo.arguments;
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
        console.error('Error details:', {
            message: error.message,
            code: error.code,
            method: error.method,
            transaction: error.transaction,
        });

        // If the error is due to execution reverted, try to get the revert reason
        if (error.code === 'CALL_EXCEPTION') {
            try {
                const tx = await contract.populateTransaction[targetFunction](...args);
                const result = await provider.call(tx);
                console.error('Revert reason:', result);
            } catch (revertError) {
                console.error('Failed to get revert reason:', revertError.message);
            }
        }

        throw error;
    }
}


module.exports = { executeTask };