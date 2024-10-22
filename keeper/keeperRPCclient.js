const express = require('express');
const axios = require('axios');
const { executeTask } = require('./keeperTimeExecution');

function createRPCClient(aggregatorUrl) {
    const app = express();
    app.use(express.json());

    async function fetchAndStandardizeData(req, res, next) {
        const { argType, apiEndpoint } = req.body;

        if (argType !== 'Dynamic') {
            req.standardizedData = { data: { value: null } };
            return next();
        }

        try {
            const response = await axios.get(apiEndpoint);
            // Implement standardization logic here
            req.standardizedData = response.data;
            next();
        } catch (error) {
            console.error(`Error fetching data from ${apiEndpoint}:`, error.message);
            req.standardizedData = { data: { value: null } };
            next();
        }
    }

    app.post('/execute-task', fetchAndStandardizeData, async (req, res) => {
        const task = req.body;
        task.standardizedData = req.standardizedData;

        console.log('>>> Keeper received task:', task);

        try {
            // This function should be imported from keeperTimeExecution.js
            const result = await executeTask(task);
            
            console.log(`Result: `, result);
            console.log(`---------------------------------------------------------------------------------------------------------------`);
            
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
