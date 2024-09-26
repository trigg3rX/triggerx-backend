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
const jobListingAddress = '0x5edB869670B0FF939F4Ad28972b4329af8f5ea33'; 

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


// function convertBigIntToString(obj) {
//     if (typeof obj === 'bigint') {
//         return obj.toString();
//     } else if (Array.isArray(obj)) {
//         return obj.map(convertBigIntToString);
//     } else if (typeof obj === 'object' && obj !== null) {
//         return Object.fromEntries(
//             Object.entries(obj).map(([key, value]) => [key, convertBigIntToString(value)])
//         );
//     }
//     return obj;
// }

// Function to send a job to the keeper
async function sendJobToKeeper(job) {
    const keeperUrl = `http://localhost:${keeperPort}/execute-task`;
    
    // Prepare taskData
    const taskData = {
        jobId: job.jobId.toString(),
        jobType: job.jobType,
        contractAddress: job.contractAddress,
        targetFunction: job.targetFunction,
        argType: job.argType,  
        argumentInfo: {
            type: job.argType,
            arguments: job.arguments
        },
        apiEndpoint: job.apiEndpoint  // Add apiEndpoint to the task data
    };

    // Convert any BigInt in the taskData object to strings
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

    // Convert all BigInts in taskData
    const convertedTaskData = convertNestedBigInt(taskData);
    console.log("TaskData after conversion:", convertedTaskData);

    try {
        const response = await axios.post(keeperUrl, convertedTaskData);
        console.log(`Task sent to keeper. Response: ${response.status} ${response.statusText}`);
    } catch (error) {
        console.error('Error sending task to keeper:', error.message);
    }
}



// Function to fetch job arguments
async function fetchJobArguments(jobId) {
    const argumentCount = await jobListingContract.getJobArgumentCount(jobId);
    const arguments = [];

    for (let i = 0; i < argumentCount; i++) {
        const arg = await jobListingContract.getJobArgument(jobId, i);
        arguments.push(arg);
    }

    return arguments;
}

// Function to process a new job
async function processJob(jobId) {
    try {
        const job = await jobListingContract.getJob(jobId);
        console.log('Job fetched:', job); // Log the entire job object

        if (activeJobs[jobId]) {
            console.log(`Job ${jobId} is already active.`);
            return;
        }

        // Fetch job arguments
        const jobArguments = await fetchJobArguments(jobId);
        job.arguments = jobArguments.map(arg => arg.toString());

        // Log job properties to see their structure
        console.log('Job properties:', {
            jobId: job[0].toString(),
            jobType: job[1],
            status: job[2],
            createdBy: job[3],
            timeInterval: job[4].toString(),
            timeframe: job[5].toString(),
            blockNumber: job[6].toString(),
            contractAddress: job[8],
            targetFunction: job[9] 
        });

        // Extract contract address and target function
        const contractAddress = job[8]?.toString(); // Convert to string if defined
        const targetFunction = job[9]; // Target function can remain as is

        console.log(`Contract Address: ${contractAddress}`); // Log the contract address
        console.log(`Target Function: ${targetFunction}`); // Log the target function

        if (!contractAddress || !targetFunction) {
            console.error('Contract address or target function is missing:', job);
            return; // Prevent sending the task
        }

        // Assign contract address and target function to job object
        // job.contract_add = contractAddress;
        // job.target_fnc = targetFunction;

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
