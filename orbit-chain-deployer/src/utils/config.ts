import dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

// Interface for configuration
export interface Config {
  // Server configuration
  port: number;
  nodeEnv: string;
  
  // Go Backend configuration
  goBackendUrl: string;
  goBackendApiKey: string;
  
  // Arbitrum configuration
  arbitrumRpcUrl: string;
  deployerPrivateKey: string;
  batchPosterPrivateKey: string;
  validatorPrivateKey: string;
  parentChainBeaconRpcUrl?: string;
  
  // Etherscan configuration
  etherscanApiKey?: string;
  
  // Deployment configuration
  deploymentTimeout: number;
  maxRetries: number;
  
  // Node management configuration
  node: NodeConfig;
}

// Interface for node configuration
export interface NodeConfig {
  // Docker configuration
  dockerImage: string;
  containerPrefix: string;
  
  // Binary configuration (fallback)
  nitroBinaryPath: string;
  
  // Resource limits
  memoryLimit: string;
  cpuLimit: string;
  
  // Port configuration
  rpcPortRange: {
    start: number;
    end: number;
  };
  
  // Directory configuration
  nodeConfigDir: string;
  
  // Health check configuration
  healthCheckInterval: number;
  startupTimeout: number;
  
  // Default ports
  defaultRpcPort: number;
  defaultExplorerPort: number;
}

// Interface for contract configuration
export interface ContractConfig {
  // Contract addresses (populated after deployment)
  taskExecutionAddress: string;
  gasRegistryAddress: string;
  jobRegistryAddress: string;
  attesttationCenterAddress: string;
  
  // Fixed configuration for forge scripts
  fixed: {
    baseRPC: string;
    lzEndpoint: string;
    lzEIDBase: number;
    tgPerEth: number;
    // Salt configuration for deterministic deployments
    jobRegistrySalt: string;
    jobRegistryImplSalt: string;
    gasRegistrySalt: string;
    gasRegistryImplSalt: string;
    taskExecutionSalt: string;
    taskExecutionImplSalt: string;
  };
}

// Validate required environment variables
const validateConfig = (): void => {
  if (!process.env.GO_BACKEND_API_KEY) {
    throw new Error('Missing required environment variable: GO_BACKEND_API_KEY');
  }
  if (!process.env.ARBITRUM_RPC_URL) {
    throw new Error('Missing required environment variable: ARBITRUM_RPC_URL');
  }
  if (!process.env.DEPLOYER_PRIVATE_KEY) {
    throw new Error('Missing required environment variable: DEPLOYER_PRIVATE_KEY');
  }
  if (!process.env.BATCH_POSTER_PRIVATE_KEY) {
    throw new Error('Missing required environment variable: BATCH_POSTER_PRIVATE_KEY');
  }
  if (!process.env.VALIDATOR_PRIVATE_KEY) {
    throw new Error('Missing required environment variable: VALIDATOR_PRIVATE_KEY');
  }
  
  // Ensure all three private keys are different
  const deployerKey = process.env.DEPLOYER_PRIVATE_KEY!;
  const batchPosterKey = process.env.BATCH_POSTER_PRIVATE_KEY!;
  const validatorKey = process.env.VALIDATOR_PRIVATE_KEY!;
  
  if (deployerKey === batchPosterKey) {
    throw new Error('DEPLOYER_PRIVATE_KEY and BATCH_POSTER_PRIVATE_KEY must be different');
  }
  if (deployerKey === validatorKey) {
    throw new Error('DEPLOYER_PRIVATE_KEY and VALIDATOR_PRIVATE_KEY must be different');
  }
  if (batchPosterKey === validatorKey) {
    throw new Error('BATCH_POSTER_PRIVATE_KEY and VALIDATOR_PRIVATE_KEY must be different');
  }
};

