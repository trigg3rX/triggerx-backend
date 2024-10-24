const jobManagerABI = [
    {
      "inputs": [
        { "indexed": true, "name": "jobId", "type": "uint32" },
        { "indexed": true, "name": "creator", "type": "address" },
        { "name": "stakeAmount", "type": "uint256" }
      ],
      "name": "JobCreated",
      "type": "Event"
    },
    {
      "inputs": [
        { "indexed": true, "name": "jobId", "type": "uint32" },
        { "indexed": true, "name": "creator", "type": "address" },
        { "name": "stakeRefunded", "type": "uint256" }
      ],
      "name": "JobDeleted",
      "type": "Event"
    },
    {
      "inputs": [{ "indexed": true, "name": "jobId", "type": "uint32" }],
      "name": "JobUpdated",
      "type": "Event"
    },
    {
      "inputs": [
        { "name": "jobId", "type": "uint32" },
        { "name": "taskId", "type": "uint32" }
      ],
      "name": "addTaskId",
      "stateMutability": "Nonpayable",
      "type": "Function"
    },
    {
      "outputs": [{ "type": "uint32" }],
      "inputs": [
        { "name": "jobType", "type": "string" },
        { "name": "timeframe", "type": "uint32" },
        { "name": "contractAddress", "type": "address" },
        { "name": "targetFunction", "type": "string" },
        { "name": "timeInterval", "type": "uint256" },
        { "name": "argType", "type": "uint8" },
        { "name": "arguments", "type": "bytes[]" },
        { "name": "apiEndpoint", "type": "string" }
      ],
      "name": "createJob",
      "stateMutability": "Payable",
      "type": "Function"
    },
    {
      "inputs": [
        { "name": "jobId", "type": "uint32" },
        { "name": "stakeConsumed", "type": "uint256" }
      ],
      "name": "deleteJob",
      "stateMutability": "Nonpayable",
      "type": "Function"
    },
    {
      "outputs": [{ "type": "bytes[]" }],
      "inputs": [{ "name": "jobId", "type": "uint32" }],
      "name": "getJobArgs",
      "stateMutability": "view",
      "type": "function"
    },
    {
      "outputs": [{ "type": "uint32[]" }],
      "inputs": [{ "name": "jobId", "type": "uint32" }],
      "name": "getTaskIds",
      "stateMutability": "view",
      "type": "function"
    },
    {
      "outputs": [
        { "name": "jobId", "type": "uint32" },
        { "name": "jobType", "type": "string" },
        { "name": "status", "type": "string" },
        { "name": "timeframe", "type": "uint32" },
        { "name": "blockNumber", "type": "uint256" },
        { "name": "contractAddress", "type": "address" },
        { "name": "targetFunction", "type": "string" },
        { "name": "timeInterval", "type": "uint256" },
        { "name": "argType", "type": "uint8" },
        { "name": "apiEndpoint", "type": "string" },
        { "name": "jobCreator", "type": "address" },
        { "name": "stakeAmount", "type": "uint256" }
      ],
      "inputs": [{ "type": "uint32" }],
      "name": "jobs",
      "stateMutability": "view",
      "type": "function"
    },
    {
      "inputs": [
        { "name": "jobId", "type": "uint32" },
        { "name": "status", "type": "string" }
      ],
      "name": "setJobStatus",
      "stateMutability": "Nonpayable",
      "type": "Function"
    },
    {
      "inputs": [
        { "name": "jobId", "type": "uint32" },
        { "name": "jobType", "type": "string" },
        { "name": "timeframe", "type": "uint32" },
        { "name": "contractAddress", "type": "address" },
        { "name": "targetFunction", "type": "string" },
        { "name": "timeInterval", "type": "uint256" },
        { "name": "argType", "type": "uint8" },
        { "name": "arguments", "type": "bytes[]" },
        { "name": "apiEndpoint", "type": "string" },
        { "name": "stakeAmount", "type": "uint256" }
      ],
      "name": "updateJob",
      "stateMutability": "Nonpayable",
      "type": "Function"
    },
    {
      "outputs": [{ "type": "uint32" }],
      "inputs": [{ "type": "address" }, { "type": "uint256" }],
      "name": "userJobs",
      "stateMutability": "view",
      "type": "function"
    },
    {
      "outputs": [{ "type": "uint32" }],
      "inputs": [{ "type": "address" }],
      "name": "userJobsCount",
      "stateMutability": "view",
      "type": "function"
    },
    {
      "outputs": [{ "type": "uint256" }],
      "inputs": [{ "type": "address" }],
      "name": "userTotalStake",
      "stateMutability": "view",
      "type": "function"
    }
  ];

module.exports = { jobManagerABI };