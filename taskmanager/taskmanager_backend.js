require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const TronWeb = require('tronweb');
const taskManagerABI = require('./taskManagerABI.json');
const keeperConfigs = require('./keeperConfig');

// Addresses for smart contracts
const jobCreatorAddress = 'TAjmTb3v6FDEQyxktBn9heYjSt5VGeNMVr';  // Tron contract address on Nile
const taskManagerAddress = '0xa3aB4285c28b5B444ccc55d0F70f6ba5001a48B5';  // Ethereum contract address on Holesky
let jobCreatorContract;
let taskManagerContract;

// Express app setup
const app = express();
const port = 3000;
const keeperPort = 3005;
app.use(express.json());

// TronWeb initialization
function initializeWallets() {
    // Initialize TronWeb
    const tronWeb = new TronWeb({
        fullHost: 'https://nile.trongrid.io',
        privateKey: process.env.NILE_PRIVATE_KEY
    });

    if (!tronWeb) {
        console.error("!!! Tron Wallet not initialization failed.");
        process.exit(1);
    }

    // Initialize Ethereum Holesky wallet
    const holeskyProvider = new ethers.JsonRpcProvider('https://eth-holesky.g.alchemy.com/v2/9eCzjtGExJJ6c_WwQ01h6Hgmj8bjAdrc');
    let holeskyWallet;
    try {
        holeskyWallet = new ethers.Wallet(process.env.HOLESKY_PRIVATE_KEY, holeskyProvider);
    } catch (error) {
        console.error("!!! Holesky wallet initialization failed:", error.message);
        process.exit(1);
    }

    console.log(">>> Wallets initialized.");

    return { tronWeb, holeskyWallet };
}

async function initializeContracts() {
    // Fetch JobCreator ABI
    let jobCreatorABI;
    try {
        jobCreatorContract = await tronWeb.contract().at(jobCreatorAddress);
        jobCreatorABI = JSON.stringify(jobCreatorContract.abi);
    } catch (error) {
        console.error("!!! Error fetching JobCreator ABI:", error);
        throw error;
    }

    // Create TaskManager contract instance
    taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, holeskyWallet);

    console.log(">>> Contracts initialized.");
}

const { tronWeb, holeskyWallet } = initializeWallets();
// console.log(taskManagerContract);
const activeJobs = {};

function initializeKeepers() {
    const keepers = keeperConfigs.map(config => {
        return {
            id: config.id,
            port: config.port,
            publicKey: config.publicKey,
            privateKey: config.privateKey,
            trx: config.trx
        };
    });
    return keepers;
}

const keepers = initializeKeepers();
console.log(">>> Keepers initialized:", keepers);

// Function to listen for JobCreated events on the Nile network
async function listenForJobCreatedEvents() {
    await tronWeb.contract().at(jobCreatorAddress).then(contract => {
        contract.JobCreated().watch((err, event) => {
            if (err) return console.error("!!! Error with JobCreated event:", err);

            // Log the entire event object for debugging
            console.log(">>> JobCreated event received:", JSON.stringify(event, null, 2));

            const jobId = event.result.jobId;  // Changed from toNumber() to direct access
            console.log(">>> New job created: #", jobId);

            if (verifyJobData(jobId)) {
                createTasks(jobId);
            }
        });
    });

    console.log("Listening for JobCreated events...");
}

async function verifyJobData(jobId) {
    console.log('Verifying job data for jobID: #', jobId);

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

    if (arguments.length > 0) {
        console.log(`>>> JobID #${jobId} has valid arguments.`);
        return true;
    } else {
        console.error(`!!! JobID #${jobId} has no valid arguments.`);
        return false;
    }
}

async function createTasks(jobId) {
    try {
        console.log(taskManagerContract);
        // Create task in the Ethereum TaskManager contract
        const tx = await taskManagerContract.createTask(jobId);
        const receipt = await tx.wait();

        // Get the taskId from the emitted event
        const taskCreatedEvent = receipt.logs.find(log => log.eventName === 'TaskCreated');
        const taskId = taskCreatedEvent.args.taskId;    

        console.log(`>>> New task created for jobID #${jobId} with ID: #${taskId}.`);

        // Call the JobCreator contract to add the taskId
        try {
            const result = await jobCreatorContract.addTaskId(jobId, taskId).send();
            console.log(`>>> Added taskID #${taskId} to jobID #${jobId}.`);
        } catch (error) {
            console.error(`!!! Error adding taskID #${taskId} to jobID #${jobId}:`, error);
        }

        // Return the taskId for further use if needed
        createTaskData(jobId, taskId);
    } catch (error) {
        console.error("!!! Error creating task:", error);
        throw error;
    }
}

