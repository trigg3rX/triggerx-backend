require('dotenv').config();
const express = require('express');
const { Worker } = require('worker_threads');
const path = require('path');
const { TronWeb } = require('tronweb');
const { jobManagerABI } = require('../utils/abi/JobManager');

// Addresses for smart contracts
const jobManagerAddress = 'TECz49UXN9KhEF12WCrGHr4gV3CfUFiKsD';

let jobManagerContract;
let tronWeb;

const app = express();
const port = 3000;

const activeWorkers = new Map();

app.use(express.json());
 
function initializeWallets() {
    tronWeb = new TronWeb({
        fullHost: process.env.TRON_FULL_HOST,
        privateKey: process.env.TRON_PRIVATE_KEY
    });

    console.log(">>> Wallets initialized.");
}

async function initializeContracts() {
    try {
        jobManagerContract = await tronWeb.contract(jobManagerABI, jobManagerAddress);
    } catch (error) {
        console.error("!!! Error initializing JobManager:", error);
        throw error;
    }

    console.log(">>> Contracts initialized.");
}

async function getEventsOfLatestBlock(jobLimit) {
    const events = await tronWeb.event.getEventsByContractAddress(
        jobManagerAddress,
        {
            orderBy: 'block_timestamp,desc',
            limit: jobLimit,
        }
      );
    return events.data;
}

async function listenForJobManagerEvents() {
    console.log(`JobManager listener running on port ${port}...`);

    let jobLimit = 1;
    let lastJobId = 3;
    
    setInterval(async () => {
        const jobManagerEvents = await getEventsOfLatestBlock(jobLimit);
        
        for (const event of jobManagerEvents) {
            // console.log(event);
            const jobId = event.result.jobId;
            if (event.event_name === 'JobCreated') {
                if (jobId > lastJobId) {
                    lastJobId = jobId;
                
                    console.log(">>> New job created: #", jobId);

                    if (!activeWorkers.has(jobId)) {
                        const worker = new Worker(path.join(__dirname, 'jobScheduler.js'), {
                            workerData: { jobId }
                        });

                        activeWorkers.set(jobId, worker);

                        worker.on('message', (message) => {
                            console.log(`Worker message for Job #${jobId}:`, message);
                            if (message.status === 'completed' || message.status === 'error') {
                                worker.terminate();
                                activeWorkers.delete(jobId);
                            }
                        });

                        worker.on('error', (error) => {
                            console.error(`Worker error for Job #${jobId}:`, error);
                            worker.terminate();
                            activeWorkers.delete(jobId);
                        });

                        worker.on('exit', (code) => {
                            if (code !== 0) {
                                console.error(`Worker for Job #${jobId} stopped with exit code ${code}`);
                            }
                            activeWorkers.delete(jobId);
                        });
                    }
                }
            } else if (event.event_name === 'JobDeleted') {
                console.log(">>> Job Deleted: #", jobId);
                const worker = activeWorkers.get(jobId);
                if (worker) {
                    worker.postMessage('stop');
                    activeWorkers.delete(jobId);
                }
            } else if (event.event_name === 'JobUpdated') {
                console.log(">>> Job Updated: #", jobId);
                // You might want to restart the worker with new job data
                // For now, we'll just log it
            }
        }
        jobLimit = 5;
    }, 500);
}

app.listen(port, async () => {
    initializeWallets();
    await initializeContracts();
    await listenForJobManagerEvents();
});