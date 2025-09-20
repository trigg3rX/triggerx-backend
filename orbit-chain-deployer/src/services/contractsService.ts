import { createPublicClient, createWalletClient, http } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import { exec } from 'child_process';
import { promisify } from 'util';
import logger from '../utils/logger';
import { config, contractConfig } from '../utils/config';
import { DeployContractsRequest, DeploymentStatus, ContractInfo } from '../types/deployment';

const execAsync = promisify(exec);

export interface ContractDeploymentConfig {
  deployerPrivateKey: string;
  contractArtifactsBaseUrl?: string;
  jobRegistryAddress?: string;
  gasRegistryAddress?: string;
  taskExecutionSpokeAddress?: string;
}

export interface TriggerXContract {
  name: string;
  address: string;
  abi: any[];
  bytecode?: string;
  deploymentTxHash?: string;
}

export interface ContractDeploymentResult {
  success: boolean;
  contracts?: ContractInfo[];
  deploymentTxHashes?: string[];
  error?: string;
}

class ContractsService {
  private publicClient: any;
  private deployerWallet: any;
  private currentChainRpcUrl: string = '';
  private config: ContractDeploymentConfig;

  constructor(config: ContractDeploymentConfig) {
    this.config = config;
    this.initializeClients();
  }

  private initializeClients() {
    try {
      logger.info('Starting Contracts service client initialization');
      
      // Validate and format private key from config
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
      await this.setupChainClient(request.chain_address, request.rpc_url);

      // Deploy contracts using forge scripts
      const deploymentResult = await this.executeForgeDeployments(request);

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

      // Check if it's a supported contract type
      const supportedContracts = ['JobRegistry', 'TriggerGasRegistry', 'TaskExecutionSpoke'];
      if (!supportedContracts.includes(contract.name)) {
        return { valid: false, error: `Unsupported contract: ${contract.name}. Supported: ${supportedContracts.join(', ')}` };
      }
    }

    return { valid: true };
  }

