const express = require('express');
const axios = require('axios');
const { executeTask } = require('./keeperTimeExecution');

function createRPCClient(aggregatorUrl) {
    const app = express();
    app.use(express.json());

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
    

    app.post('/execute-task', fetchAndStandardizeData, async (req, res) => {
        const task = req.body;
        task.standardizedData = req.standardizedData;

        console.log('Keeper received task:', task);

        try {
            // This function should be imported from keeperTimeExecution.js
            const result = await executeTask(task);
            
            console.log(`Result: `, result);
            console.log(`-------------------------------------------------------------------------`);
            
            await axios.post(aggregatorUrl, { 
                jobId: task.jobId, 
                taskId: task.taskId,
                blockNumber: task.blockNumber,
                quorumNumbers: task.quorumNumbers,
                result: result
            });
            console.log('Task result sent to aggregator');

            res.status(200).send('Task executed and result sent to aggregator');
        } catch (error) {
            console.error('Error executing task:', error);
            res.status(500).json({
                error: 'Error executing task',
                details: error.message,
                stack: error.stack
            });
        }
    });

    return app;
}

module.exports = { createRPCClient };


// require('dotenv').config();
// const { ethers } = require('ethers');
// const TronWeb = require('tronweb');
// const fs = require('fs');
// const path = require('path');

// const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
// const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

// const tronWeb = new TronWeb({
//     fullHost: process.env.TRON_FULL_NODE,
//     privateKey: process.env.TRON_PRIVATE_KEY
// });

// const avsAddresses = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/contracts/avsAddresses.json'), 'utf8'));

// const serviceManagerAddress = avsAddresses.serviceManager;
// const taskManagerAddress = avsAddresses.taskManager;

// const serviceManagerABI = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/ServiceManager.json'), 'utf8'));
// const taskManagerAbi = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/TaskManager.json'), 'utf8'));

// const serviceManager = new ethers.Contract(serviceManagerAddress, serviceManagerABI, wallet);
// const taskManager = new ethers.Contract(taskManagerAddress, taskManagerAbi, wallet);

// const listenToTaskManager = async () => {
//     taskManager.on('NewTaskCreated', async (task) => {
//         console.log('Task created:', task);
        
//         // Step 4: Task Execution
//         const taskData = await taskManager.getTaskData(task);
//         console.log('Task data:', taskData);
        
//         const keeper = await getKeeper(taskData);
//         console.log('Keeper:', keeper);
        
//         await keeper.executeTask(taskData);
//     });
// }



// listenToTaskManager().catch(console.error);