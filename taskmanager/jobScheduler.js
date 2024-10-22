require('dotenv').config();
const axios = require('axios');
const { parentPort, workerData } = require('worker_threads');
const cron = require('node-cron');
const { TronWeb } = require('tronweb');
const taskManagerABI = require('../utils/abi/TaskManager.json');
const { jobManagerABI } = require('../utils/abi/JobManager');
const { ethers } = require('ethers');
const { keeperConfig } = require('../utils/keeperConfig');

// Addresses for smart contracts
const jobManagerAddress = 'TECz49UXN9KhEF12WCrGHr4gV3CfUFiKsD';
const taskManagerAddress = '0x2FE0D258fb2eF69BAa3DD8c17469ea23B1952503';

let taskManagerContract;
let jobManagerContract;
let holeskyWallet;
let tronWeb;

async function initializeThings() {
    try {
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

        jobManagerContract = await tronWeb.contract(jobManagerABI, jobManagerAddress);

        taskManagerContract = new ethers.Contract(taskManagerAddress, taskManagerABI, holeskyWallet);
    } catch (error) {
        console.error("!!! Error initializing Things:", error);
        throw error;
    }
}

async function scheduleJobs(jobId) {
    try {
        parentPort.postMessage({ status: 'starting', jobId });

        const jobData = await createJobData(jobId);
        
        parentPort.postMessage({ status: 'jobData', jobId, data: jobData });
        parentPort.postMessage({ status: 'Arguments', jobId, data: jobData.arguments });

        const timeIntervalSeconds = Math.max(10, jobData.timeInterval);
        const cronExpression = `*/${timeIntervalSeconds} * * * * *`;
        let executionCount = Math.ceil(jobData.timeframe / jobData.timeInterval);
        
        parentPort.postMessage({ status: 'scheduling', jobId, cronExpression, executionCount });

        const task = cron.schedule(cronExpression, async () => {
            try {
                parentPort.postMessage({ status: 'executing', jobId });
                // console.log(executionCount);
                const taskCreatedEvent = await createTasks(jobId, jobData);

                parentPort.postMessage({ status: 'taskCreatedEvent', jobId, data: taskCreatedEvent });

                const taskExecuted = await sendTaskRequest(taskCreatedEvent, jobData);

                if (--executionCount === 0) {
                    parentPort.postMessage({ status: 'completed', jobId, totalExecutions: executionCount });
                    task.stop();
                    process.exit(0);
                }
            } catch (error) {
                parentPort.postMessage({ status: 'execution_error', jobId, error: error.message, execution: executionCount });
            }
        }, {
            scheduled: true,
            timezone: "UTC"
        });

        parentPort.on('message', (message) => {
            if (message === 'stop') {
                task.stop();
                parentPort.postMessage({ status: 'stopped', jobId });
                process.exit(0);
            }
        });

        parentPort.postMessage({ status: 'running', jobId, cronExpression });

    } catch (error) {
        parentPort.postMessage({ status: 'error', jobId, error: error.message });
        process.exit(1);
    }
}

async function createJobData(jobId) {
    try {
        const rawJobData = await jobManagerContract.jobs(jobId).call();
        const rawArguments = await jobManagerContract.getJobArgs(jobId).call();
        const formattedJob = formatJobData(rawJobData, rawArguments);
        return formattedJob;
    } catch (error) {
        console.error("!!! Error creating job data:", error);
        throw error;
    }
}

