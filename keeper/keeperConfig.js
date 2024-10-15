require('dotenv').config();
const { ethers } = require('ethers');
const { TronWeb } = require('tronweb');
const fs = require('fs');
const path = require('path');
const { keeperConfig } = require('../utils/keeperConfig');

const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
const wallet = new ethers.Wallet(process.env.ETHEREUM_PRIVATE_KEY, provider);

const tronWeb = new TronWeb({
    fullHost: process.env.TRON_FULL_HOST,
    privateKey: process.env.TRON_PRIVATE_KEY
});

const eigenAddresses = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/contracts/eigenAddresses.json'), 'utf8'));
const avsAddresses = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/contracts/avsAddresses.json'), 'utf8'));

const delegationManagerAddress = eigenAddresses.delegationManager;
const avsDirectoryAddress = eigenAddresses.avsDirectory;

const serviceManagerAddress = avsAddresses.serviceManager;
const ecdsaStakeRegistryAddress = avsAddresses.ecdsaStakeRegistry;

const delegationManagerABI = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/IDelegationManager.json'), 'utf8'));
const avsDirectoryABI = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/IAVSDirectory.json'), 'utf8'));

const serviceManagerABI = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/ServiceManager.json'), 'utf8'));
const ecdsaStakeRegistryABI = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../utils/abi/ECDSAStakeRegistry.json'), 'utf8'));

const delegationManager = new ethers.Contract(delegationManagerAddress, delegationManagerABI, wallet);
const serviceManager = new ethers.Contract(serviceManagerAddress, serviceManagerABI, wallet);
const ecdsaStakeRegistry = new ethers.Contract(ecdsaStakeRegistryAddress, ecdsaStakeRegistryABI, wallet);
const avsDirectory = new ethers.Contract(avsDirectoryAddress, avsDirectoryABI, wallet);

class Config{
    constructor(privateKey, port, name) {
        this.privateKey = privateKey;
        this.port = port;
        this.name = name;
    }
}

function isValidPrivateKey(privateKey) {
    try {
        new ethers.Wallet(privateKey);
        return true;
    } catch (error) {
        return false;
    }
}

function isValidPort(port) {
    return Number.isInteger(port) && port > 0 && port <= 65535;
}

function isValidName(name) {
    const regex = /^[a-zA-Z0-9]{4,12}$/;
    return regex.test(name);
}

function getEthWalletAddres(privateKey) {
    return new ethers.Wallet(privateKey).address;
}

function getTrxWalletAddress(privateKey) {
    return tronWeb.address.fromPrivateKey(privateKey);
}

function getConfigById(id) {
    const configData = keeperConfig[id];
    
    if (!configData) {
        throw new Error(`Configuration with ID ${id} not found`);
    }
    
    const config = new Config(configData.privateKey, configData.port, configData.name);
    
    return config;
}

function getConfigByData(data) {
    const { privateKey, port, name } = data;

    if (!isValidPrivateKey(privateKey)) {
        throw new Error('Invalid private key');
    }

    if (!isValidPort(port)) {
        throw new Error('Invalid port number');
    }

    if (!isValidName(name)) {
        throw new Error('Invalid name. It should be an alphanumeric string with length 4-12 characters');
    }

    return new Config(privateKey, port, name);
}

const registerKeeperOnEigenLayer = async (privateKey, strategyAddress) => {
    
    try {
        const tx1 = await delegationManager.registerAsOperator({
            __deprecated_earningsReceiver: getEthWalletAddres(privateKey),
            delegationApprover: strategyAddress,
            stakerOptOutWindowBlocks: 0
        }, "");

        await tx1.wait();
        console.log("Registered as operator on Eigen Layer");
    } catch (error) {
        console.log("Error while registering as operator on Eigen Layer", error);
    }

    const salt = ethers.hexlify(ethers.randomBytes(32));
    const expiry = Math.floor(Date.now() / 1000) + 3600;

    let operatorSignatureWithSaltAndExpiry = {
        signature: "",
        salt: salt,
        expiry: expiry
    };

    const operatorDigestHash = await avsDirectory.calculateOperatorAVSRegistrationDigestHash(
        wallet.address, 
        await helloWorldServiceManager.getAddress(), 
        salt, 
        expiry
    );
    console.log(operatorDigestHash);

    console.log("Signing digest hash with operator's private key");

    const operatorSigningKey = new ethers.SigningKey(process.env.PRIVATE_KEY);
    const operatorSignedDigestHash = operatorSigningKey.sign(operatorDigestHash);

    // Encode the signature in the required format
    operatorSignatureWithSaltAndExpiry.signature = ethers.Signature.from(operatorSignedDigestHash).serialized;

    console.log("Registering Operator to AVS Registry contract");

    const tx2 = await ecdsaRegistryContract.registerOperatorWithSignature(
        operatorSignatureWithSaltAndExpiry,
        wallet.address
    );
    await tx2.wait();
    console.log("Operator registered on AVS successfully");
}


// console.log(getConfigById(1));

module.exports = { getConfigById, getConfigByData };