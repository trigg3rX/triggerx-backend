const axios = require('axios');
const { ethers } = require('ethers');

const provider = new ethers.JsonRpcProvider(process.env.OP_SEPOLIA_RPC_URL);
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

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

// Helper function to convert BigInt to string or number
function convertBigIntResult(result) {
    if (typeof result === 'bigint') {
        // Convert BigInt to string to preserve exact value
        return result.toString();
    } else if (Array.isArray(result)) {
        // Handle arrays of results
        return result.map(item => convertBigIntResult(item));
    } else if (typeof result === 'object' && result !== null) {
        // Handle objects
        const convertedResult = {};
        for (const key in result) {
            convertedResult[key] = convertBigIntResult(result[key]);
        }
        return convertedResult;
    }
    return result;
}

async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, argumentInfo, standardizedData } = task;

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);
    
    // Check contract deployment
    const code = await provider.getCode(contractAddress);
    console.log('Contract bytecode:', code);
    if (code === '0x' || code === '0x0') {
        throw new Error(`No contract found at address ${contractAddress}`);
    }

    // Fetch ABI
    const abi = await fetchABI(contractAddress);
    
    // Find the function in the ABI
    const functionABI = abi.find(item => item.name === targetFunction);
    if (!functionABI) {
        throw new Error(`Function ${targetFunction} not found in ABI`);
    }

    // Log function details for debugging
    console.log('Function ABI:', functionABI);
    console.log('Required parameters:', functionABI.inputs);
    console.log('Argument type:', argType);
    console.log('Argument info:', argumentInfo);

    // Prepare arguments based on the function ABI and input type
    let args = [];
    
    const getDefaultValue = (input) => {
        switch (input.type) {
            case 'uint256':
            case 'int256':
                return 1;
            case 'string':
                return '';
            case 'bool':
                return false;
            default:
                return 0;
        }
    };

    switch (argType) {
        case '0':
        case 'None':
            args = functionABI.inputs.map(input => getDefaultValue(input));
            console.log('Using default arguments:', args);
            break;

        case '1':
        case 'Static':
            if (!argumentInfo || !argumentInfo.arguments) {
                args = functionABI.inputs.map(input => getDefaultValue(input));
                console.log('No valid arguments provided, using defaults:', args);
            } else if (argumentInfo.arguments.length === 0 && functionABI.inputs.length > 0) {
                args = functionABI.inputs.map(input => getDefaultValue(input));
                console.log('Empty arguments array provided, using defaults:', args);
            } else {
                args = argumentInfo.arguments;
                console.log('Using provided static arguments:', args);
            }
            break;

        case '2':
        case 'Dynamic':
            if (standardizedData?.data?.value !== null && standardizedData?.data?.value !== undefined) {
                args = [standardizedData.data.value];
            } else {
                args = functionABI.inputs.map(input => getDefaultValue(input));
            }
            console.log('Using dynamic arguments:', args);
            break;
    }

    // Create contract instance
    const contract = new ethers.Contract(contractAddress, abi, wallet);

    try {
        // Validate argument count
        if (args.length !== functionABI.inputs.length) {
            throw new Error(`Invalid number of arguments. Expected ${functionABI.inputs.length}, got ${args.length}`);
        }

        console.log(`Executing ${targetFunction} with arguments:`, args);

        // Try read-only call first
        try {
            const result = await contract[targetFunction].staticCall(...args);
            console.log('Read-only call successful. Result:', result);
            // Convert BigInt result to string before returning
            const convertedResult = convertBigIntResult(result);
            return convertedResult;
        } catch (staticCallError) {
            console.log('Static call failed, attempting state-changing call...');
            const tx = await contract[targetFunction](...args);
            console.log('Transaction sent:', tx.hash);
            const receipt = await tx.wait();
            console.log('Transaction confirmed:', receipt);
            // Convert any BigInt values in receipt
            const convertedReceipt = convertBigIntResult(receipt);
            return convertedReceipt;
        }
    } catch (error) {
        console.error('Execution error:', error);
        if (error.data) {
            try {
                const iface = new ethers.Interface(abi);
                const decodedError = iface.parseError(error.data);
                console.error('Decoded error:', decodedError);
            } catch (decodeError) {
                console.error('Could not decode error:', decodeError);
            }
        }
        throw error;
    }
}

module.exports = { executeTask };