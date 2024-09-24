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

    if (intervalInSeconds < 10) {
        console.warn(`Job ${jobId} has an interval less than 10 seconds. Setting to 10 seconds.`);
    }

    const cronExpression = `*/${Math.max(10, intervalInSeconds)} * * * * *`; // Use seconds granularity

    const task = cron.schedule(cronExpression, () => {
        sendJobToKeeper(job);
    }, {
        scheduled: true,
        timezone: "UTC" // optional: set timezone if necessary
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


function convertBigIntToString(obj) {
    if (typeof obj === 'bigint') {
        return obj.toString();
    } else if (Array.isArray(obj)) {
        return obj.map(convertBigIntToString);
    } else if (typeof obj === 'object' && obj !== null) {
        return Object.fromEntries(
            Object.entries(obj).map(([key, value]) => [key, convertBigIntToString(value)])
        );
    }
    return obj;
}

// Function to send a job to the keeper
async function sendJobToKeeper(job) {
    const keeperUrl = `http://localhost:${keeperPort}/execute-task`;

    const taskData = convertBigIntToString({
        jobId: job.jobId.toString(),
        jobType: job.jobType,
        contractAddress: job.contract_add,
        targetFunction: job.target_fnc,
        argType: job.argType, 
        argumentInfo: {
            type: job.argType, 
            details: "Placeholder for argument details"
        }
    });

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

console.log('Script setup complete');
