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
  arbitrumRpcUrl?: string;
  deployerPrivateKey?: string;
  
  // Deployment configuration
  deploymentTimeout: number;
  maxRetries: number;
}

// Interface for contract configuration
export interface ContractConfig {
  taskExecutionAddress: string;
  gasRegistryAddress: string;
  jobRegistryAddress: string;
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
    
    // Deployment configuration
    deploymentTimeout: parseInt(process.env.DEPLOYMENT_TIMEOUT || '300000', 10), // 5 minutes default
    maxRetries: parseInt(process.env.MAX_RETRIES || '3', 10),
  };
};

const createContractConfig = (): ContractConfig => {
  return {
    taskExecutionAddress: process.env.TASK_EXECUTION_ADDRESS || '',
    gasRegistryAddress: process.env.GAS_REGISTRY_ADDRESS || '',
    jobRegistryAddress: process.env.JOB_REGISTRY_ADDRESS || '',
  };
};

// Export the configuration
export const config = createConfig();
export const contractConfig = createContractConfig();

// Export configuration validation helper
export const isDevelopment = config.nodeEnv === 'development';