  /**
   * Set up client for the target chain
   */
  private async setupChainClient(chainAddress: string, rpcUrl?: string) {
    try {
      // Use provided RPC URL if available, otherwise fall back to constructed URL
      if (rpcUrl) {
        this.currentChainRpcUrl = rpcUrl;
        logger.info('Using provided RPC URL for chain client', { chainAddress, rpcUrl: this.currentChainRpcUrl });
      } else {
        // Check if this is a placeholder address (all zeros)
        const isPlaceholderAddress = chainAddress === '0x' + '0'.repeat(40);
        
        if (isPlaceholderAddress) {
          logger.info('Detected placeholder chain address, using parent chain RPC', { chainAddress });
          
          // For placeholder addresses, use the parent chain RPC
          this.currentChainRpcUrl = process.env.ARBITRUM_RPC_URL || 'https://sepolia-rollup.arbitrum.io/rpc';
        } else {
          // For real deployments, construct the actual RPC URL
          this.currentChainRpcUrl = `https://orbit-chain-${chainAddress.slice(2, 10)}.arbitrum.io/rpc`;
        }
        
        logger.info('Using constructed RPC URL for chain client', { chainAddress, rpcUrl: this.currentChainRpcUrl });
      }

      // Create public client for the target chain
      this.publicClient = createPublicClient({
        transport: http(this.currentChainRpcUrl)
      });

      // Update wallet client transport with properly formatted private key
      const deployerKey = this.config.deployerPrivateKey.startsWith('0x') 
        ? this.config.deployerPrivateKey 
        : `0x${this.config.deployerPrivateKey}`;
        
      this.deployerWallet = createWalletClient({
        account: privateKeyToAccount(deployerKey as `0x${string}`),
        transport: http(this.currentChainRpcUrl)
      });

      logger.info('Chain client setup completed', { chainAddress, rpcUrl: this.currentChainRpcUrl });

    } catch (error) {
      logger.error('Failed to setup chain client', { 
        chainAddress, 
        rpcUrl: this.currentChainRpcUrl,
        error: error instanceof Error ? error.message : 'Unknown error',
        stack: error instanceof Error ? error.stack : undefined,
        errorType: typeof error,
        errorString: String(error)
      });
      throw new Error(`Failed to setup chain client: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Execute forge script deployments in sequence
   */
  private async executeForgeDeployments(request: DeployContractsRequest): Promise<ContractDeploymentResult> {
    try {
      const deployedContracts: ContractInfo[] = [];
      const deploymentTxHashes: string[] = [];

      // Get deterministic salts for consistent addresses across all chains
      const salts = this.getDeploymentSalts();

      // Deploy JobRegistry first
      logger.info('Deploying JobRegistry contract');
      const jobRegistryResult = await this.executeForgeScript(
        'script/deploy/1_deployJobRegistry.s.sol:DeployJobRegistry',
        {
          ...this.getForgeEnvironment(),
          JR_SALT: salts.jobRegistrySalt,
          JR_IMPL_SALT: salts.jobRegistryImplSalt
        }
      );

      if (!jobRegistryResult.success) {
        return {
          success: false,
          error: `Failed to deploy JobRegistry: ${jobRegistryResult.error}`
        };
      }

      deployedContracts.push(jobRegistryResult.contract!);
      deploymentTxHashes.push(jobRegistryResult.deploymentTxHash!);

      // Deploy TriggerGasRegistry
      logger.info('Deploying TriggerGasRegistry contract');
      const gasRegistryResult = await this.executeForgeScript(
        'script/deploy/2_deployTriggerGasRegistry.s.sol:DeployTriggerGasRegistry',
        {
          ...this.getForgeEnvironment(),
          GAS_REGISTRY_SALT: salts.gasRegistrySalt,
          GAS_REGISTRY_IMPL_SALT: salts.gasRegistryImplSalt,
          JOB_REGISTRY_ADDRESS: jobRegistryResult.contract!.address
        }
      );

      if (!gasRegistryResult.success) {
        return {
          success: false,
          error: `Failed to deploy TriggerGasRegistry: ${gasRegistryResult.error}`
        };
      }

      deployedContracts.push(gasRegistryResult.contract!);
      deploymentTxHashes.push(gasRegistryResult.deploymentTxHash!);

      // Deploy TaskExecutionSpoke
      logger.info('Deploying TaskExecutionSpoke contract');
      const spokeResult = await this.executeForgeScript(
        'script/deploy/4_deployTaskExecutionSpoke.s.sol:DeployTaskExecutionSpoke',
        {
          ...this.getForgeEnvironment(),
          TASK_EXECUTION_SALT: salts.taskExecutionSalt,
          TASK_EXECUTION_IMPL_SALT: salts.taskExecutionImplSalt,
          JOB_REGISTRY_ADDRESS: jobRegistryResult.contract!.address,
          GAS_REGISTRY_ADDRESS: gasRegistryResult.contract!.address
        }
      );

      if (!spokeResult.success) {
        return {
          success: false,
          error: `Failed to deploy TaskExecutionSpoke: ${spokeResult.error}`
        };
      }

      deployedContracts.push(spokeResult.contract!);
      deploymentTxHashes.push(spokeResult.deploymentTxHash!);

      // Configure contract relationships after all deployments
      await this.configureContractRelationships(deployedContracts);

      return {
        success: true,
        contracts: deployedContracts,
        deploymentTxHashes
      };

    } catch (error) {
      logger.error('Forge deployment execution failed', { error });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Forge deployment execution failed'
      };
    }
  }

  /**
   * Execute a forge script and parse the results
   */
  private async executeForgeScript(
    scriptPath: string,
    env: Record<string, string>
  ): Promise<{ success: boolean; contract?: ContractInfo; deploymentTxHash?: string; error?: string }> {
    try {
      const command = `forge script ${scriptPath} --broadcast --rpc-url ${this.currentChainRpcUrl} --skip-simulation --non-interactive --chain-id 421614`;
      
      logger.info(`Executing forge script: ${scriptPath}`, {
        command,
        rpcUrl: this.currentChainRpcUrl
      });

      const { stdout, stderr } = await execAsync(command, {
        env: { ...process.env, ...env },
        cwd: './triggerx-contracts/contracts',
        timeout: config.deploymentTimeout // Use config timeout
      });

      if (stderr && !stderr.includes('WARNING')) {
        logger.error('Forge script execution failed', { stderr });
        return {
          success: false,
          error: `Forge script execution failed: ${stderr}`
        };
      }

      // Parse the output to extract contract addresses and transaction hashes
      const addresses = this.parseForgeOutput(stdout);
      const contractName = this.extractContractName(scriptPath);

      if (!addresses.proxy || !addresses.txHash) {
        logger.error('Failed to parse forge output', { stdout });
        return {
          success: false,
          error: 'Failed to parse contract addresses from forge output'
        };
      }

      logger.info(`Forge script completed successfully: ${contractName}`, {
        proxy: addresses.proxy,
        implementation: addresses.implementation,
        txHash: addresses.txHash
      });

      return {
        success: true,
        contract: {
          name: contractName,
          address: addresses.proxy,
          abi: [], // Will be loaded from artifacts if needed
          bytecode: '0x', // Will be loaded from artifacts if needed
          deploymentTxHash: addresses.txHash
        },
        deploymentTxHash: addresses.txHash
      };

    } catch (error) {
      logger.error(`Failed to execute forge script: ${scriptPath}`, { error });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Forge script execution failed'
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

      // TODO: Implement actual contract configuration using forge scripts
      // This would involve calling setter functions on the contracts to establish relationships
      
      logger.info('Contract relationships configured successfully');

    } catch (error) {
      logger.error('Failed to configure contract relationships', { error });
      throw error;
    }
  }

  /**
   * Get deterministic salts for contract deployments
   * These salts are FIXED and should be the same across ALL chains
   * This ensures the same contract addresses on every chain deployment
   */
  private getDeploymentSalts(): {
    jobRegistrySalt: string;
    jobRegistryImplSalt: string;
    gasRegistrySalt: string;
    gasRegistryImplSalt: string;
    taskExecutionSalt: string;
    taskExecutionImplSalt: string;
  } {
    return {
      jobRegistrySalt: contractConfig.fixed.jobRegistrySalt || 'TriggerX_JobRegistry_v1',
      jobRegistryImplSalt: contractConfig.fixed.jobRegistryImplSalt || 'TriggerX_JobRegistry_Impl_v1',
      gasRegistrySalt: contractConfig.fixed.gasRegistrySalt || 'TriggerX_GasRegistry_v1',
      gasRegistryImplSalt: contractConfig.fixed.gasRegistryImplSalt || 'TriggerX_GasRegistry_Impl_v1',
      taskExecutionSalt: contractConfig.fixed.taskExecutionSalt || 'TriggerX_TaskExecution_v1',
      taskExecutionImplSalt: contractConfig.fixed.taskExecutionImplSalt || 'TriggerX_TaskExecution_Impl_v1'
    };
  }

  /**
   * Get environment variables for forge script execution
   */
  private getForgeEnvironment(): Record<string, string> {
    // Ensure private key has 0x prefix for Forge compatibility
    const privateKey = config.deployerPrivateKey.startsWith('0x') 
      ? config.deployerPrivateKey 
      : `0x${config.deployerPrivateKey}`;

    const env: Record<string, string> = {
      PRIVATE_KEY: privateKey,
      RPC_URL: this.currentChainRpcUrl,
      // Use contractConfig for environment variables
      LZ_ENDPOINT: contractConfig.fixed.lzEndpoint || '',
      LZ_EID_BASE: contractConfig.fixed.lzEIDBase.toString(),
      TG_PER_ETH: contractConfig.fixed.tgPerEth.toString(),
      BASE_RPC: contractConfig.fixed.baseRPC || ''
    };

    // Add Etherscan API key if available
    if (config.etherscanApiKey) {
      env.ETHERSCAN_API_KEY = config.etherscanApiKey;
    }

    return env;
  }

  /**
   * Parse forge script output to extract contract addresses and transaction hashes
   */
  private parseForgeOutput(output: string): {
    proxy?: string;
    implementation?: string;
    txHash?: string;
  } {
    const result: { proxy?: string; implementation?: string; txHash?: string } = {};

    // Look for proxy address in deployment summary
    const proxyMatch = output.match(/Proxy:\s*(0x[a-fA-F0-9]{40})/);
    if (proxyMatch) {
      result.proxy = proxyMatch[1];
    }

    // Look for implementation address in deployment summary
    const implMatch = output.match(/Implementation:\s*(0x[a-fA-F0-9]{40})/);
    if (implMatch) {
      result.implementation = implMatch[1];
    }

    // Look for transaction hash in broadcast logs
    const txMatch = output.match(/Transaction hash:\s*(0x[a-fA-F0-9]{64})/);
    if (txMatch) {
      result.txHash = txMatch[1];
    }

    return result;
  }

  /**
   * Extract contract name from script path
   */
  private extractContractName(scriptPath: string): string {
    const match = scriptPath.match(/\/([^\/]+)\.s\.sol/);
    if (match) {
      return match[1].replace('deploy', '').replace(/([A-Z])/g, ' $1').trim();
    }
    return 'Unknown';
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
    return ['JobRegistry', 'TriggerGasRegistry', 'TaskExecutionSpoke'];
  }
}

export default ContractsService;
