const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const jobListingABI = require('./jobListingABI.json');

console.log('Script started');

const app = express();
const port = 3000;
const keeperPort = 3001; 
app.use(express.json());

console.log('Express app created');

// Address of the deployed JobListing contract
const jobListingAddress = '0xe94843C5fb22D6752049442Db3A03B7f8bfcAEe4'; 

// Connect to an Ethereum node 
const provider = new ethers.JsonRpcProvider('https://eth-sepolia.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');

console.log('Provider created');

// Create a contract instance
const jobListingContract = new ethers.Contract(jobListingAddress, jobListingABI, provider);

console.log('Contract instance created');

// Object to store active jobs
const activeJobs = {};

// Function to schedule a job
function scheduleJob(job) {
    const { jobId, timeInterval, timeframe } = job;
    const intervalInSeconds = Number(timeInterval);
    const timeframeInSeconds = Number(timeframe);

    if (intervalInSeconds < 60) {
        console.warn(`Job ${jobId} has an interval less than 1 minute. Setting to 1 minute.`);
    }

    const cronExpression = `*/${Math.max(1, Math.floor(intervalInSeconds / 60))} * * * *`;

    const task = cron.schedule(cronExpression, () => {
        sendJobToKeeper(job);
    });

    // Schedule job termination
    setTimeout(() => {
        task.stop();
        delete activeJobs[jobId];
        console.log(`Job ${jobId} completed its timeframe and has been terminated.`);
    }, timeframeInSeconds * 1000);

    activeJobs[jobId] = task;
    console.log(`Job ${jobId} scheduled with cron: ${cronExpression}`);
}

// Function to send a job to the keeper
async function sendJobToKeeper(job) {
    const keeperUrl = `http://localhost:${keeperPort}/execute-task`;
    const taskData = {
        jobId: job.jobId.toString(),
        jobType: job.jobType,
        contractAddress: job.contract_add,
        targetFunction: job.target_fnc,
        argType: job.argType, 
        argumentInfo: {
            type: job.argType, 
            details: "Placeholder for argument details" 
        }
    };

    try {
        const response = await axios.post(keeperUrl, taskData);
        console.log(`Task sent to keeper. Response: ${response.status} ${response.statusText}`);
    } catch (error) {
        console.error('Error sending task to keeper:', error.message);
    }
}

// Function to process a new job
async function processJob(jobId) {
    try {
        const job = await jobListingContract.getJob(jobId);

        if (activeJobs[jobId]) {
            console.log(`Job ${jobId} is already active.`);
            return;
        }

        scheduleJob(job);

    } catch (error) {
        console.error('Error processing job:', error);
    }
}

// Listen for JobCreated events
jobListingContract.on('JobCreated', (jobId, jobType, contractAdd, timeInterval, event) => {
    console.log('New job created:', jobId.toString());
    processJob(jobId);
});

console.log('Event listener set up');

// API endpoint to manually trigger job processing
app.get('/process-job/:jobId', async (req, res) => {
    const jobId = req.params.jobId;
    await processJob(jobId);
    res.send(`Processing job ${jobId}`);
});

// Start the server
app.listen(port, () => {
    console.log(`Job keeper backend listening at http://localhost:${port}`);
});

// Keeper service (for demonstration purposes)
const keeperApp = express();
keeperApp.use(express.json());

keeperApp.post('/execute-task', (req, res) => {
    const task = req.body;
    console.log('Keeper received task:', task);
    res.status(200).send('Task received and will be executed');
});

keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service listening at http://localhost:${keeperPort}`);
});

console.log('Script setup complete');