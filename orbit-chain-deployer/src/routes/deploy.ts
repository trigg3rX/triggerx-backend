import { Router, Request, Response } from 'express';
import {
  DeployChainRequest,
  DeployContractsRequest,
  ChainDeploymentResponse,
  ContractDeploymentResponse,
  DeploymentStatus
} from '../types/deployment';
import { OrbitService, ContractsService, StatusService } from '../services';
import { config } from '../utils/config';
import logger from '../utils/logger';

const router = Router();

// Initialize services using config
const orbitService = new OrbitService({
  parentChainRpc: config.arbitrumRpcUrl || 'https://sepolia-rollup.arbitrum.io/rpc',
  parentChainId: parseInt(process.env.PARENT_CHAIN_ID || '421614'),
  deployerPrivateKey: config.deployerPrivateKey,
  batchPosterPrivateKey: config.batchPosterPrivateKey,
  validatorPrivateKey: config.validatorPrivateKey,
  parentChainBeaconRpcUrl: process.env.ETHEREUM_BEACON_RPC_URL,
  nodeConfigDir: config.node.nodeConfigDir,
  defaultRpcPort: config.node.defaultRpcPort,
  defaultExplorerPort: config.node.defaultExplorerPort,
  contractArtifactsBaseUrl: process.env.CONTRACT_ARTIFACTS_BASE_URL,
  dockerImage: config.node.dockerImage,
  containerPrefix: config.node.containerPrefix,
  nitroBinaryPath: config.node.nitroBinaryPath,
  memoryLimit: config.node.memoryLimit,
  cpuLimit: config.node.cpuLimit,
  startupTimeout: config.node.startupTimeout,
  healthCheckInterval: config.node.healthCheckInterval
});

const contractsService = new ContractsService({
  deployerPrivateKey: config.deployerPrivateKey || '',
  contractArtifactsBaseUrl: process.env.CONTRACT_ARTIFACTS_BASE_URL || '',
  jobRegistryAddress: process.env.JOB_REGISTRY_ADDRESS || '',
  gasRegistryAddress: process.env.GAS_REGISTRY_ADDRESS || '',
  taskExecutionSpokeAddress: process.env.TASK_EXECUTION_SPOKE_ADDRESS || ''
});

const statusService = new StatusService({
  goBackendUrl: config.goBackendUrl,
  apiKey: config.goBackendApiKey
});

