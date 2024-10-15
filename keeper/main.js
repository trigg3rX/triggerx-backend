require('dotenv').config();
const { createRPCClient } = require('./keeperRPCclient');
const { executeTask } = require('./keeperTimeExecution');
const { getConfigById, getConfigByData } = require('./keeperConfig');

const keeperConfig = getConfigById(1);

console.log(' >>> Keeper Initialized with Config: ');
console.log(keeperConfig);

const keeperPort = keeperConfig.port;
const keeperName = keeperConfig.name;

const aggregatorUrl = process.env.AGGREGATOR_URL || 'http://localhost:3006/receive-result';

const keeperApp = createRPCClient(aggregatorUrl);

keeperApp.set('executeTask', executeTask);
keeperApp.set('keeperConfig', keeperConfig);

keeperApp.listen(keeperPort, () => {
    console.log(`Keeper service ${keeperName} listening at http://localhost:${keeperPort}`);
});