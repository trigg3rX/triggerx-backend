import dotenv from 'dotenv';

// Load environment variables from .env file
dotenv.config();

// Interface for configuration
export interface Config {
  // Server configuration
  port: number;
  nodeEnv: string;
  
  // Database configuration
  dbServerUrl: string;
  
  // Logging configuration
  logLevel: string;
  
  // Arbitrum configuration
  arbitrumRpcUrl?: string;
  arbitrumPrivateKey?: string;
  
  // Deployment configuration
  deploymentTimeout: number;
  maxRetries: number;
}

// Validate required environment variables
const validateConfig = (): void => {
  if (!process.env.DBSERVER_URL) {
    throw new Error('Missing required environment variable: DBSERVER_URL');
  }
};

// Create configuration object
const createConfig = (): Config => {
  validateConfig();
  
  return {
    // Server configuration
    port: parseInt(process.env.PORT || '3001', 10),
    nodeEnv: process.env.NODE_ENV || 'development',
    
    // Database configuration
    dbServerUrl: process.env.DBSERVER_URL!,
    
    // Logging configuration
    logLevel: process.env.LOG_LEVEL || 'info',
    
    // Arbitrum configuration
    arbitrumRpcUrl: process.env.ARBITRUM_RPC_URL,
    arbitrumPrivateKey: process.env.ARBITRUM_PRIVATE_KEY,
    
    // Deployment configuration
    deploymentTimeout: parseInt(process.env.DEPLOYMENT_TIMEOUT || '300000', 10), // 5 minutes default
    maxRetries: parseInt(process.env.MAX_RETRIES || '3', 10),
  };
};

// Export the configuration
export const config = createConfig();

// Export configuration validation helper
export const isDevelopment = config.nodeEnv === 'development';