async function formatJobData(rawJobData, rawArguments) {
    const [
        jobId, jobType, status, timeframe, blockNumber, contractAddress,
        targetFunction, timeInterval, argType, apiEndpoint, userAddress, stakeAmount
    ] = rawJobData;

    const argTypes = targetFunction.match(/\((.*?)\)/)?.[1]?.split(',') || [];

    const formattedArguments = rawArguments.map((arg, index) => {
        const type = argTypes[index]?.trim();
        let value;

        switch (type) {
            case 'uint256':
            case 'uint128':
            case 'uint64':
            case 'uint32':
            case 'uint16':
            case 'uint8':
                value = BigInt(`0x${arg.slice(2)}`).toString();
                break;
            case 'int256':
            case 'int128':
            case 'int64':
            case 'int32':
            case 'int16':
            case 'int8':
                value = BigInt(`0x${arg.slice(2)}`).toString();
                if (value[0] !== '-' && BigInt(value) > BigInt(2) ** BigInt(type.slice(3) - 1)) {
                    value = '-' + (BigInt(2) ** BigInt(type.slice(3)) - BigInt(value)).toString();
                }
                break;
            case 'address':
                value = tronWeb.address.fromHex(`0x${arg.slice(2)}`);
                break;
            case 'bool':
                value = arg === '0x01';
                break;
            case 'string':
                value = tronWeb.toUtf8(arg);
                break;
            case 'bytes':
            case 'bytes32':
            case 'bytes16':
            case 'bytes8':
            case 'bytes4':
            case 'bytes1':
                value = `0x${arg.slice(2)}`;
                break;
            default:
                console.warn(`Unsupported or unknown type: ${type}`);
                value = arg;
        }

        return value;
    });

    return {
        jobId: typeof jobId === 'bigint' ? Number(jobId) : jobId,
        jobType: jobType || '',
        timeframe: typeof timeframe === 'bigint' ? Number(timeframe) : timeframe,
        contractAddress: contractAddress ? tronWeb.address.fromHex(contractAddress) : '',
        targetFunction: targetFunction.split('(')[0] || '',
        timeInterval: typeof timeInterval === 'bigint' ? Number(timeInterval) : timeInterval,
        argType: ['None', 'Static', 'Dynamic'][Number(argType) || 0],
        arguments: formattedArguments,
        apiEndpoint: apiEndpoint === 'null' ? null : apiEndpoint
    };
}

async function createTasks(jobId, jobData) {
    try {
        // console.log(`Creating task for jobID #${jobId}`);
        const quorumNumbers = 1;
        const quorumNumbersBytes32 = '0x' + quorumNumbers.toString(16).padStart(64, '0');
        const tx = await taskManagerContract.createNewTask(jobId, quorumNumbersBytes32);
        // console.log('Transaction sent, waiting for receipt...');
        const receipt = await tx.wait();

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
        
        // console.log(`>>> New task created for jobID #${jobId} with ID: #${taskCreatedEvent.taskId}.`);

        try {
            const result = await jobManagerContract.addTaskId(taskCreatedEvent.task.jobId, taskCreatedEvent.taskId).send();
            // console.log(`>>> Added taskID #${taskCreatedEvent.taskId} to jobID #${taskCreatedEvent.task.jobId}.`);
        } catch (error) {
            console.error(`!!! Error adding taskID #${taskCreatedEvent.taskId} to jobID #${taskCreatedEvent.task.jobId}:`, error);
        }

        return taskCreatedEvent;
    } catch (error) {
        console.error("!!! Error creating task:", error);
        console.log('Full error object:', error);
        return null;
    }
}

// function getRandomKeeper() {
    // const keeperIds = Object.keys(keeperConfigs);
    // const randomIndex = Math.floor(Math.random() * keeperIds.length);
    // const randomKeeperId = keeperIds[randomIndex];
    // return keeperConfigs[randomKeeperId];
// }  

async function sendTaskRequest(taskCreatedEvent, jobData) {
    const keeper = keeperConfig[1];

    console.log(`>>> Keeper: ${keeper.name} on port ${keeper.port}`);

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

    const taskData = {
        taskId: taskCreatedEvent.taskId,
        jobId: taskCreatedEvent.task.jobId,
        jobType: jobData.jobType,
        contractAddress: jobData.contractAddress,
        targetFunction: jobData.targetFunction,
        argType: jobData.argType,
        argumentInfo: jobData.arguments,
        apiEndpoint: jobData.apiEndpoint,
        blockNumber: taskCreatedEvent.task.taskCreatedBlock,
        quorumNumbers: taskCreatedEvent.task.quorumNumbers
    };

    const convertedTaskData = convertNestedBigInt(taskData);
    // console.log(`Sending task to keeper #${keeper.id} on port ${keeper.port}`);

    try {
        const response = await axios.post(keeperUrl, convertedTaskData);
        console.log(`>>> Task sent to keeper. Response: ${response.status} ${response.statusText}`);
        return true;
    } catch (error) {
        console.error("!!! Error sending task to keeper:", error.message);
        return false;
    }
}

async function main() {
    await initializeThings();
    scheduleJobs(workerData.jobId);
}

main();