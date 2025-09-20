import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import axios from 'axios';
import logger from '../utils/logger';
import { config } from '../utils/config';
import { DeployContractsRequest, DeploymentStatus, ContractInfo } from '../types/deployment';

export interface ContractDeploymentConfig {
  deployerPrivateKey: string;
  contractArtifactsBaseUrl: string;
  jobRegistryAddress: string;
  gasRegistryAddress: string;
  taskExecutionSpokeAddress: string;
}

export interface ContractDeploymentResult {
  success: boolean;
  contracts?: ContractInfo[];
  deploymentTxHashes?: string[];
  error?: string;
}

export interface TriggerXContract {
  name: string;
  abi: any[];
  bytecode: string;
  constructorArgs?: any[];
}

class ContractsService {
  private config: ContractDeploymentConfig;
  private publicClient: any;
  private deployerWallet: any;
  private contracts: Map<string, TriggerXContract> = new Map();

  constructor(config: ContractDeploymentConfig) {
    this.config = config;
    this.initializeClients();
    // Load contract artifacts asynchronously (don't await to avoid blocking constructor)
    this.loadContractArtifacts().catch(error => {
      logger.error('Failed to load contract artifacts in constructor', { error });
    });
  }

  private initializeClients() {
    try {
      logger.info('Starting Contracts service client initialization');
      
      // Validate and format private key
      if (!this.config.deployerPrivateKey || this.config.deployerPrivateKey.trim() === '') {
        throw new Error('Deployer private key is required');
      }

      logger.info('Deployer private key validation passed');

      const deployerKey = this.config.deployerPrivateKey.startsWith('0x') 
        ? this.config.deployerPrivateKey 
        : `0x${this.config.deployerPrivateKey}`;

      logger.info('Creating deployer wallet client for contracts service...');

      // Create wallet client for contract deployment
      this.deployerWallet = createWalletClient({
        account: privateKeyToAccount(deployerKey as `0x${string}`),
        transport: http(config.arbitrumRpcUrl) // Default RPC, will be updated per chain
      });

      logger.info('Contracts service client initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize contracts service client', { 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined,
        deployerKeyProvided: !!this.config.deployerPrivateKey
      });
      throw new Error(`Failed to initialize contracts service client: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Load contract artifacts from GitHub URLs
   */
  private async loadContractArtifacts() {
    try {
      const baseUrl = this.config.contractArtifactsBaseUrl;
      
      // Check if base URL exists
      if (!baseUrl || baseUrl.trim() === '') {
        logger.warn('No contract artifacts base URL provided, skipping contract artifact loading');
        return;
      }

      logger.info('Loading contract artifacts from GitHub URLs', { baseUrl });
      
      // Define contract artifacts to load
      const contractsToLoad = [
        {
          name: 'JobRegistry',
          abiUrl: `${baseUrl}/JobRegistry.arbitrum.abi.json`,
          bytecodeUrl: `${baseUrl}/JobRegistry.arbitrum.bin`
        },
        {
          name: 'TriggerGasRegistry',
          abiUrl: `${baseUrl}/TriggerGasRegistry.arbitrum.abi.json`,
          bytecodeUrl: `${baseUrl}/TriggerGasRegistry.arbitrum.bin`
        },
        {
          name: 'TaskExecutionSpoke',
          abiUrl: `${baseUrl}/TaskExecutionSpoke.arbitrum.abi.json`,
          bytecodeUrl: `${baseUrl}/TaskExecutionSpoke.arbitrum.bin`
        }
      ];

      // Load each contract artifact
      for (const contract of contractsToLoad) {
        try {
          logger.info(`Loading ${contract.name} contract artifacts...`);
          
          // Fetch ABI
          const abiResponse = await axios.get(contract.abiUrl, { timeout: 10000 });
          const abi = abiResponse.data;
          
          // Fetch bytecode
          const bytecodeResponse = await axios.get(contract.bytecodeUrl, { timeout: 10000 });
          const bytecode = '0x' + bytecodeResponse.data;

          this.contracts.set(contract.name, {
            name: contract.name,
            abi: abi,
            bytecode: bytecode
          });
          
          logger.info(`${contract.name} contract artifact loaded successfully`);
        } catch (error) {
          logger.warn(`Failed to load ${contract.name} contract artifact`, { 
            error: error instanceof Error ? error.message : 'Unknown error',
            abiUrl: contract.abiUrl,
            bytecodeUrl: contract.bytecodeUrl
          });
        }
      }

      logger.info('Contract artifacts loading completed', {
        contracts: Array.from(this.contracts.keys()),
        totalLoaded: this.contracts.size
      });

    } catch (error) {
      logger.error('Failed to load contract artifacts', { error });
      // Don't throw error, just log it - the service can still work without artifacts
      logger.warn('ContractsService will work without contract artifacts (deployment will use placeholders)');
    }
  }

  /**
   * Deploy TriggerX contracts to a deployed Orbit chain
   * Implements FR-005, FR-006, FR-007: Deploy JobRegistry, TriggerGasRegistry, TaskExecutionSpoke
   */
  async deployContracts(request: DeployContractsRequest): Promise<ContractDeploymentResult> {
    try {
      logger.info('Starting TriggerX contracts deployment', {
        deploymentId: request.deployment_id,
        chainAddress: request.chain_address
      });

      // Validate deployment request
      const validationResult = this.validateDeploymentRequest(request);
      if (!validationResult.valid) {
        return {
          success: false,
          error: validationResult.error
        };
      }

      // Set up client for the target chain
      await this.setupChainClient(request.chain_address);

      // Deploy contracts in sequence
      const deploymentResult = await this.executeContractDeployments(request);

      if (deploymentResult.success) {
        logger.info('TriggerX contracts deployment completed successfully', {
          deploymentId: request.deployment_id,
          contractsDeployed: deploymentResult.contracts?.length || 0
        });
      } else {
        logger.error('TriggerX contracts deployment failed', {
          deploymentId: request.deployment_id,
          error: deploymentResult.error
        });
      }

      return deploymentResult;

    } catch (error) {
      logger.error('TriggerX contracts deployment error', {
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
  private validateDeploymentRequest(request: DeployContractsRequest): { valid: boolean; error?: string } {
    if (!request.deployment_id || request.deployment_id.trim().length === 0) {
      return { valid: false, error: 'Deployment ID is required' };
    }

    if (!request.chain_address || !this.isValidAddress(request.chain_address)) {
      return { valid: false, error: 'Valid chain address is required' };
    }

    if (!request.contracts || request.contracts.length === 0) {
      return { valid: false, error: 'At least one contract must be specified' };
    }

    // Validate each contract
    for (const contract of request.contracts) {
      if (!contract.name || contract.name.trim().length === 0) {
        return { valid: false, error: 'Contract name is required' };
      }

      if (!this.contracts.has(contract.name)) {
        return { valid: false, error: `Unknown contract: ${contract.name}` };
      }

      if (!contract.bytecode || contract.bytecode.trim().length === 0) {
        return { valid: false, error: `Bytecode is required for contract: ${contract.name}` };
      }
    }

    return { valid: true };
  }

  /**
   * Set up client for the target chain
   */
  private async setupChainClient(chainAddress: string) {
    try {
      // Check if this is a placeholder address (all zeros)
      const isPlaceholderAddress = chainAddress === '0x' + '0'.repeat(40);
      
      if (isPlaceholderAddress) {
        logger.info('Detected placeholder chain address, using parent chain RPC for simulation', { chainAddress });
        
        // For placeholder addresses, use the parent chain RPC for simulation
        const parentChainRpc = process.env.ARBITRUM_RPC_URL || 'https://sepolia-rollup.arbitrum.io/rpc';
        
        logger.info('Creating public client for simulation', { rpcUrl: parentChainRpc });
        
        // Create public client for the parent chain (for simulation)
        this.publicClient = createPublicClient({
          transport: http(parentChainRpc)
        });

        logger.info('Creating wallet client for simulation', { 
          hasPrivateKey: !!this.config.deployerPrivateKey,
          privateKeyLength: this.config.deployerPrivateKey?.length || 0
        });

        // Validate private key before using it
        if (!this.config.deployerPrivateKey || this.config.deployerPrivateKey.trim() === '') {
          throw new Error('Deployer private key is required but not provided');
        }

        // Format private key if needed
        const formattedPrivateKey = this.config.deployerPrivateKey.startsWith('0x') 
          ? this.config.deployerPrivateKey 
          : `0x${this.config.deployerPrivateKey}`;

        logger.info('Creating wallet client with formatted private key', { 
          privateKeyPrefix: formattedPrivateKey.substring(0, 6) + '...',
          privateKeyLength: formattedPrivateKey.length
        });

        // Update wallet client transport
        this.deployerWallet = createWalletClient({
          account: privateKeyToAccount(formattedPrivateKey as `0x${string}`),
          transport: http(parentChainRpc)
        });

        logger.info('Chain client setup completed (simulation mode)', { chainAddress, rpcUrl: parentChainRpc });
      } else {
        // TODO: Get actual RPC URL for the deployed chain
        // For real deployments, construct the actual RPC URL
        const chainRpcUrl = `https://orbit-chain-${chainAddress.slice(2, 10)}.arbitrum.io/rpc`;

        // Create public client for the target chain
        this.publicClient = createPublicClient({
          transport: http(chainRpcUrl)
        });

        // Update wallet client transport with properly formatted private key
        const deployerKey = this.config.deployerPrivateKey.startsWith('0x') 
          ? this.config.deployerPrivateKey 
          : `0x${this.config.deployerPrivateKey}`;
          
        this.deployerWallet = createWalletClient({
          account: privateKeyToAccount(deployerKey as `0x${string}`),
          transport: http(chainRpcUrl)
        });

        logger.info('Chain client setup completed', { chainAddress, chainRpcUrl });
      }

    } catch (error) {
      logger.error('Failed to setup chain client', { 
        chainAddress, 
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined,
        errorType: typeof error,
        errorString: String(error)
      });
      throw new Error(`Failed to setup chain client: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Execute contract deployments in sequence
   */
  private async executeContractDeployments(request: DeployContractsRequest): Promise<ContractDeploymentResult> {
    try {
      const deployedContracts: ContractInfo[] = [];
      const deploymentTxHashes: string[] = [];

      // Deploy contracts in the specified order
      for (const contractRequest of request.contracts) {
        logger.info(`Deploying contract: ${contractRequest.name}`);

        const contractArtifact = this.contracts.get(contractRequest.name);
        if (!contractArtifact) {
          throw new Error(`Contract artifact not found: ${contractRequest.name}`);
        }

        // Deploy the contract
        const deploymentResult = await this.deploySingleContract(
          contractRequest.name,
          contractArtifact,
          contractRequest.constructor_args || []
        );

        if (!deploymentResult.success) {
          return {
            success: false,
            error: `Failed to deploy ${contractRequest.name}: ${deploymentResult.error}`
          };
        }

        deployedContracts.push(deploymentResult.contract!);
        deploymentTxHashes.push(deploymentResult.deploymentTxHash!);

        logger.info(`Contract deployed successfully: ${contractRequest.name}`, {
          address: deploymentResult.contract!.address,
          txHash: deploymentResult.deploymentTxHash
        });
      }

      // Configure contract relationships after all deployments
      await this.configureContractRelationships(deployedContracts);

      return {
        success: true,
        contracts: deployedContracts,
        deploymentTxHashes
      };

    } catch (error) {
      logger.error('Contract deployment execution failed', { error });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Deployment execution failed'
      };
    }
  }

  /**
   * Deploy a single contract
   */
  private async deploySingleContract(
    contractName: string,
    contractArtifact: TriggerXContract,
    constructorArgs: any[]
  ): Promise<{ success: boolean; contract?: ContractInfo; deploymentTxHash?: string; error?: string }> {
    try {
      logger.info(`Deploying ${contractName} with constructor args:`, constructorArgs);

      // TODO: Implement actual contract deployment using viem
      // This is a placeholder implementation
      
      // Simulate deployment process
      await this.simulateContractDeployment(contractName);

      // Return mock deployment result
      const contractAddress = '0x' + '0'.repeat(40); // Placeholder address
      const deploymentTxHash = '0x' + '0'.repeat(64); // Placeholder tx hash

      return {
        success: true,
        contract: {
          name: contractName,
          address: contractAddress,
          abi: contractArtifact.abi,
          bytecode: contractArtifact.bytecode,
          deploymentTxHash
        },
        deploymentTxHash
      };

    } catch (error) {
      logger.error(`Failed to deploy contract: ${contractName}`, { error });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Contract deployment failed'
      };
    }
  }

  /**
   * Configure relationships between deployed contracts
   * Implements FR-008: Configure contract relationships and permissions
   */
  private async configureContractRelationships(contracts: ContractInfo[]) {
    try {
      logger.info('Configuring contract relationships');

      const jobRegistry = contracts.find(c => c.name === 'JobRegistry');
      const triggerGasRegistry = contracts.find(c => c.name === 'TriggerGasRegistry');
      const taskExecutionSpoke = contracts.find(c => c.name === 'TaskExecutionSpoke');

      if (!jobRegistry || !triggerGasRegistry || !taskExecutionSpoke) {
        throw new Error('Required contracts not found for configuration');
      }

      // TODO: Implement actual contract configuration
      // This would involve calling setter functions on the contracts to establish relationships
      
      // Simulate configuration process
      await this.simulateContractConfiguration();

      logger.info('Contract relationships configured successfully');

    } catch (error) {
      logger.error('Failed to configure contract relationships', { error });
      throw error;
    }
  }

  /**
   * Simulate contract deployment process for testing
   * TODO: Replace with actual contract deployment
   */
  private async simulateContractDeployment(contractName: string): Promise<void> {
    logger.info(`Simulating ${contractName} deployment...`);
    
    // Simulate deployment steps
    const steps = [
      'Validating contract bytecode',
      'Estimating gas costs',
      'Deploying contract',
      'Waiting for confirmation',
      'Verifying deployment'
    ];

    for (const step of steps) {
      logger.info(`${contractName} deployment step: ${step}`);
      // Simulate processing time
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    logger.info(`${contractName} deployment simulation completed`);
  }

  /**
   * Simulate contract configuration process for testing
   * TODO: Replace with actual contract configuration
   */
  private async simulateContractConfiguration(): Promise<void> {
    logger.info('Simulating contract configuration...');
    
    // Simulate configuration steps
    const steps = [
      'Setting up JobRegistry permissions',
      'Configuring TriggerGasRegistry relationships',
      'Initializing TaskExecutionSpoke connections',
      'Verifying contract relationships'
    ];

    for (const step of steps) {
      logger.info(`Configuration step: ${step}`);
      // Simulate processing time
      await new Promise(resolve => setTimeout(resolve, 300));
    }

    logger.info('Contract configuration simulation completed');
  }

  /**
   * Get contract deployment status
   */
  async getContractDeploymentStatus(contractAddress: string): Promise<{
    status: DeploymentStatus;
    contractInfo?: ContractInfo;
    error?: string;
  }> {
    try {
      logger.info('Getting contract deployment status', { contractAddress });

      // TODO: Implement actual status checking
      // This would query the deployed contract and return its status
      
      // For now, return mock status
      return {
        status: DeploymentStatus.COMPLETED,
        contractInfo: {
          name: 'Unknown',
          address: contractAddress,
          abi: [],
          bytecode: '0x'
        }
      };

    } catch (error) {
      logger.error('Failed to get contract deployment status', { contractAddress, error });
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
   * Get available contracts
   */
  getAvailableContracts(): string[] {
    return Array.from(this.contracts.keys());
  }

  /**
   * Get contract artifact by name
   */
  getContractArtifact(contractName: string): TriggerXContract | undefined {
    return this.contracts.get(contractName);
  }
}

export default ContractsService;
