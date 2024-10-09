require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const TronWeb = require('tronweb');
const taskManagerABI = require('./taskManagerABI.json');

// Addresses for smart contracts
const jobCreatorAddress = process.env.CONTRACT_ADDRESS;  // Tron contract address on Nile
const taskManagerAddress = '0xDaa3d01f71F638952db924c9FE4f1CDa847A23Ad';  // Ethereum contract address on Holesky
let jobCreatorContract;
let taskManagerContract;

// Express app setup
const app = express();
const port = 3000;
const keeperPort = 3001;
app.use(express.json());

// TronWeb initialization
function initializeWallets() {
    // Initialize TronWeb
    const tronWeb = new TronWeb({
        fullHost: 'https://nile.trongrid.io',
        privateKey: process.env.TRON_PRIVATE_KEY
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
        // Create task in the Ethereum TaskManager contract
        const tx = await taskManagerContract.createNewTask(jobId, '0x01');
        const receipt = await tx.wait();

        // Log the entire receipt to inspect the structure
        console.log("Transaction receipt logs:", JSON.stringify(receipt.logs, null, 2));

        // Search for the 'TaskCreated' event
        const taskCreatedEvent = receipt.logs.find(log => log.eventName === 'NewTaskCreated');

        // If taskCreatedEvent is not found, log an error
        if (!taskCreatedEvent) {
            throw new Error("TaskCreated event not found in logs.");
        }

        const taskId = taskCreatedEvent.args.latestTasknum;    
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
        const encodedJobData = await jobCreatorContract.jobs(jobId).call();
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
        "uint32[]",// taskIds
        "address", // creator address
        "uint256"  // stake amount
    ];

    const decodedJob = tronWeb.utils.abi.decodeParams(abiTypes, encodedJobData);

    if (Array.isArray(encodedJobData)) {
        return {
            jobId: encodedJobData[0] ? encodedJobData[0].toString() : '',
            jobType: encodedJobData[1] || '',
            status: encodedJobData[2] || '',
            timeframe: Number(encodedJobData[3] || 0),
            blockNumber: encodedJobData[4] && encodedJobData[4]._isBigNumber ? 
                        encodedJobData[4].toString() : 
                        (encodedJobData[4] || '').toString(),
            contractAddress: typeof encodedJobData[5] === 'string' ? 
                            tronWeb.address.fromHex(encodedJobData[5]) : 
                            encodedJobData[5],
            targetFunction: encodedJobData[6] || '',
            timeInterval: encodedJobData[7] && encodedJobData[7]._isBigNumber ? 
                         Number(encodedJobData[7].toString()) : 
                         Number(encodedJobData[7] || 0),
            argType: Number(encodedJobData[8] || 0),
            arguments: encodedJobData[9] === 'null' ? [] : [encodedJobData[9]],
            apiEndpoint: encodedJobData[10] || '',
            creatorAddress: typeof encodedJobData[11] === 'string' ? 
                           tronWeb.address.fromHex(encodedJobData[11]) : 
                           encodedJobData[11],
            stakeAmount: encodedJobData[12] && encodedJobData[12]._isBigNumber ? 
                        encodedJobData[12].toString() : 
                        (encodedJobData[12] || '0').toString()
        };
    }

    try {
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
                : [],
            creatorAddress: tronWeb.address.fromHex(decodedJob[12]),
            stakeAmount: decodedJob[13] ? decodedJob[13].toString() : '0'
        };
    } catch (error) {
        console.error("Error decoding job data:", error);
        throw error;
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