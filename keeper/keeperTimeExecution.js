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
    const { taskId, jobId, jobType, contractAddress, targetFunction, argType, argumentInfo, apiEndpoint, standardizedData } = task;

    console.log(`Executing task ${taskId} of type ${jobType} for contract ${contractAddress}`);

    const baseAddress = tronWeb.address.fromHex(contractAddress);
    const contract = await tronWeb.contract().at(baseAddress);

    let args;
    switch (argType) {
        case '0':
        case 'None': 
            args = [];
            break;
        case '1':
        case 'Static': 
            args = argumentInfo;
            break;
        case '2':
        case 'Dynamic': 
            args = [standardizedData.data.value];
            break;
        default: 
            throw new Error(`Invalid argument type: ${argType}`);
    }

    try {
        console.log(`Calling function ${targetFunction} with arguments:`, args);

        const method = contract[targetFunction](...args);
        const callerAddress = tronWeb.defaultAddress.base58;
        
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