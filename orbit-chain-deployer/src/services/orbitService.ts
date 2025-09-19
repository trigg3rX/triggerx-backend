// Dynamic imports for Orbit SDK to handle ES module compatibility
import { createPublicClient, createWalletClient, http, type Address } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { arbitrumSepolia } from 'viem/chains';
import logger from '../utils/logger';
import { DeployChainRequest, DeploymentStatus, ChainDeploymentResponse, DeployContractsRequest } from '../types/deployment';
import NodeManagementService, { NodeManagementConfig, NodeConfig, NodeStartupResult } from './nodeService';
import ContractsService, { ContractDeploymentConfig } from './contractsService';

export interface OrbitDeploymentConfig {
  parentChainRpc: string;
  parentChainId: number;
  deployerPrivateKey: string;
  batchPosterPrivateKey: string;
  validatorPrivateKey: string;
  parentChainBeaconRpcUrl?: string;
  nodeConfigDir: string;
  defaultRpcPort: number;
  defaultExplorerPort: number;
  contractArtifactsBaseUrl?: string;
}

export interface OrbitDeploymentResult {
  success: boolean;
  chainAddress?: string;
  deploymentTxHash?: string;
  confirmationTxHash?: string;
  rpcUrl?: string;
  explorerUrl?: string;
  nodeConfigPath?: string;
  contracts?: any[];
  error?: string;
}

class OrbitService {
  private config: OrbitDeploymentConfig;
  private publicClient: any;
  private deployerWallet: any;
  private batchPosterWallet: any;
  private validatorWallet: any;
  private nodeService!: NodeManagementService;
  private contractsService?: ContractsService;

  constructor(config: OrbitDeploymentConfig) {
    this.config = config;
    this.initializeClients();
    this.initializeServices();
  }

