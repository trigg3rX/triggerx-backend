const axios = require('axios');
const { TronWeb } = require('tronweb');

async function fetchAndStandardizeAPIData(req, res, next) {
    const { argType, apiEndpoint } = req.body;

    if (argType !== 'Dynamic') {
        req.standardizedData = { data: { value: null } };
        return next();
    }

    try {
        const response = await axios.get(apiEndpoint);
        req.standardizedData = response.data;
        next();
    } catch (error) {
        console.error(`Error fetching data from ${apiEndpoint}:`, error.message);
        req.standardizedData = { data: { value: null } };
        next();
    }
}

async function executeTask(task) {
    const tronWeb = new TronWeb({
        fullHost: process.env.TRON_FULL_HOST,
        privateKey: process.env.KEEPER_PRIVATE_KEY
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
            args = [standardizedData.data.value];
            break;
        default: 
            throw new Error(`Invalid argument type: ${argType}`);
    }

    try {
        // console.log(`Calling function ${targetFunction} with arguments:`, args);

        const method = contract[targetFunction](...args);
        const callerAddress = tronWeb.defaultAddress.base58;
        
        let result = await method.send({
            from: callerAddress,
            callValue: 1
        });

        const serializedResult = result.toString();

        // console.log(`Function call ${targetFunction} executed successfully.`);
        // console.log(`---------------------------------------------------------------------------------------------------------------`);
        return serializedResult;
    } catch (error) {
        console.error(`Error executing function ${targetFunction}:`, error);
        throw error;
    }
}

module.exports = { executeTask, fetchAndStandardizeAPIData };