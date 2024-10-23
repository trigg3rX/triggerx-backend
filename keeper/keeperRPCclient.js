const express = require('express');
const axios = require('axios');
const { executeTask, fetchAndStandardizeAPIData } = require('./keeperTimeExecution');

function createRPCClient(aggregatorUrl) {
    const app = express();
    app.use(express.json());

    app.post('/execute-task', fetchAndStandardizeAPIData, async (req, res) => {
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
