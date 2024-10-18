require('dotenv').config();
const express = require('express');
const { ethers } = require('ethers');
const cron = require('node-cron');
const axios = require('axios');
const taskManagerABI = require('../utils/abi/TaskManager.json');
const jobManagerABI = require('../utils/abi/JobManager.json');
const { keeperConfig: keeperConfigs } = require('../utils/keeperConfig');

// Addresses for smart contracts
const jobManagerAddress = '0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d';
const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';

let taskManagerContract;
let jobManagerContract;
let opSepoliaWallet;

const app = express();
const port = 3000;

app.use(express.json());

const activeJobs = {};

function initializeWallets() {
    const opSepoliaProvider = new ethers.JsonRpcProvider(process.env.OP_SEPOLIA_RPC_URL);
    const holeskyProvider = new ethers.JsonRpcProvider(process.env.ETHEREUM_RPC_URL);
    try {
        opSepoliaWallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, opSepoliaProvider);
    } catch (error) {
        console.error("!!! OP-Sepolia wallet initialization failed:", error.message);
        process.exit(1);
    }

    try {
        holeskyWallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, holeskyProvider);
    } catch (error) {
        console.error("!!! Holesky wallet initialization failed:", error.message);
        process.exit(1);
    }
    console.log(">>> Wallet initialized.");
}

async function initializeContracts() {
    try {
        // Fix 3: Add error checking for ABI imports
        if (!jobManagerABI || !taskManagerABI) {
            throw new Error("ABI files not loaded properly");
        }

        console.log("Job Manager ABI:", typeof jobManagerABI);
        console.log("Task Manager ABI:", typeof taskManagerABI);

        // Fix 4: Initialize contracts with proper error handling
        try {
            jobManagerContract = new ethers.Contract(
                jobManagerAddress, 
                jobManagerABI, 
                opSepoliaWallet
            );
        } catch (error) {
            console.error("Failed to initialize Job Manager contract:", error);
            throw error;
        }

        try {
            taskManagerContract = new ethers.Contract(
                taskManagerAddress, 
                taskManagerABI, 
                holeskyWallet
            );
        } catch (error) {
            console.error("Failed to initialize Task Manager contract:", error);
            throw error;
        }

        console.log(">>> Contracts initialized successfully.");
    } catch (error) {
        console.error("!!! Error initializing contracts:", error);
        throw error;
    }
}
async function getEventsOfLatestBlock(jobLimit) {
    try {
         // Set up a filter for JobCreated events

        const provider = new ethers.JsonRpcProvider(process.env.OP_SEPOLIA_RPC_URL);
        const latestBlock = await provider.getBlockNumber();
        
         // Adjust the block range as needed

        // console.log(`Found ${events.length} JobCreated events from the last 100 blocks:`);
        // events.forEach((event) => {
        //     console.log("EVENTTTT:: ", event);
        //     const jobId = event.args[0];
        //     console.log(">>> JobCreated Event: JobID =", jobId.toString());
        // });

        return {};
    } catch (error) {
        console.error("Error fetching JobCreated events:", error);
        throw error;
    }
}

async function listenForJobManagerEvents() {
    console.log(`JobManager listener running on port ${port}...`);
    const filter = jobManagerContract.filters.JobCreated;
    
    jobManagerContract.on("JobCreated", async (jobId, creator, amt, event) => {
        console.log("Job Created for #", jobId);
        
        if (await verifyJobData(jobId)) {
            createTasks(jobId);
        }
    });
}

async function verifyJobData(jobId) {
    console.log('Verifying job data for jobID: #', jobId);

    try {
        const argumentCount = await jobManagerContract.getJobArgumentCount(jobId);
        const arguments = [];
        for (let i = 0; i < argumentCount; i++) {
            const arg = await jobManagerContract.getJobArgument(jobId, i);
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

        const taskCreatedEvent = {
            eventName: receipt.logs[0].fragment.name,
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
            const result = await jobManagerContract.addTaskId(taskCreatedEvent.task.jobId, taskCreatedEvent.taskId);
            await result.wait();
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
        const jobData = await jobManagerContract.jobs(jobId);
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

    // Using ethers address format instead of TronWeb
    return {
        jobId: typeof jobId === 'bigint' ? Number(jobId) : jobId,
        jobType: jobType || '',
        status: status || '',
        timeframe: typeof timeframe === 'bigint' ? Number(timeframe) : timeframe,
        blockNumber: typeof blockNumber === 'bigint' ? blockNumber.toString() : blockNumber,
        contractAddress: contractAddress || ethers.ZeroAddress,
        targetFunction: targetFunction || '',
        timeInterval: typeof timeInterval === 'bigint' ? Number(timeInterval) : timeInterval,
        argType: ['None', 'Static', 'Dynamic'][Number(argType) || 0],
        arguments: Array.isArray(arguments) && arguments !== 'null' ? arguments : [],
        apiEndpoint: apiEndpoint === 'null' ? null : apiEndpoint,
        taskIds: [],
        stakeAmount: typeof stakeAmount === 'bigint' ? stakeAmount.toString() : stakeAmount
    };
}

async function sendTaskToKeeper(taskData) {
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