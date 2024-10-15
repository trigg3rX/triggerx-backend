require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const { TronWeb } = require('tronweb');
const taskManagerABI = require('../utils/abi/TaskManager.json');
const { jobManagerABI } = require('../utils/abi/JobManager');
const keeperConfigs = require('../utils/keeperConfig');

// Addresses for smart contracts
const jobManagerAddress = 'TEsKaf2n8aF6pta7wyG5gwukzR4NoHre59';
const taskManagerAddress = '0xDaa3d01f71F638952db924c9FE4f1CDa847A23Ad';

let taskManagerContract;
let jobManagerContract;
let holeskyWallet;
let tronWeb;

const app = express();
const port = 3000;

app.use(express.json());

const activeJobs = {};

function initializeWallets() {
    tronWeb = new TronWeb({
        fullHost: process.env.TRON_FULL_HOST,
        privateKey: process.env.TRON_PRIVATE_KEY
    });

    const holeskyProvider = new ethers.JsonRpcProvider(process.env.ETHEREUM_RPC_URL);
    try {
        holeskyWallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, holeskyProvider);
    } catch (error) {
        console.error("!!! Holesky wallet initialization failed:", error.message);
        process.exit(1);
    }

    console.log(">>> Wallets initialized.");
}

async function initializeContracts() {
    try {
        jobManagerContract = await tronWeb.contract(jobManagerABI, jobManagerAddress);
    } catch (error) {
        console.error("!!! Error initializing JobManager:", error);
        throw error;
    }

    taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, holeskyWallet);

    console.log(">>> Contracts initialized.");
}

async function getEventsOfLatestBlock() {
    const events = await tronWeb.event.getEventsOfLatestBlock({
        only_confirmed: false
    });

    if (events.data.length > 0) {
        console.log(events.data[0].block_number);
    } else {
        console.log('No events');
    }
    return events.data.filter(event => event.contract_address === jobManagerAddress);
}

async function listenForJobManagerEvents() {
    console.log(`JobManager listener running on port ${port}`);

    setInterval(async () => {
        const jobManagerEvents = await getEventsOfLatestBlock();
        
        for (const event of jobManagerEvents) {
            console.log(event.block_number);
            if (event.event_name === 'JobCreated') {
                const jobId = event.result.jobId;
                console.log(">>> New job created: #", jobId);

                if (await verifyJobData(jobId)) {
                    createTasks(jobId);
                }
            }
            if (event.event_name === 'JobDeleted') {
                console.log(">>> Job Deleted: #", jobId);
            }
            if (event.event_name === 'JobUpdated') {
                console.log(">>> Job Updated: #", jobId);
            }
        }
    }, 2500);

    console.log("Listening for JobManager events...");
}

async function verifyJobData(jobId) {
    console.log('Verifying job data for jobID: #', jobId);

    try {
        const argumentCount = await jobManagerContract.getJobArgumentCount(jobId).call();
        const arguments = [];
        for (let i = 0; i < argumentCount; i++) {
            const arg = await jobManagerContract.getJobArgument(jobId, i).call();
            arguments.push(arg);
        }

        if (arguments.length > 0) {
            console.log(`>>> JobID #${jobId} has valid arguments.`);
            return true;
        } else {
            console.error(`!!! JobID #${jobId} has no valid arguments.`);
            return false;
        }
    } catch (error) {
        console.error(`!!! Error verifying job data for jobID #${jobId}:`, error);
        return false;
    }
}

async function createTasks(jobId) {
    try {
        const tx = await taskManagerContract.createNewTask(jobId, "0x01");
        const receipt = await tx.wait();

        const taskCreatedEvent = receipt.logs.find(log => log.eventName === 'TaskCreated');
        const taskId = taskCreatedEvent.args.taskId;    

        console.log(`>>> New task created for jobID #${jobId} with ID: #${taskId}.`);

        try {
            const result = await jobManagerContract.addTaskId(jobId, taskId).send();
            console.log(`>>> Added taskID #${taskId} to jobID #${jobId}.`);
        } catch (error) {
            console.error(`!!! Error adding taskID #${taskId} to jobID #${jobId}:`, error);
        }

        createTaskData(jobId, taskId);
    } catch (error) {
        console.error("!!! Error creating task:", error);
        throw error;
    }
}

async function createTaskData(jobId, taskId) {
    try {
        const encodedJobData = await jobManagerListener.getJobData(jobId);
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
        "uint256", "string", "string", "uint32", "uint256", "address", 
        "string", "uint256", "uint8", "bytes[]", "string", "uint32[]"
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

function getRandomKeeper() {
    const keeperIds = Object.keys(keeperConfigs);
    const randomIndex = Math.floor(Math.random() * keeperIds.length);
    const randomKeeperId = keeperIds[randomIndex];
    // return keeperConfigs[randomKeeperId];
    return keeperConfigs[1];
}  

async function sendTaskToKeeper(taskData) {
    const keeper = getRandomKeeper();
    const keeperUrl = `http://localhost:${keeper.port}/execute-task`;

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

app.listen(port, async () => {
    initializeWallets();
    await initializeContracts();
    await listenForJobManagerEvents();
});