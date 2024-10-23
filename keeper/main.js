require('dotenv').config();
const { createRPCClient } = require('./keeperRPCclient');
const { executeTask } = require('./keeperTimeExecution');
const { getConfigById, getConfigByData } = require('./keeperConfig');

const keeperId = parseInt(process.env.KEEPER_ID) || 1;
const keeperConfig = getConfigById(keeperId);

console.log(' >>> Keeper Initialized with Config: ');
console.log(keeperConfig);

process.env.KEEPER_PRIVATE_KEY = keeperConfig.privateKey;

const aggregatorUrl = `http://${process.env.HOST_IP}:3006/receive-result` || 'http://localhost:3006/receive-result';

const keeperApp = createRPCClient(aggregatorUrl);

keeperApp.set('executeTask', executeTask);
keeperApp.set('keeperConfig', keeperConfig);

keeperApp.listen(keeperConfig.port, '0.0.0.0',  () => {
    console.log(`Keeper service ${keeperConfig.name} listening at http://localhost:${keeperConfig.port}`);
});