require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const TronWeb = require('tronweb');
const taskManagerABI = require('./taskManagerABI.json');

// Addresses for smart contracts
const jobCreatorAddress = 'TJFouw53gKu5btB6MZ7kvsZ3NesefXgxTS';  // Tron contract address on Nile
const taskManagerAddress = '0x781E170238288e57B08F269d6714E3a28dc345A8';  // Ethereum contract address on Holesky

// Express app setup
const app = express();
const port = 3000;
const keeperPort = 3001;
app.use(express.json());

// TronWeb initialization
const tronWeb = new TronWeb({
    fullHost: 'https://nile.trongrid.io',
    privateKey: process.env.NILE_PRIVATE_KEY
});

if (!tronWeb.defaultAddress.base58) {
    console.error('TronWeb not properly initialized. Make sure NILE_PRIVATE_KEY is set correctly.');
    process.exit(1);
}

// Ethereum Holesky testnet provider and Task Manager contract instance
const holeskyProvider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
const holeskyWallet = new ethers.Wallet(process.env.HOLESKY_PRIVATE_KEY, holeskyProvider);

// TaskManager contract instance
const taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, holeskyWallet);

// Store active jobs to avoid duplicating tasks
const activeJobs = {};

// Function to listen for JobCreated events on the Nile network
async function listenForJobCreatedEvents() {
    const eventListener = await tronWeb.contract().at(jobCreatorAddress).then(contract => {
        contract.JobCreated().watch((err, event) => {
            if (err) return console.error('Error with JobCreated event:', err);

            // Log the entire event object for debugging
            console.log('JobCreated event received:', JSON.stringify(event, null, 2));

            const jobId = event.result.jobId;  // Changed from toNumber() to direct access
            console.log('New job created:', jobId);
            processJob(jobId);
        });
    });

    console.log('Listening for JobCreated events...');
}

// Function to schedule a job with cron
function scheduleJob(job) {
    const { jobId, timeInterval } = job;
    const intervalInSeconds = Math.max(Number(timeInterval), 10); // Ensure a minimum interval of 10 seconds

    const cronExpression = `*/${intervalInSeconds} * * * * *`;

    // Schedule the job with cron
    const task = cron.schedule(cronExpression, () => {
        createTaskAndSendToKeeper(job);
    }, {
        scheduled: true,
        timezone: "UTC"
    });

    activeJobs[jobId] = task;
    console.log(`Job ${jobId} scheduled with cron: ${cronExpression}`);
}

// Function to create a task in the TaskManager contract and send it to the keeper
async function createTaskAndSendToKeeper(job) {
    try {
        // Create task in the Ethereum TaskManager contract
        const tx = await taskManagerContract.createTask(job.jobType, "Created");
        const receipt = await tx.wait();
        console.log(`Task created for job ${job.jobId}, transaction hash: ${receipt.hash}`);

        // Fetch the task count from the contract (this will be the new taskId)
        const taskId = await taskManagerContract.taskCount();

        // Prepare task data to send to the keeper
        const taskData = {
            taskId: taskId.toString(),
            jobId: job.jobId.toString(),
            jobType: job.jobType,
            contractAddress: job.contractAddress,
            targetFunction: job.targetFunction,
            argType: job.argType,
            argumentInfo: {
                type: job.argType,
                arguments: job.arguments
            },
            apiEndpoint: job.apiEndpoint
        };

        // Send the task data to the keeper for execution
        await sendTaskToKeeper(taskData);
    } catch (error) {
        console.error('Error creating task and sending to keeper:', error);
    }
}

// Helper function to convert nested BigInt types before sending the task data
async function sendTaskToKeeper(taskData) {
    const keeperUrl = `http://localhost:${keeperPort}/execute-task`;

    const convertNestedBigInt = (obj) => {
        if (typeof obj === 'bigint') {
            return obj.toString();
        } else if (Array.isArray(obj)) {
            return obj.map(convertNestedBigInt);
        } else if (typeof obj === 'object' && obj !== null) {
            return Object.fromEntries(
                Object.entries(obj).map(([key, value]) => [key, convertNestedBigInt(value)])
            );
        }
        return obj;
    };

    const convertedTaskData = convertNestedBigInt(taskData);
    console.log("TaskData after conversion:", convertedTaskData);

    try {
        const response = await axios.post(keeperUrl, convertedTaskData);
        console.log(`Task sent to keeper. Response: ${response.status} ${response.statusText}`);
    } catch (error) {
        console.error('Error sending task to keeper:', error.message);
    }
}

// Function to fetch and process job data from the Tron contract
async function processJob(jobId) {
    try {
        console.log('Fetching job with ID:', jobId);

        // Fetch the job details
        const job = await tronWeb.contract().at(jobCreatorAddress).then(contract => {
            return contract.getJob(jobId).call();
        });

        console.log('Raw job data fetched:', job);

        if (!job || Object.keys(job).length === 0) {
            console.error(`Job ${jobId} not found or empty.`);
            return;
        }

        // Check if the job is already active
        if (activeJobs[jobId]) {
            console.log(`Job ${jobId} is already active.`);
            return;
        }

        // Fetch the arguments for the job
        const argumentCount = await tronWeb.contract().at(jobCreatorAddress).then(contract => {
            return contract.getJobArgumentCount(jobId).call();
        });

        const arguments = [];
        for (let i = 0; i < argumentCount; i++) {
            const arg = await tronWeb.contract().at(jobCreatorAddress).then(contract => {
                return contract.getJobArgument(jobId, i).call();
            });
            arguments.push(arg);
        }

        // Only proceed if we have valid job data
        const processedJob = {
            jobId: job.jobId.toString(),
            jobType: job.jobType,
            status: job.status,
            timeInterval: job.timeInterval.toString(),
            timeframe: job.timeframe.toString(),
            blockNumber: job.blockNumber.toString(),
            contractAddress: job.contractAddress,
            targetFunction: job.targetFunction,
            argType: job.argType,
            arguments: arguments,
            apiEndpoint: job.apiEndpoint
        };
 
        console.log('Processed job data:', processedJob);
        
        // Schedule the job if it has valid arguments
        if (arguments.length > 0) {
            scheduleJob(processedJob);
        } else {
            console.error(`Job ${jobId} has no valid arguments.`);
        }
    } catch (error) {
        console.error('Error processing job:', error);
    }
}

// Start the Express server and begin listening for JobCreated events
app.listen(port, () => {
    console.log(`Task Manager backend listening on port ${port}`);
    listenForJobCreatedEvents();
});