  private initializeClients() {
    try {
      logger.info('Starting Orbit service client initialization');
      
      // Validate required configuration
      if (!this.config.deployerPrivateKey || this.config.deployerPrivateKey.trim() === '') {
        throw new Error('Deployer private key is required');
      }

      if (!this.config.parentChainRpc || this.config.parentChainRpc.trim() === '') {
        throw new Error('Parent chain RPC URL is required');
      }

      logger.info('Configuration validation passed', {
        hasDeployerKey: !!this.config.deployerPrivateKey,
        rpcUrl: this.config.parentChainRpc
      });

      // Create public client for reading from parent chain
      logger.info('Creating public client...');
      this.publicClient = createPublicClient({
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });
      logger.info('Public client created successfully');

      // Create deployer wallet client
      logger.info('Creating deployer wallet client...');
      const deployerKey = this.config.deployerPrivateKey.startsWith('0x') 
        ? this.config.deployerPrivateKey 
        : `0x${this.config.deployerPrivateKey}`;
      
      this.deployerWallet = createWalletClient({
        account: privateKeyToAccount(deployerKey as `0x${string}`),
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });
      logger.info('Deployer wallet client created successfully');

      // Use deployer key for batch poster if not provided
      const batchPosterKey = this.config.batchPosterPrivateKey && this.config.batchPosterPrivateKey.trim() !== '' 
        ? this.config.batchPosterPrivateKey 
        : this.config.deployerPrivateKey;

      logger.info('Creating batch poster wallet client...');
      const batchPosterKeyFormatted = batchPosterKey.startsWith('0x') 
        ? batchPosterKey 
        : `0x${batchPosterKey}`;
        
      this.batchPosterWallet = createWalletClient({
        account: privateKeyToAccount(batchPosterKeyFormatted as `0x${string}`),
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });
      logger.info('Batch poster wallet client created successfully');

      // Use deployer key for validator if not provided
      const validatorKey = this.config.validatorPrivateKey && this.config.validatorPrivateKey.trim() !== '' 
        ? this.config.validatorPrivateKey 
        : this.config.deployerPrivateKey;

      logger.info('Creating validator wallet client...');
      const validatorKeyFormatted = validatorKey.startsWith('0x') 
        ? validatorKey 
        : `0x${validatorKey}`;
        
      this.validatorWallet = createWalletClient({
        account: privateKeyToAccount(validatorKeyFormatted as `0x${string}`),
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpc)
      });
      logger.info('Validator wallet client created successfully');

      logger.info('Orbit service clients initialized successfully', {
        usingSameKeyForAllRoles: batchPosterKey === this.config.deployerPrivateKey && validatorKey === this.config.deployerPrivateKey
      });
    } catch (error) {
      logger.error('Failed to initialize Orbit service clients', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined,
        deployerKeyProvided: !!this.config.deployerPrivateKey,
        rpcUrlProvided: !!this.config.parentChainRpc
      });
      throw new Error(`Failed to initialize Orbit service clients: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private initializeServices() {
    try {
      logger.info('Initializing Orbit service dependencies');
      
      // Initialize NodeManagementService
      const nodeConfig: NodeManagementConfig = {
        deployerPrivateKey: this.config.deployerPrivateKey,
        batchPosterPrivateKey: this.config.batchPosterPrivateKey,
        validatorPrivateKey: this.config.validatorPrivateKey,
        parentChainRpcUrl: this.config.parentChainRpc,
        parentChainBeaconRpcUrl: this.config.parentChainBeaconRpcUrl,
        nodeConfigDir: this.config.nodeConfigDir,
        defaultRpcPort: this.config.defaultRpcPort,
        defaultExplorerPort: this.config.defaultExplorerPort
      };
      
      this.nodeService = new NodeManagementService(nodeConfig);
      
      // Initialize ContractsService if contract artifacts URL is provided
      if (this.config.contractArtifactsBaseUrl) {
        const contractConfig: ContractDeploymentConfig = {
          deployerPrivateKey: this.config.deployerPrivateKey,
          contractArtifactsBaseUrl: this.config.contractArtifactsBaseUrl,
          jobRegistryAddress: '', // Will be set after deployment
          gasRegistryAddress: '', // Will be set after deployment
          taskExecutionSpokeAddress: '' // Will be set after deployment
        };
        
        this.contractsService = new ContractsService(contractConfig);
      }
      
      logger.info('Orbit service dependencies initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize Orbit service dependencies', { error });
      throw new Error(`Failed to initialize Orbit service dependencies: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Deploy a new Arbitrum Orbit chain with complete setup
   * Implements the correct deployment flow: Orbit deployment -> Node config -> Node startup -> Contract deployment
   */
  async deployChain(request: DeployChainRequest): Promise<OrbitDeploymentResult> {
    try {
      logger.info('Starting complete Orbit chain deployment', { 
        deploymentId: request.deployment_id,
        chainName: request.chain_name,
        chainId: request.chain_id 
      });

      // Validate deployment request
      const validationResult = this.validateDeploymentRequest(request);
      if (!validationResult.valid) {
        return {
          success: false,
          error: validationResult.error
        };
      }

      // Step 1: Deploy the Orbit chain using Orbit SDK
      logger.info('Step 1: Deploying Orbit chain', { deploymentId: request.deployment_id });
      const orbitConfig = this.createOrbitConfig(request);
      const orbitDeploymentResult = await this.executeOrbitDeployment(orbitConfig);
      
      if (!orbitDeploymentResult.success || !orbitDeploymentResult.deploymentTxHash) {
        logger.error('Step 1 failed: Orbit deployment failed', {
          deploymentId: request.deployment_id,
          error: orbitDeploymentResult.error
        });
        return {
          success: false,
          error: `Orbit deployment failed: ${orbitDeploymentResult.error}`
        };
      }

      logger.info('Step 1 completed: Orbit chain deployed', {
        deploymentId: request.deployment_id,
        chainAddress: orbitDeploymentResult.chainAddress,
        deploymentTxHash: orbitDeploymentResult.deploymentTxHash
      });

      // Step 2: Generate node configuration
      logger.info('Step 2: Generating node configuration', { deploymentId: request.deployment_id });
      let nodeConfig: any;
      let nodeStartupResult: any;
      
      try {
        nodeConfig = await this.nodeService.generateNodeConfig(
          orbitDeploymentResult.deploymentTxHash,
          request.chain_name
        );
        
        logger.info('Step 2 completed: Node configuration generated', { deploymentId: request.deployment_id });

        // Step 3: Start the node
        logger.info('Step 3: Starting Orbit node', { deploymentId: request.deployment_id });
        nodeStartupResult = await this.nodeService.startNode(nodeConfig);
        
        if (!nodeStartupResult.success || !nodeStartupResult.rpcUrl) {
          logger.error('Step 3 failed: Node startup failed', {
            deploymentId: request.deployment_id,
            error: nodeStartupResult.error
          });
          return {
            success: false,
            error: `Node startup failed: ${nodeStartupResult.error}`
          };
        }

        logger.info('Step 3 completed: Orbit node started', {
          deploymentId: request.deployment_id,
          rpcUrl: nodeStartupResult.rpcUrl,
          explorerUrl: nodeStartupResult.explorerUrl
        });
      } catch (error) {
        logger.error('Step 2 or 3 failed: Node configuration or startup failed', {
          deploymentId: request.deployment_id,
          error: error instanceof Error ? error.message : 'Unknown error'
        });
        return {
          success: false,
          error: `Node configuration or startup failed: ${error instanceof Error ? error.message : 'Unknown error'}`
        };
      }

      // Step 4: Deploy TriggerX contracts (if contracts service is available)
      let contractsResult: any[] = [];
      if (this.contractsService) {
        logger.info('Step 4: Deploying TriggerX contracts', { deploymentId: request.deployment_id });
        
        const contractsRequest: DeployContractsRequest = {
          deployment_id: request.deployment_id,
          chain_address: orbitDeploymentResult.chainAddress!,
          contracts: [
            { name: 'JobRegistry', bytecode: '' },
            { name: 'TriggerGasRegistry', bytecode: '' },
            { name: 'TaskExecutionSpoke', bytecode: '' }
          ]
        };
        
        const contractDeploymentResult = await this.contractsService.deployContracts(contractsRequest);
        
        if (contractDeploymentResult.success && contractDeploymentResult.contracts) {
          contractsResult = contractDeploymentResult.contracts;
          logger.info('Step 4 completed: TriggerX contracts deployed', {
            deploymentId: request.deployment_id,
            contractsDeployed: contractsResult.length
          });
        } else {
          logger.warn('Step 4 failed: Contract deployment failed', {
            deploymentId: request.deployment_id,
            error: contractDeploymentResult.error
          });
          // Don't fail the entire deployment if contracts fail
        }
      } else {
        logger.info('Step 4 skipped: No contracts service configured', { deploymentId: request.deployment_id });
      }

      logger.info('Complete Orbit chain deployment finished successfully', {
        deploymentId: request.deployment_id,
        chainName: request.chain_name,
        chainAddress: orbitDeploymentResult.chainAddress,
        rpcUrl: nodeStartupResult.rpcUrl,
        contractsDeployed: contractsResult.length
      });

      return {
        success: true,
        chainAddress: orbitDeploymentResult.chainAddress,
        deploymentTxHash: orbitDeploymentResult.deploymentTxHash,
        confirmationTxHash: orbitDeploymentResult.confirmationTxHash,
        rpcUrl: nodeStartupResult.rpcUrl,
        explorerUrl: nodeStartupResult.explorerUrl,
        nodeConfigPath: nodeStartupResult.nodeConfigPath,
        contracts: contractsResult
      };

    } catch (error) {
      logger.error('Complete Orbit chain deployment error', {
        deploymentId: request.deployment_id,
        error: error instanceof Error ? error.message : 'Unknown error'
      });

      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown deployment error'
      };
    }
  }

