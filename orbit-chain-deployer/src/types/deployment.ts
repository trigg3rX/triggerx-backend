// Deployment status types
export enum DeploymentStatus {
  PENDING = 'pending',
  DEPLOYING_ORBIT = 'deploying_orbit',
  ORBIT_DEPLOYED = 'orbit_deployed',
  GENERATING_NODE_CONFIG = 'generating_node_config',
  STARTING_NODE = 'starting_node',
  NODE_READY = 'node_ready',
  DEPLOYING_CONTRACTS = 'deploying_contracts',
  CONFIGURING_CONTRACTS = 'configuring_contracts',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled'
}

// Deployment types
export enum DeploymentType {
  CHAIN = 'chain',
  CONTRACTS = 'contracts'
}

// Orbit chain data structure matching database schema
export interface OrbitChainData {
  chain_id: number;
  chain_name: string;
  rpc_url?: string;
  user_address: string;
  deployment_status: DeploymentStatus;
  orbit_chain_address?: string;
  created_at: string;
  updated_at: string;
}

// Base deployment interface
export interface BaseDeployment {
  id: string;
  type: DeploymentType;
  status: DeploymentStatus;
  createdAt: Date;
  updatedAt: Date;
  error?: string;
}

// Chain deployment specific interface
export interface ChainDeployment extends BaseDeployment {
  type: DeploymentType.CHAIN;
  chainId?: string;
  chainAddress?: string;
  orbitConfig?: OrbitChainConfig;
  deploymentTxHash?: string;
  confirmationTxHash?: string;
}

// Contract deployment specific interface
export interface ContractDeployment extends BaseDeployment {
  type: DeploymentType.CONTRACTS;
  chainAddress: string;
  contracts: ContractInfo[];
  deploymentTxHashes: string[];
}

// Orbit chain configuration
export interface OrbitChainConfig {
  name: string;
  symbol: string;
  baseChain: string;
  ownerAddress: string;
  nativeToken?: string;
  gasPriceOracle?: string;
  dataAvailability?: {
    mode: 'sequencer' | 'blobstream';
    l1GasPriceOracle?: string;
  };
}

// Contract information
export interface ContractInfo {
  name: string;
  address: string;
  abi: any[];
  bytecode?: string;
  deploymentTxHash?: string;
}

// Deployment request interfaces
export interface DeployChainRequest {
  deployment_id: string;
  chain_name: string;
  chain_id: number;
  owner_address: string;
  batch_poster?: string;
  validator?: string;
  user_address: string;
  
  // ERC20 Token Details (if custom gas token)
  native_token?: string; // Address of ERC20 token, empty for ETH
  token_name?: string;
  token_symbol?: string;
  token_decimals?: number;
  
  // Optional Chain Configuration
  max_data_size?: number;
  max_fee_per_gas_for_retryables?: string;
}

export interface DeployContractsRequest {
  deployment_id: string;
  chain_address: string;
  rpc_url?: string; // RPC URL of the Orbit chain node
  contracts: {
    name: string;
    bytecode: string;
    constructor_args?: any[];
  }[];
}

// Deployment response interfaces
export interface DeploymentResponse {
  success: boolean;
  deployment_id: string;
  status: DeploymentStatus;
  message: string;
  error?: string;
}

export interface ChainDeploymentResponse extends DeploymentResponse {
  chain_address?: string;
  chain_id?: string;
  deployment_tx_hash?: string;
}

export interface ContractDeploymentResponse extends DeploymentResponse {
  contracts?: ContractInfo[];
  deployment_tx_hashes?: string[];
}

// Chain deployment status - matches Go backend ChainDeploymentStatus
export interface ChainDeploymentStatus {
  chain_id: number;
  chain_name: string;
  user_address: string;
  deployment_status: DeploymentStatus;
  orbit_chain_address?: string;
  created_at: string;
  updated_at: string;
  error_message?: string;
}

// Node process information for real deployment
export interface NodeProcess {
  id: string;
  chainName: string;
  configPath: string;
  rpcUrl: string;
  explorerUrl: string;
  pid: number;
  startTime: Date;
  containerName?: string;
  dataDir?: string;
  logsDir?: string;
  process?: any; // For binary-based deployment
}

// Node management configuration
export interface NodeManagementConfig {
  deployerPrivateKey: string;
  batchPosterPrivateKey: string;
  validatorPrivateKey: string;
  parentChainRpcUrl: string;
  parentChainBeaconRpcUrl?: string;
  nodeConfigDir: string;
  defaultRpcPort: number;
  defaultExplorerPort: number;
  dockerImage?: string;
  containerPrefix?: string;
  nitroBinaryPath?: string;
  memoryLimit?: string;
  cpuLimit?: string;
  startupTimeout?: number;
  healthCheckInterval?: number;
}

// Node startup result
export interface NodeStartupResult {
  success: boolean;
  nodeConfigPath?: string;
  rpcUrl?: string;
  explorerUrl?: string;
  containerName?: string;
  pid?: number;
  error?: string;
}

// Node status information
export interface NodeStatus {
  isRunning: boolean;
  rpcUrl?: string;
  explorerUrl?: string;
  blockNumber?: number;
  chainId?: number;
  containerName?: string;
  pid?: number;
  uptime?: number;
  error?: string;
}

// Docker container information
export interface DockerContainerInfo {
  name: string;
  id: string;
  status: string;
  ports: string[];
  created: string;
  image: string;
}

// Health check response
export interface HealthCheckResponse {
  status: 'healthy' | 'unhealthy';
  timestamp: string;
  version: string;
  environment: string;
  services: {
    goBackend: {
      status: 'connected' | 'disconnected';
      url?: string;
    };
    arbitrum: {
      status: 'connected' | 'disconnected';
      rpc_url?: string;
    };
    docker: {
      status: 'available' | 'unavailable';
      version?: string;
    };
  };
  configuration: {
    port: number;
    deployment_timeout: number;
    max_retries: number;
  };
  uptime: number;
  memory_usage: NodeJS.MemoryUsage;
  node_version: string;
  running_nodes: number;
  error?: string;
}