// Deploy Chain endpoint - called by Go backend to deploy a new Orbit chain
router.post('/deploy-chain', async (req: Request, res: Response) => {
  const { v4: uuidv4 } = await import('uuid');
  const requestId = uuidv4();
  
  try {
    logger.info('Chain deployment request received', { requestId });
    
    const deployRequest: DeployChainRequest = req.body;
    
    // Validate required fields
    if (!deployRequest.deployment_id || !deployRequest.chain_name || !deployRequest.owner_address) {
      const missing = [];
      if (!deployRequest.deployment_id) missing.push('deployment_id');
      if (!deployRequest.chain_name) missing.push('chain_name');
      if (!deployRequest.owner_address) missing.push('owner_address');
      
      return res.status(400).json({
        success: false,
        deployment_id: deployRequest.deployment_id || '',
        status: DeploymentStatus.FAILED,
        message: `Missing required fields: ${missing.join(', ')}`
      });
    }
    
    // Start deployment tracking
    const progress = statusService.startDeploymentTracking(
      deployRequest.deployment_id,
      deployRequest.chain_name
    );

    // Update status to deploying orbit
    await statusService.updateDeploymentStatus(
      deployRequest.deployment_id,
      DeploymentStatus.DEPLOYING_ORBIT,
      { deploymentLogs: 'Starting Orbit chain deployment' }
    );

    // Deploy the Orbit chain
    const deploymentResult = await orbitService.deployChain(deployRequest);

    if (deploymentResult.success) {
      // Update status to orbit deployed
      await statusService.updateDeploymentStatus(
        deployRequest.deployment_id,
        DeploymentStatus.ORBIT_DEPLOYED,
        {
          chainAddress: deploymentResult.chainAddress,
          deploymentLogs: 'Orbit chain deployed successfully'
        }
      );

      // Update local progress
      statusService.updateDeploymentProgress(deployRequest.deployment_id, {
        status: DeploymentStatus.ORBIT_DEPLOYED,
        progress: 50,
        currentStep: 'Orbit chain deployed, ready for contracts',
        chainAddress: deploymentResult.chainAddress,
        log: 'Orbit chain deployed successfully'
      });

      const response: ChainDeploymentResponse = {
        success: true,
        deployment_id: deployRequest.deployment_id,
        status: DeploymentStatus.ORBIT_DEPLOYED,
        message: 'Chain deployment completed successfully',
        chain_address: deploymentResult.chainAddress,
        deployment_tx_hash: deploymentResult.deploymentTxHash
      };

      res.status(200).json(response);
    } else {
      // Update status to failed
      await statusService.updateDeploymentStatus(
        deployRequest.deployment_id,
        DeploymentStatus.FAILED,
        { errorMessage: deploymentResult.error }
      );

      statusService.completeDeploymentTracking(
        deployRequest.deployment_id,
        false,
        deploymentResult.error
      );

      const response: ChainDeploymentResponse = {
        success: false,
        deployment_id: deployRequest.deployment_id,
        status: DeploymentStatus.FAILED,
        message: 'Chain deployment failed',
        error: deploymentResult.error
      };

      res.status(500).json(response);
    }
    
    return;
    
  } catch (error) {
    logger.error('Chain deployment failed', { 
      requestId, 
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    
    res.status(500).json({
      success: false,
      deployment_id: req.body?.deployment_id || '',
      status: DeploymentStatus.FAILED,
      message: 'Internal server error during chain deployment',
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    return;
  }
});

// Deploy Contracts endpoint - called by Go backend to deploy TriggerX contracts
router.post('/deploy-contracts', async (req: Request, res: Response) => {
  const { v4: uuidv4 } = await import('uuid');
  const requestId = uuidv4();
  
  try {
    logger.info('Contract deployment request received', { requestId });
    
    const deployRequest: DeployContractsRequest = req.body;
    
    // Validate required fields
    if (!deployRequest.deployment_id || !deployRequest.chain_address || !deployRequest.contracts?.length) {
      const missing = [];
      if (!deployRequest.deployment_id) missing.push('deployment_id');
      if (!deployRequest.chain_address) missing.push('chain_address');
      if (!deployRequest.contracts?.length) missing.push('contracts');
      
      return res.status(400).json({
        success: false,
        deployment_id: deployRequest.deployment_id || '',
        status: DeploymentStatus.FAILED,
        message: `Missing required fields: ${missing.join(', ')}`
      });
    }
    
    // Update status to deploying contracts
    await statusService.updateDeploymentStatus(
      deployRequest.deployment_id,
      DeploymentStatus.DEPLOYING_CONTRACTS,
      { deploymentLogs: 'Starting TriggerX contracts deployment' }
    );

    // Update local progress
    statusService.updateDeploymentProgress(deployRequest.deployment_id, {
      status: DeploymentStatus.DEPLOYING_CONTRACTS,
      progress: 60,
      currentStep: 'Deploying TriggerX contracts',
      log: 'Starting contracts deployment'
    });

    // Deploy the contracts
    const deploymentResult = await contractsService.deployContracts(deployRequest);

    if (deploymentResult.success) {
      // Update status to configuring contracts
      await statusService.updateDeploymentStatus(
        deployRequest.deployment_id,
        DeploymentStatus.CONFIGURING_CONTRACTS,
        {
          contracts: deploymentResult.contracts,
          deploymentLogs: 'Contracts deployed, configuring relationships'
        }
      );

      // Update local progress
      statusService.updateDeploymentProgress(deployRequest.deployment_id, {
        status: DeploymentStatus.CONFIGURING_CONTRACTS,
        progress: 90,
        currentStep: 'Configuring contract relationships',
        contracts: deploymentResult.contracts,
        log: 'Contracts deployed successfully'
      });

      // Final status update to completed
      await statusService.updateDeploymentStatus(
        deployRequest.deployment_id,
        DeploymentStatus.COMPLETED,
        {
          contracts: deploymentResult.contracts,
          deploymentLogs: 'Deployment completed successfully'
        }
      );

      // Complete deployment tracking
      statusService.completeDeploymentTracking(deployRequest.deployment_id, true);

      const response: ContractDeploymentResponse = {
        success: true,
        deployment_id: deployRequest.deployment_id,
        status: DeploymentStatus.COMPLETED,
        message: 'Contracts deployment completed successfully',
        contracts: deploymentResult.contracts,
        deployment_tx_hashes: deploymentResult.deploymentTxHashes
      };

      res.status(200).json(response);
    } else {
      // Update status to failed
      await statusService.updateDeploymentStatus(
        deployRequest.deployment_id,
        DeploymentStatus.FAILED,
        { errorMessage: deploymentResult.error }
      );

      statusService.completeDeploymentTracking(
        deployRequest.deployment_id,
        false,
        deploymentResult.error
      );

      const response: ContractDeploymentResponse = {
        success: false,
        deployment_id: deployRequest.deployment_id,
        status: DeploymentStatus.FAILED,
        message: 'Contracts deployment failed',
        error: deploymentResult.error
      };

      res.status(500).json(response);
    }
    
    return;
    
  } catch (error) {
    logger.error('Contract deployment failed', { 
      requestId, 
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    
    res.status(500).json({
      success: false,
      deployment_id: req.body?.deployment_id || '',
      status: DeploymentStatus.FAILED,
      message: 'Internal server error during contract deployment',
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    return;
  }
});

// Get deployment status endpoint
router.get('/deployment-status/:deploymentId', async (req: Request, res: Response) => {
  const { v4: uuidv4 } = await import('uuid');
  const requestId = uuidv4();
  
  try {
    const { deploymentId } = req.params;
    
    logger.info('Deployment status request received', { requestId, deploymentId });
    
    // Get local progress
    const localProgress = statusService.getDeploymentProgress(deploymentId);
    
    // Get status from Go backend
    const backendStatus = await statusService.getDeploymentStatus(deploymentId);
    
    const response = {
      deployment_id: deploymentId,
      local_progress: localProgress,
      backend_status: backendStatus,
      timestamp: new Date().toISOString()
    };
    
    res.status(200).json(response);
    return;
    
  } catch (error) {
    logger.error('Failed to get deployment status', { 
      requestId, 
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    
    res.status(500).json({
      success: false,
      message: 'Failed to get deployment status',
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    return;
  }
});

// Get deployment statistics endpoint
router.get('/deployment-stats', async (req: Request, res: Response) => {
  const { v4: uuidv4 } = await import('uuid');
  const requestId = uuidv4();
  
  try {
    logger.info('Deployment statistics request received', { requestId });
    
    const stats = statusService.getDeploymentStatistics();
    
    res.status(200).json({
      success: true,
      statistics: stats,
      timestamp: new Date().toISOString()
    });
    return;
    
  } catch (error) {
    logger.error('Failed to get deployment statistics', { 
      requestId, 
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    
    res.status(500).json({
      success: false,
      message: 'Failed to get deployment statistics',
      error: error instanceof Error ? error.message : 'Unknown error'
    });
    return;
  }
});

export default router;
