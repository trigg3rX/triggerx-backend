import { createPublicClient, http, type Address } from 'viem';
import { arbitrumSepolia } from 'viem/chains';
import { 
  prepareNodeConfig, 
  createRollupPrepareTransaction,
  createRollupPrepareTransactionReceipt,
  type ChainConfig,
  type PrepareNodeConfigParams
} from '@arbitrum/orbit-sdk';
import { getParentChainLayer } from '@arbitrum/orbit-sdk/utils';
import { writeFile, readFile, mkdir, access } from 'fs/promises';
import { constants } from 'fs';
import { join } from 'path';
import axios from 'axios';
import logger from '../utils/logger';
import { config } from '../utils/config';
import { DeploymentStatus } from '../types/deployment';

// Node configuration interfaces
export interface NodeConfig {
  chainConfig: ChainConfig;
  coreContracts: any;
  chainName: string;
  batchPosterPrivateKey: string;
  validatorPrivateKey: string;
  stakeToken?: string;
  parentChainId: number;
  parentChainRpcUrl: string;
  parentChainBeaconRpcUrl?: string;
}

export interface NodeStartupResult {
  success: boolean;
  nodeConfigPath?: string;
  rpcUrl?: string;
  explorerUrl?: string;
  error?: string;
}

export interface NodeStatus {
  isRunning: boolean;
  rpcUrl?: string;
  explorerUrl?: string;
  blockNumber?: number;
  chainId?: number;
  error?: string;
}

export interface NodeManagementConfig {
  deployerPrivateKey: string;
  batchPosterPrivateKey: string;
  validatorPrivateKey: string;
  parentChainRpcUrl: string;
  parentChainBeaconRpcUrl?: string;
  nodeConfigDir: string;
  defaultRpcPort: number;
  defaultExplorerPort: number;
}

class NodeManagementService {
  private config: NodeManagementConfig;
  private publicClient: any;
  private runningNodes: Map<string, NodeProcess> = new Map();

  constructor(config: NodeManagementConfig) {
    this.config = config;
    this.initializeClients();
  }