// Function to create a task in the TaskManager contract and send it to the keeper
async function createTaskData(jobId, taskId) {
    try {
        const encodedJobData = await jobCreatorContract.getJob(jobId).call();
        console.log("Raw job data:", encodedJobData);

        const decodedJob = decodeJobData(encodedJobData);
        console.log("Decoded job data:", decodedJob);

        const taskData = {
            taskId: taskId.toString(),
            jobId: jobId.toString(),
            jobType: decodedJob.jobType,
            contractAddress: decodedJob.contractAddress,
            targetFunction: decodedJob.targetFunction,
            argType: decodedJob.argType,
            argumentInfo: {
                type: decodedJob.argType,
                arguments: decodedJob.arguments
            },
            apiEndpoint: decodedJob.apiEndpoint,
            timeInterval: decodedJob.timeInterval
        };

        console.log("Structured task data:", taskData);

        // Ensure timeInterval is at least 1 second
        const timeIntervalSeconds = Math.max(1, decodedJob.timeInterval);
        const cronExpression = `*/${timeIntervalSeconds} * * * * *`;
        console.log("Cron expression:", cronExpression);

        const task = cron.schedule(cronExpression, () => {
            sendTaskToKeeper(taskData);
        }, {
            scheduled: true,
            timezone: "UTC"
        });

        activeJobs[taskId] = task;
        console.log(`>>> Task #${taskData.taskId} scheduled with cron: ${cronExpression}`);

        await sendTaskToKeeper(taskData);
    } catch (error) {
        console.error("!!! Error creating task and sending to keeper:", error);
    }
}

function decodeJobData(encodedJobData) {
    const abiTypes = [
        "uint256", // jobId
        "string",  // jobType
        "string",  // status
        "uint32",  // timeframe
        "uint256", // blockNumber
        "address", // contractAddress
        "string",  // targetFunction
        "uint256", // timeInterval
        "uint8",   // argType
        "bytes[]", // arguments
        "string",  // apiEndpoint
        "uint32[]" // taskIds
    ];

    const decodedJob = tronWeb.utils.abi.decodeParams(abiTypes, encodedJobData);

    return {
        jobId: decodedJob[0] ? decodedJob[0].toString() : '',
        jobType: decodedJob[1] || '',
        status: decodedJob[2] || '',
        timeframe: Number(decodedJob[3] || 0),
        blockNumber: decodedJob[4] ? decodedJob[4].toString() : '',
        contractAddress: tronWeb.address.fromHex(decodedJob[5]),
        targetFunction: decodedJob[6] || '',
        timeInterval: decodedJob[7] ? Number(decodedJob[7]) : 0,
        argType: decodedJob[8] || 0,
        arguments: Array.isArray(decodedJob[9]) ? decodedJob[9] : [],
        apiEndpoint: decodedJob[10] || '',
        taskIds: Array.isArray(decodedJob[11]) 
            ? decodedJob[11].map(id => (typeof id === 'bigint' ? id.toString() : Number(id))) 
            : []
    };
}



// Helper function to convert nested BigInt types before sending the task data
function getRandomKeeper() {
    const randomIndex = Math.floor(Math.random() * keepers.length);
    return keepers[randomIndex];
}

// Function to send task to a random keeper
async function sendTaskToKeeper(taskData) {
    const keeper = getRandomKeeper(); // Select a random keeper
    const keeperUrl = `http://localhost:${keeper.port}/execute-task`; // Keeper URL based on its port

    const convertNestedBigInt = (obj) => {
        if (typeof obj === 'bigint') {
            return obj.toString();
        } else if (Array.isArray(obj)) {
            return obj.map(convertNestedBigInt);
        } else if (typeof obj === 'object' && obj !== null) {
            return Object.fromEntries(
                Object.entries(obj).map(([key, value]) => [key, convertNestedBigInt(value)]));
        }
        return obj;
    };

    const convertedTaskData = convertNestedBigInt(taskData);
    console.log(`Sending task to keeper #${keeper.id} on port ${keeper.port}`);

    try {
        const response = await axios.post(keeperUrl, convertedTaskData);
        console.log(`>>> Task sent to keeper. Response: ${response.status} ${response.statusText}`);
    } catch (error) {
        console.error("!!! Error sending task to keeper:", error.message);
    }
}
// Start the Express server and begin listening for JobCreated events
app.listen(port, () => {
    initializeContracts();
    console.log(`Task Manager backend listening on port ${port}`);
    listenForJobCreatedEvents();
});