// Create configuration object
const createConfig = (): Config => {
  validateConfig();

  return {
    // Server configuration
    port: parseInt(process.env.PORT || '3001', 10),
    nodeEnv: process.env.NODE_ENV || 'development',
    
    // Go Backend configuration
    goBackendUrl: process.env.GO_BACKEND_URL || 'http://localhost:9002',
    goBackendApiKey: process.env.GO_BACKEND_API_KEY!,
    
    // Arbitrum configuration
    arbitrumRpcUrl: process.env.ARBITRUM_RPC_URL!,
    deployerPrivateKey: process.env.DEPLOYER_PRIVATE_KEY!,
    batchPosterPrivateKey: process.env.BATCH_POSTER_PRIVATE_KEY!,
    validatorPrivateKey: process.env.VALIDATOR_PRIVATE_KEY!,
    parentChainBeaconRpcUrl: process.env.ETHEREUM_BEACON_RPC_URL,
    
    // Etherscan configuration
    etherscanApiKey: process.env.ETHERSCAN_API_KEY,
    
    // Deployment configuration
    deploymentTimeout: parseInt(process.env.DEPLOYMENT_TIMEOUT || '300000', 10), // 5 minutes default
    maxRetries: parseInt(process.env.MAX_RETRIES || '3', 10),
    
    // Node management configuration
    node: {
      // Docker configuration
      dockerImage: process.env.NITRO_DOCKER_IMAGE || 'arbitrum/nitro:v3.6.0-fc07dd2',
      containerPrefix: process.env.NODE_CONTAINER_PREFIX || 'orbit-node-',
      
      // Binary configuration (fallback)
      nitroBinaryPath: process.env.NITRO_BINARY_PATH || '/usr/local/bin/nitro',
      
      // Resource limits
      memoryLimit: process.env.NODE_MEMORY_LIMIT || '2g',
      cpuLimit: process.env.NODE_CPU_LIMIT || '1',
      
      // Port configuration
      rpcPortRange: {
        start: parseInt(process.env.RPC_PORT_START || '8449', 10),
        end: parseInt(process.env.RPC_PORT_END || '8549', 10)
      },
      
      // Directory configuration
      nodeConfigDir: process.env.NODE_CONFIG_DIR || './node-configs',
      
      // Health check configuration
      healthCheckInterval: parseInt(process.env.HEALTH_CHECK_INTERVAL || '30000', 10),
      startupTimeout: parseInt(process.env.NODE_STARTUP_TIMEOUT || '1800000', 10), // avg is 15 minutes
      
      // Default ports
      defaultRpcPort: parseInt(process.env.DEFAULT_RPC_PORT || '8449', 10),
      defaultExplorerPort: parseInt(process.env.DEFAULT_EXPLORER_PORT || '8448', 10)
    }
  };
};

const createContractConfig = (): ContractConfig => {
  return {
    // Contract addresses (populated after deployment)
    taskExecutionAddress: process.env.TASK_EXECUTION_ADDRESS || '',
    gasRegistryAddress: process.env.GAS_REGISTRY_ADDRESS || '',
    jobRegistryAddress: process.env.JOB_REGISTRY_ADDRESS || '',
    attesttationCenterAddress: process.env.ATTESTTATION_CENTER_ADDRESS || '',
    
    // Fixed configuration for forge scripts
    fixed: {
      // LayerZero configuration
      baseRPC: process.env.BASE_RPC || '',
      lzEndpoint: process.env.LZ_ENDPOINT || '',
      lzEIDBase: parseInt(process.env.LZ_EID_BASE || '0', 10),
      tgPerEth: parseInt(process.env.TG_PER_ETH || '0', 10),
      
      // Salt configuration for deterministic deployments
      jobRegistrySalt: process.env.JOB_REGISTRY_SALT || '',
      jobRegistryImplSalt: process.env.JOB_REGISTRY_IMPL_SALT || '',
      gasRegistrySalt: process.env.GAS_REGISTRY_SALT || '',
      gasRegistryImplSalt: process.env.GAS_REGISTRY_IMPL_SALT || '',
      taskExecutionSalt: process.env.TASK_EXECUTION_SALT || '',
      taskExecutionImplSalt: process.env.TASK_EXECUTION_IMPL_SALT || ''
    }
  };
};

// Export the configuration
export const config = createConfig();
export const contractConfig = createContractConfig();

// Export configuration validation helper
export const isDevelopment = config.nodeEnv === 'development';