  /**
   * Validate deployment request parameters
   */
  private validateDeploymentRequest(request: DeployChainRequest): { valid: boolean; error?: string } {
    if (!request.chain_name || request.chain_name.trim().length === 0) {
      return { valid: false, error: 'Chain name is required' };
    }

    if (!request.chain_id || request.chain_id <= 0) {
      return { valid: false, error: 'Valid chain ID is required' };
    }

    if (!request.owner_address || !this.isValidAddress(request.owner_address)) {
      return { valid: false, error: 'Valid owner address is required' };
    }

    // Note: batch_poster and validator are now optional since we use wallet addresses
    // Only validate if they are provided
    if (request.batch_poster && !this.isValidAddress(request.batch_poster)) {
      return { valid: false, error: 'Valid batch poster address is required if provided' };
    }

    if (request.validator && !this.isValidAddress(request.validator)) {
      return { valid: false, error: 'Valid validator address is required if provided' };
    }

    if (!request.user_address || !this.isValidAddress(request.user_address)) {
      return { valid: false, error: 'Valid user address is required' };
    }

    // Validate custom token parameters if provided
    if (request.native_token && request.native_token.trim().length > 0) {
      if (!this.isValidAddress(request.native_token)) {
        return { valid: false, error: 'Valid native token address is required' };
      }

      if (!request.token_name || request.token_name.trim().length === 0) {
        return { valid: false, error: 'Token name is required when using custom token' };
      }

      if (!request.token_symbol || request.token_symbol.trim().length === 0) {
        return { valid: false, error: 'Token symbol is required when using custom token' };
      }

      if (!request.token_decimals || request.token_decimals < 0 || request.token_decimals > 18) {
        return { valid: false, error: 'Valid token decimals (0-18) is required when using custom token' };
      }
    }

    return { valid: true };
  }

