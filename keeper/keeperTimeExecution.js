const axios = require('axios');
const { TronWeb } = require('tronweb');

async function fetchAndStandardizeAPIData(req, res, next) {
    const { argType, apiEndpoint } = req.body;

    if (argType !== 'Dynamic') {
        req.standardizedData = { data: { value: '0' } };
        return next();
    }

    try {
        const response = await axios.get(apiEndpoint);
        
        if (!response.data?.data?.value) {
            throw new Error('Invalid API response format');
        }
        
        // Ensure the value is converted to a string to avoid BigNumber issues
        const value = response.data.data.value.toString();
        
        req.standardizedData = {
            data: {
                value: value
            }
        };
        next();
    } catch (error) {
        console.error(`Error fetching data from ${apiEndpoint}:`, error.message);
        req.standardizedData = { data: { value: '0' } };
        next();
    }
}

async function executeTask(task) {
    const fullNode = process.env.TRON_FULL_HOST;
    const tronWeb = new TronWeb({
        fullHost: fullNode,
        solidityNode: fullNode,
        eventServer: fullNode,
        privateKey: process.env.TRON_PRIVATE_KEY
    });

    const { taskId, jobId, jobType, contractAddress, targetFunction, argType, argumentInfo, apiEndpoint, standardizedData } = task;

    console.log(`>>> Executing Task #${taskId} of Job #${jobId} for contract ${contractAddress}`);

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
            // Handle the value as a string and ensure it's properly formatted
            const value = standardizedData.data.value || '0';
            // Convert to string first, then to BigNumber to ensure proper formatting
            const bigNumberValue = TronWeb.toHex(value);
            args = [bigNumberValue];
            break;
        default: 
            throw new Error(`Invalid argument type: ${argType}`);
    }

    try {
        const method = contract[targetFunction](...args);
        const callerAddress = tronWeb.defaultAddress.base58;
        
        let result = await method.send({
            from: callerAddress,
            callValue: 1
        });

        const serializedResult = result.toString();
        return serializedResult;
    } catch (error) {
        console.error(`Error executing function ${targetFunction}:`, error);
        throw error;
    }
}

module.exports = { executeTask, fetchAndStandardizeAPIData };