  private initializeClients() {
    try {
      logger.info('Initializing NodeManagementService clients');
      
      // Create public client for parent chain
      this.publicClient = createPublicClient({
        chain: arbitrumSepolia,
        transport: http(this.config.parentChainRpcUrl)
      });

      logger.info('NodeManagementService clients initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize NodeManagementService clients', { error });
      throw new Error(`Failed to initialize NodeManagementService clients: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Generate node configuration from deployment transaction hash
   * Implements the missing step between Orbit deployment and node startup
   */
  async generateNodeConfig(deploymentTxHash: string, chainName?: string): Promise<NodeConfig> {
    try {
      logger.info('Generating node configuration', { deploymentTxHash, chainName });

      // Get the deployment transaction
      const tx = createRollupPrepareTransaction(
        await this.publicClient.getTransaction({ hash: deploymentTxHash as `0x${string}` })
      );

      // Get the transaction receipt
      const txReceipt = createRollupPrepareTransactionReceipt(
        await this.publicClient.getTransactionReceipt({ hash: deploymentTxHash as `0x${string}` })
      );

      // Extract chain configuration from transaction inputs
      const config = tx.getInputs()[0].config;
      const chainConfig: ChainConfig = JSON.parse(config.chainConfig);
      
      // Get core contracts from transaction receipt
      const coreContracts = txReceipt.getCoreContracts();

      logger.info('Extracted chain configuration', {
        chainId: chainConfig.chainId,
        hasCoreContracts: !!coreContracts
      });

      // Prepare node configuration parameters
      const nodeConfigParameters: PrepareNodeConfigParams = {
        chainName: chainName || 'My Orbit Chain',
        chainConfig,
        coreContracts,
        batchPosterPrivateKey: this.config.batchPosterPrivateKey as `0x${string}`,
        validatorPrivateKey: this.config.validatorPrivateKey as `0x${string}`,
        stakeToken: config.stakeToken,
        parentChainId: this.publicClient.chain.id,
        parentChainRpcUrl: this.config.parentChainRpcUrl,
      };

      // For L2 Orbit chains settling to Ethereum mainnet or testnet
      if (getParentChainLayer(this.publicClient.chain.id) === 1) {
        if (!this.config.parentChainBeaconRpcUrl) {
          throw new Error('ETHEREUM_BEACON_RPC_URL is required for L2 Orbit chains');
        }
        nodeConfigParameters.parentChainBeaconRpcUrl = this.config.parentChainBeaconRpcUrl;
      }

      // Generate the node configuration
      const nodeConfig = prepareNodeConfig(nodeConfigParameters);

      // Save node configuration to file
      await this.saveNodeConfig(nodeConfig, chainName || 'My Orbit Chain');

      logger.info('Node configuration generated successfully', {
        chainName: nodeConfigParameters.chainName,
        configSaved: true
      });

      return {
        chainConfig,
        coreContracts,
        chainName: nodeConfigParameters.chainName,
        batchPosterPrivateKey: this.config.batchPosterPrivateKey,
        validatorPrivateKey: this.config.validatorPrivateKey,
        stakeToken: config.stakeToken,
        parentChainId: this.publicClient.chain.id,
        parentChainRpcUrl: this.config.parentChainRpcUrl,
        parentChainBeaconRpcUrl: this.config.parentChainBeaconRpcUrl
      };

    } catch (error) {
      logger.error('Failed to generate node configuration', { 
        deploymentTxHash, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      throw new Error(`Failed to generate node configuration: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Start node using configuration
   * This would typically involve starting a Docker container or local process
   */
  async startNode(nodeConfig: NodeConfig): Promise<NodeStartupResult> {
    try {
      logger.info('Starting Orbit node', { chainName: nodeConfig.chainName });

      // Ensure node config directory exists
      await this.ensureNodeConfigDir();

      // Generate unique node identifier
      const nodeId = this.generateNodeId(nodeConfig.chainName);
      
      // Check if node is already running
      if (this.runningNodes.has(nodeId)) {
        const existingNode = this.runningNodes.get(nodeId)!;
        if (await this.isNodeProcessRunning(existingNode)) {
          logger.warn('Node is already running', { nodeId, chainName: nodeConfig.chainName });
          return {
            success: true,
            nodeConfigPath: existingNode.configPath,
            rpcUrl: existingNode.rpcUrl,
            explorerUrl: existingNode.explorerUrl
          };
        } else {
          // Clean up stale entry
          this.runningNodes.delete(nodeId);
        }
      }

      // Start the node process
      const nodeProcess = await this.startNodeProcess(nodeConfig, nodeId);

      // Wait for node to be ready (reduced timeout for simulation)
      const isReady = await this.waitForNodeReady(nodeProcess.rpcUrl, 10000); // 10 second timeout for simulation
      
      if (!isReady) {
        await this.stopNodeProcess(nodeProcess);
        throw new Error('Node failed to start within timeout period');
      }

      // Store running node info
      this.runningNodes.set(nodeId, nodeProcess);

      logger.info('Orbit node started successfully', {
        nodeId,
        chainName: nodeConfig.chainName,
        rpcUrl: nodeProcess.rpcUrl,
        explorerUrl: nodeProcess.explorerUrl
      });

      return {
        success: true,
        nodeConfigPath: nodeProcess.configPath,
        rpcUrl: nodeProcess.rpcUrl,
        explorerUrl: nodeProcess.explorerUrl
      };

    } catch (error) {
      logger.error('Failed to start node', { 
        chainName: nodeConfig.chainName,
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Node startup failed'
      };
    }
  }

  /**
   * Check if node is running and RPC is available
   */
  async isNodeReady(rpcUrl: string): Promise<boolean> {
    try {
      logger.debug('Checking node readiness', { rpcUrl });

      // For simulation mode, we'll assume the node is ready after a short delay
      // In a real implementation, this would test the actual RPC endpoint
      
      // Simulate node startup time (2-3 seconds)
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      logger.debug('Node is ready (simulated)', { rpcUrl });
      return true;

    } catch (error) {
      logger.debug('Node is not ready', { rpcUrl, error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  /**
   * Get detailed node status
   */
  async getNodeStatus(rpcUrl: string): Promise<NodeStatus> {
    try {
      // For simulation mode, return mock status
      // In a real implementation, this would query the actual RPC endpoint
      
      return {
        isRunning: true,
        rpcUrl,
        explorerUrl: this.generateExplorerUrl(rpcUrl),
        blockNumber: 1, // Mock block number
        chainId: 12345 // Mock chain ID
      };

    } catch (error) {
      return {
        isRunning: false,
        rpcUrl,
        error: error instanceof Error ? error.message : 'Status check failed'
      };
    }
  }

  /**
   * Stop a specific node
   */
  async stopNode(chainName: string): Promise<void> {
    try {
      const nodeId = this.generateNodeId(chainName);
      const nodeProcess = this.runningNodes.get(nodeId);

      if (!nodeProcess) {
        logger.warn('Node not found for stopping', { chainName, nodeId });
        return;
      }

      await this.stopNodeProcess(nodeProcess);
      this.runningNodes.delete(nodeId);

      logger.info('Node stopped successfully', { chainName, nodeId });

    } catch (error) {
      logger.error('Failed to stop node', { 
        chainName, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      throw new Error(`Failed to stop node: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Stop all running nodes
   */
  async stopAllNodes(): Promise<void> {
    try {
      logger.info('Stopping all running nodes', { nodeCount: this.runningNodes.size });

      const stopPromises = Array.from(this.runningNodes.values()).map(nodeProcess => 
        this.stopNodeProcess(nodeProcess).catch(error => {
          logger.error('Failed to stop individual node', { 
            chainName: nodeProcess.chainName, 
            error 
          });
        })
      );

      await Promise.all(stopPromises);
      this.runningNodes.clear();

      logger.info('All nodes stopped successfully');

    } catch (error) {
      logger.error('Failed to stop all nodes', { error });
      throw new Error(`Failed to stop all nodes: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get list of running nodes
   */
  getRunningNodes(): Array<{ chainName: string; rpcUrl: string; explorerUrl: string }> {
    return Array.from(this.runningNodes.values()).map(node => ({
      chainName: node.chainName,
      rpcUrl: node.rpcUrl,
      explorerUrl: node.explorerUrl
    }));
  }

  // Private helper methods

  private async saveNodeConfig(nodeConfig: any, chainName: string): Promise<string> {
    await this.ensureNodeConfigDir();
    
    const fileName = `node-config-${this.sanitizeFileName(chainName)}.json`;
    const filePath = join(this.config.nodeConfigDir, fileName);
    
    // Write file synchronously to avoid async file system operations that might trigger nodemon
    const fs = await import('fs');
    fs.writeFileSync(filePath, JSON.stringify(nodeConfig, null, 2));
    
    logger.info('Node configuration saved', { filePath, chainName });
    return filePath;
  }

  private async ensureNodeConfigDir(): Promise<void> {
    try {
      await access(this.config.nodeConfigDir, constants.F_OK);
    } catch {
      // Create directory synchronously to avoid async file system operations
      const fs = await import('fs');
      fs.mkdirSync(this.config.nodeConfigDir, { recursive: true });
      logger.info('Created node config directory', { dir: this.config.nodeConfigDir });
    }
  }

  private generateNodeId(chainName: string): string {
    return this.sanitizeFileName(chainName);
  }

  private sanitizeFileName(name: string): string {
    return name.replace(/[^a-zA-Z0-9-_]/g, '-').toLowerCase();
  }

  private generateRpcUrl(port: number): string {
    return `http://localhost:${port}`;
  }

  private generateExplorerUrl(rpcUrl: string): string {
    // Simple explorer URL generation - in production this would be more sophisticated
    return rpcUrl.replace('rpc', 'explorer').replace(':8449', ':8448');
  }

  private async startNodeProcess(nodeConfig: NodeConfig, nodeId: string): Promise<NodeProcess> {
    // In a real implementation, this would start the actual Orbit node
    // For now, we'll simulate the node startup process
    
    const rpcPort = this.config.defaultRpcPort + this.runningNodes.size;
    const explorerPort = this.config.defaultExplorerPort + this.runningNodes.size;
    
    const nodeProcess: NodeProcess = {
      id: nodeId,
      chainName: nodeConfig.chainName,
      configPath: join(this.config.nodeConfigDir, `node-config-${nodeId}.json`),
      rpcUrl: this.generateRpcUrl(rpcPort),
      explorerUrl: this.generateRpcUrl(explorerPort),
      pid: Date.now(), // Mock PID
      startTime: new Date()
    };

    // Simulate node startup delay
    await new Promise(resolve => setTimeout(resolve, 2000));

    logger.info('Node process started (simulated)', {
      nodeId,
      chainName: nodeConfig.chainName,
      rpcUrl: nodeProcess.rpcUrl,
      explorerUrl: nodeProcess.explorerUrl
    });

    return nodeProcess;
  }

  private async stopNodeProcess(nodeProcess: NodeProcess): Promise<void> {
    // In a real implementation, this would stop the actual process
    // For now, we'll simulate the shutdown
    
    logger.info('Stopping node process (simulated)', {
      nodeId: nodeProcess.id,
      chainName: nodeProcess.chainName
    });

    // Simulate shutdown delay
    await new Promise(resolve => setTimeout(resolve, 1000));
  }

  private async isNodeProcessRunning(nodeProcess: NodeProcess): Promise<boolean> {
    // In a real implementation, this would check if the process is actually running
    // For simulation mode, we'll assume the process is running if it's in our map
    return this.runningNodes.has(nodeProcess.id);
  }

  private async waitForNodeReady(rpcUrl: string, timeoutMs: number): Promise<boolean> {
    const startTime = Date.now();
    const checkInterval = 2000; // Check every 2 seconds

    while (Date.now() - startTime < timeoutMs) {
      if (await this.isNodeReady(rpcUrl)) {
        return true;
      }
      await new Promise(resolve => setTimeout(resolve, checkInterval));
    }

    return false;
  }
}

// Node process interface
interface NodeProcess {
  id: string;
  chainName: string;
  configPath: string;
  rpcUrl: string;
  explorerUrl: string;
  pid: number;
  startTime: Date;
}

export default NodeManagementService;