  /**
   * Create Orbit chain configuration from deployment request
   */
  private createOrbitConfig(request: DeployChainRequest) {
    const config = {
      // Basic chain configuration
      name: request.chain_name,
      chainId: request.chain_id,
      owner: request.owner_address as Address,
      // Use wallet addresses from initialized clients instead of request
      batchPoster: this.batchPosterWallet.account.address as Address,
      validator: this.validatorWallet.account.address as Address,

      // Data availability configuration
      dataAvailability: {
        mode: 'sequencer' as const,
        l1GasPriceOracle: '0x' // Will be set by Orbit SDK
      },

      // Chain parameters
      maxDataSize: request.max_data_size || 104857, // Default 104KB
      maxFeePerGasForRetryables: request.max_fee_per_gas_for_retryables || '100000000', // 0.1 gwei

      // Native token configuration
      nativeToken: request.native_token && request.native_token.trim().length > 0 
        ? {
            address: request.native_token as Address,
            name: request.token_name!,
            symbol: request.token_symbol!,
            decimals: request.token_decimals!
          }
        : undefined, // Use ETH if no custom token specified
    };

    logger.info('Created Orbit chain configuration', { 
      chainName: config.name,
      chainId: config.chainId,
      hasCustomToken: !!config.nativeToken,
      batchPoster: config.batchPoster,
      validator: config.validator,
      owner: config.owner,
      batchPosterWalletAddress: this.batchPosterWallet.account.address,
      validatorWalletAddress: this.validatorWallet.account.address
    });

    return config;
  }

  /**
   * Execute the actual Orbit chain deployment using Orbit SDK
   */
  private async executeOrbitDeployment(config: any): Promise<OrbitDeploymentResult> {
    try {
      logger.info('Executing Orbit chain deployment with SDK', { chainName: config.name });

      // Dynamic import of Orbit SDK functions
      const { 
        createRollup,
        createRollupPrepareDeploymentParamsConfig,
        prepareChainConfig
      } = await import('@arbitrum/orbit-sdk');

      // Prepare chain configuration for Orbit SDK
      const chainConfig = prepareChainConfig({
        chainId: config.chainId,
        arbitrum: {
          InitialChainOwner: config.owner,
          DataAvailabilityCommittee: true,
        },
      });

      logger.info('Prepared chain configuration', { chainConfig });

      // Prepare deployment parameters (matching example structure)
      const createRollupConfig = createRollupPrepareDeploymentParamsConfig(
        this.publicClient,
        {
          chainId: BigInt(config.chainId),
          owner: config.owner,
          chainConfig,
        }
      );

      logger.info('Prepared deployment parameters', { createRollupConfig });

      // Deploy the rollup using Orbit SDK (matching example structure)
      logger.info('Starting rollup deployment...');
      
      // Build rollup parameters, only including nativeToken if custom token is specified
      const rollupParams: any = {
        config: createRollupConfig,
        batchPosters: [config.batchPoster],
        validators: [config.validator],
      };
      
      // Only add nativeToken if custom token is specified
      if (config.nativeToken?.address) {
        rollupParams.nativeToken = config.nativeToken.address;
      }
      
      logger.info('Created rollup parameters', { 
        batchPosters: rollupParams.batchPosters,
        validators: rollupParams.validators,
        hasNativeToken: !!rollupParams.nativeToken,
        batchPosterType: typeof rollupParams.batchPosters[0],
        validatorType: typeof rollupParams.validators[0],
        batchPosterValue: rollupParams.batchPosters[0],
        validatorValue: rollupParams.validators[0]
      });
      
      const deploymentResult = await createRollup({
        params: rollupParams,
        account: this.deployerWallet.account,
        parentChainPublicClient: this.publicClient,
      });

      logger.info('Rollup deployment completed', {
        transactionHash: deploymentResult.transactionReceipt.transactionHash,
        coreContracts: deploymentResult.coreContracts
      });

      return {
        success: true,
        chainAddress: deploymentResult.coreContracts.rollup,
        deploymentTxHash: deploymentResult.transactionReceipt.transactionHash,
        confirmationTxHash: deploymentResult.transactionReceipt.transactionHash
      };

    } catch (error) {
      logger.error('Orbit SDK deployment execution failed', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Deployment execution failed'
      };
    }
  }

