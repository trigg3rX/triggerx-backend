require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const { TronWeb } = require('tronweb');
const taskManagerABI = require('../utils/abi/TaskManager.json');
const { jobManagerABI } = require('../utils/abi/JobManager');
const { keeperConfig: keeperConfigs } = require('../utils/keeperConfig');


// Addresses for smart contracts
const jobManagerAddress = 'TEsKaf2n8aF6pta7wyG5gwukzR4NoHre59';
const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';

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

async function getEventsOfLatestBlock(jobLimit) {
    const events = await tronWeb.event.getEventsByContractAddress(
        jobManagerAddress,
        {
            // onlyConfirmed: true,
            orderBy: 'block_timestamp,desc',
            limit: jobLimit,
            // minBlockTimestamp: Date.now() - 60000,
            // maxBlockTimestamp: Date.now()
        }
      );
    return events.data;
}

async function listenForJobManagerEvents() {
    console.log(`JobManager listener running on port ${port}...`);

    let jobLimit = 5;
    let lastJobId = 133;
    
    setInterval(async () => {
        const jobManagerEvents = await getEventsOfLatestBlock(jobLimit);
        
        for (const event of jobManagerEvents) {
            // console.log(event);
            const jobId = event.result.jobId;
            if (event.event_name === 'JobCreated') {
                if (jobId > lastJobId) {
                    lastJobId = jobId;
                
                    console.log(">>> New job created: #", jobId);

                    // if (await verifyJobData(jobId)) {
                        // createTasks(jobId);
                    // }
                }
            }
            // if (event.event_name === 'JobDeleted') {
            //     console.log(">>> Job Deleted: #", jobId);
            // }
            // if (event.event_name === 'JobUpdated') {
            //     console.log(">>> Job Updated: #", jobId);
            // }
        }
        jobLimit = 5;
    }, 3000);
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
        console.log(`Creating task for jobID #${jobId}`);
        const tx = await taskManagerContract.createNewTask(jobId, "0x0000000000000000000000000000000000000000000000000000000000000001");
        console.log('Transaction sent, waiting for receipt...');
        const receipt = await tx.wait();
        // console.log('Transaction receipt received:', receipt);

        // console.log('Events in the receipt:');
        // receipt.logs.forEach((log, index) => {
        //     console.log(`Event ${index}:`, log);
        // });

        const taskCreatedEvent = {
            event_name: receipt.logs[0].fragment.name,
            taskId: receipt.logs[0].args[0].toString(),
            task: {
                jobId: receipt.logs[0].args[1][0].toString(),
                taskCreatedBlock: Number(receipt.logs[0].args[1][1]),
                quorumNumbers: receipt.logs[0].args[1][2],
            }
        };

        console.log('Task created event:', taskCreatedEvent);
        
        console.log(`>>> New task created for jobID #${jobId} with ID: #${taskCreatedEvent.taskId}.`);

        try {
            const result = await jobManagerContract.addTaskId(taskCreatedEvent.task.jobId, taskCreatedEvent.taskId).send();
            console.log(`>>> Added taskID #${taskCreatedEvent.taskId} to jobID #${taskCreatedEvent.task.jobId}.`);
        } catch (error) {
            console.error(`!!! Error adding taskID #${taskCreatedEvent.taskId} to jobID #${taskCreatedEvent.task.jobId}:`, error);
        }

        await createTaskData(taskCreatedEvent.taskId, taskCreatedEvent.task);
    } catch (error) {
        console.error("!!! Error creating task:", error);
        console.log('Full error object:', error);
    }
}

async function getJobData(jobId) {
    try {
        const jobData = await jobManagerContract.jobs(jobId).call();
        return jobData;
    } catch (error) {
        console.error(`Error fetching job data for jobID #${jobId}:`, error);
        throw error;
    }
}

async function createTaskData(taskId, taskCreated) {
    try {
        const rawJobData = await getJobData(taskCreated.jobId);
        console.log("Raw job data:", rawJobData);

        const formattedJob = formatJobData(rawJobData);
        console.log("Decoded job data:", formattedJob);

        const taskData = {
            taskId: taskId.toString(),
            jobId: taskCreated.jobId.toString(),
            jobType: formattedJob.jobType,
            blockNumber: taskCreated.taskCreatedBlock,
            quorumNumbers: taskCreated.quorumNumbers,
            contractAddress: formattedJob.contractAddress,
            targetFunction: formattedJob.targetFunction,
            argType: formattedJob.argType,
            argumentInfo: {
                type: formattedJob.argType,
                arguments: formattedJob.arguments
            },
            apiEndpoint: formattedJob.apiEndpoint,
            timeInterval: formattedJob.timeInterval
        };

        console.log("Structured task data:", taskData);

        const timeIntervalSeconds = Math.max(1, formattedJob.timeInterval);
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

function formatJobData(rawJobData) {
    if (!Array.isArray(rawJobData) || rawJobData.length < 12) {
        console.error('!!! Invalid job data format:', rawJobData);
        return {};
    }

    const [
        jobId, jobType, status, timeframe, blockNumber, contractAddress,
        targetFunction, timeInterval, argType, arguments, apiEndpoint, stakeAmount
    ] = rawJobData;

    // Find jobCreator in the additional fields
    const jobCreator = rawJobData.find(item => item === 'jobCreator');
    const jobCreatorValue = jobCreator ? rawJobData[rawJobData.indexOf(jobCreator) + 1] : null;

    return {
        jobId: typeof jobId === 'bigint' ? Number(jobId) : jobId,
        jobType: jobType || '',
        status: status || '',
        timeframe: typeof timeframe === 'bigint' ? Number(timeframe) : timeframe,
        blockNumber: typeof blockNumber === 'bigint' ? blockNumber.toString() : blockNumber,
        contractAddress: contractAddress ? tronWeb.address.fromHex(contractAddress) : '',
        targetFunction: targetFunction || '',
        timeInterval: typeof timeInterval === 'bigint' ? Number(timeInterval) : timeInterval,
        argType: ['None', 'Static', 'Dynamic'][Number(argType) || 0],
        arguments: Array.isArray(arguments) && arguments !== 'null' ? arguments : [],
        apiEndpoint: apiEndpoint === 'null' ? null : apiEndpoint,
        taskIds: [], // Not present in the raw data, initialize as empty array
        jobCreator: jobCreatorValue ? tronWeb.address.fromHex(jobCreatorValue) : '',
        stakeAmount: typeof stakeAmount === 'bigint' ? stakeAmount.toString() : stakeAmount
    };
}

// function getRandomKeeper() {
//     const keeperIds = Object.keys(keeperConfigs);
//     const randomIndex = Math.floor(Math.random() * keeperIds.length);
//     const randomKeeperId = keeperIds[randomIndex];
//     // return keeperConfigs[randomKeeperId];
//     return keeperConfigs[1];
// }  

async function sendTaskToKeeper(taskData) {
    // console.log('Keeper Configurations:', keeperConfigs);
    const keeper = keeperConfigs[1];

    
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