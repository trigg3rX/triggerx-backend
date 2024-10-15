const axios = require('axios');
const { TronWeb } = require('tronweb');

// Initialize TronWeb (this should be moved to a config file in a real-world scenario)
const tronWeb = new TronWeb({
    fullHost: process.env.TRON_FULL_HOST,
    privateKey: process.env.TRON_PRIVATE_KEY
});

async function fetchABI(contractAddress) {
    try {
        const contract = await tronWeb.contract().at(contractAddress);
        return JSON.stringify(contract.abi, null, 2);
    } catch (error) {
        console.error("Error fetching ABI from TRON contract:", error);
        return null;
    }
}

async function executeTask(task) {
    const { jobId, jobType, contractAddress, targetFunction, argType, standardizedData } = task;

    console.log(`Executing task ${jobId} of type ${jobType} for contract ${contractAddress}`);

    const abi = await fetchABI(contractAddress);
    if (!abi) {
        throw new Error(`Failed to fetch ABI for contract ${contractAddress}`);
    }

    const baseAddress = tronWeb.address.fromHex(contractAddress);
    const contract = await tronWeb.contract().at(baseAddress);

    let args;
    switch (argType) {
        case 0: args = []; break;
        case 1: args = task.argumentInfo.arguments; break;
        case 2: args = [standardizedData.data.value]; break;
        default: throw new Error(`Invalid argument type: ${argType}`);
    }

    try {
        console.log(`Calling function ${targetFunction} with arguments:`, args);

        if (typeof contract[targetFunction] !== 'function') {
            throw new Error(`Function ${targetFunction} does not exist in the contract ABI`);
        }

        const method = contract[targetFunction](...args);
        const callerAddress = tronWeb.defaultAddress.base58;
        if (!callerAddress) {
            throw new Error('TronWeb default address is not set');
        }

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

module.exports = { executeTask };