  /**
   * Get deployment status and chain information
   * Implements FR-004: Return deployment status and chain information
   */
  async getDeploymentStatus(chainAddress: string): Promise<{
    status: DeploymentStatus;
    chainInfo?: any;
    error?: string;
  }> {
    try {
      logger.info('Getting deployment status', { chainAddress });

      // TODO: Implement actual status checking
      // This would query the deployed chain and return its status
      
      // For now, return mock status
      return {
        status: DeploymentStatus.COMPLETED,
        chainInfo: {
          address: chainAddress,
          isActive: true,
          blockNumber: 1,
          gasPrice: '1000000000' // 1 gwei
        }
      };

    } catch (error) {
      logger.error('Failed to get deployment status', { chainAddress, error });
      return {
        status: DeploymentStatus.FAILED,
        error: error instanceof Error ? error.message : 'Status check failed'
      };
    }
  }

  /**
   * Validate Ethereum address format
   */
  private isValidAddress(address: string): boolean {
    return /^0x[a-fA-F0-9]{40}$/.test(address);
  }

  /**
   * Get chain RPC URL for deployed chain
   * Now uses the actual running node RPC URL from NodeManagementService
   */
  async getChainRpcUrl(chainAddress: string, chainName?: string): Promise<string> {
    try {
      // Try to get RPC URL from running nodes
      const runningNodes = this.nodeService.getRunningNodes();
      const node = runningNodes.find(n => n.chainName === chainName);
      
      if (node && node.rpcUrl) {
        logger.info('Found running node RPC URL', { chainAddress, chainName, rpcUrl: node.rpcUrl });
        return node.rpcUrl;
      }
      
      // Fallback to default URL format if no running node found
      const fallbackUrl = `http://localhost:${this.config.defaultRpcPort}`;
      logger.warn('No running node found, using fallback RPC URL', { chainAddress, chainName, fallbackUrl });
      return fallbackUrl;
      
    } catch (error) {
      logger.error('Failed to get chain RPC URL', { chainAddress, chainName, error });
      throw new Error(`Failed to get chain RPC URL: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Check if chain is ready for contract deployment
   * Now checks if the node is running and RPC is accessible
   */
  async isChainReady(chainAddress: string, chainName?: string): Promise<boolean> {
    try {
      // Get RPC URL for the chain
      const rpcUrl = await this.getChainRpcUrl(chainAddress, chainName);
      
      // Check if the node is ready using NodeManagementService
      const isReady = await this.nodeService.isNodeReady(rpcUrl);
      
      logger.info('Chain readiness check completed', { chainAddress, chainName, rpcUrl, isReady });
      return isReady;
      
    } catch (error) {
      logger.error('Failed to check chain readiness', { chainAddress, chainName, error });
      return false;
    }
  }

  /**
   * Get detailed node status for a deployed chain
   */
  async getNodeStatus(chainAddress: string, chainName?: string): Promise<any> {
    try {
      const rpcUrl = await this.getChainRpcUrl(chainAddress, chainName);
      return await this.nodeService.getNodeStatus(rpcUrl);
    } catch (error) {
      logger.error('Failed to get node status', { chainAddress, chainName, error });
      return {
        isRunning: false,
        error: error instanceof Error ? error.message : 'Status check failed'
      };
    }
  }

  /**
   * Stop a specific chain node
   */
  async stopChainNode(chainName: string): Promise<void> {
    try {
      await this.nodeService.stopNode(chainName);
      logger.info('Chain node stopped successfully', { chainName });
    } catch (error) {
      logger.error('Failed to stop chain node', { chainName, error });
      throw error;
    }
  }

  /**
   * Stop all running chain nodes
   */
  async stopAllChainNodes(): Promise<void> {
    try {
      await this.nodeService.stopAllNodes();
      logger.info('All chain nodes stopped successfully');
    } catch (error) {
      logger.error('Failed to stop all chain nodes', { error });
      throw error;
    }
  }

  /**
   * Get list of all running nodes
   */
  getRunningNodes(): Array<{ chainName: string; rpcUrl: string; explorerUrl: string }> {
    return this.nodeService.getRunningNodes();
  }
}

export default OrbitService;
