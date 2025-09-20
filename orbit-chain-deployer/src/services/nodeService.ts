import { createPublicClient, http } from 'viem';
import { arbitrumSepolia } from 'viem/chains';
import { 
  prepareNodeConfig, 
  createRollupPrepareTransaction,
  createRollupPrepareTransactionReceipt,
  type ChainConfig,
  type PrepareNodeConfigParams,
  type NodeConfig as OrbitNodeConfig
} from '@arbitrum/orbit-sdk';
import { getParentChainLayer } from '@arbitrum/orbit-sdk/utils';
import { mkdir, access } from 'fs/promises';
import { constants } from 'fs';
import { join } from 'path';
import logger from '../utils/logger';
import { config } from '../utils/config';
import { NodeProcess, NodeStartupResult, NodeStatus, NodeManagementConfig } from '../types/deployment';
import { DockerUtils } from '../utils/dockerUtils';

// Use Orbit SDK's NodeConfig type directly

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
  async generateNodeConfig(deploymentTxHash: string, chainName?: string): Promise<OrbitNodeConfig> {
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

      // Return the actual Orbit SDK's NodeConfig for use by startNode
      return nodeConfig;

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
  async startNode(nodeConfig: OrbitNodeConfig): Promise<NodeStartupResult> {
    try {
      logger.info('Starting Orbit node', { chainName: nodeConfig.chain?.name });

      // Ensure node config directory exists
      await this.ensureNodeConfigDir();

      // Generate unique node identifier
      const nodeId = this.generateNodeId(nodeConfig.chain?.name || 'unknown');
      
      // Check if node is already running
      if (this.runningNodes.has(nodeId)) {
        const existingNode = this.runningNodes.get(nodeId)!;
        if (await this.isNodeProcessRunning(existingNode)) {
          logger.warn('Node is already running', { nodeId, chainName: nodeConfig.chain?.name });
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

      // Store running node info immediately (before health check)
      this.runningNodes.set(nodeId, nodeProcess);

      logger.info('Orbit node container started, beginning health check', {
        nodeId,
        chainName: nodeConfig.chain?.name || 'unknown',
        rpcUrl: nodeProcess.rpcUrl,
        explorerUrl: nodeProcess.explorerUrl
      });

      // Return immediately with the node info - health check will happen in background
      return {
        success: true,
        nodeConfigPath: nodeProcess.configPath,
        rpcUrl: nodeProcess.rpcUrl,
        explorerUrl: nodeProcess.explorerUrl
      };

    } catch (error) {
      logger.error('Failed to start node', { 
        chainName: nodeConfig.chain?.name || 'unknown',
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Node startup failed'
      };
    }
  }

  /**
   * Wait for node to be ready (for use after async startup)
   */
  async waitForNodeReady(rpcUrl: string, timeoutMs?: number): Promise<boolean> {
    const startupTimeout = timeoutMs || this.config.startupTimeout || config.node.startupTimeout;
    const isReady = await this.waitForNodeReadyInternal(rpcUrl, startupTimeout);
    
    if (isReady) {
      logger.info('Node is ready and responding', { rpcUrl });
    } else {
      logger.warn('Node failed to become ready within timeout', { rpcUrl, timeoutMs: startupTimeout });
    }
    
    return isReady;
  }

  /**
   * Check if node is running and RPC is available
   */
  async isNodeReady(rpcUrl: string): Promise<boolean> {
    try {
      logger.debug('Checking node readiness', { rpcUrl });

      // First check if the container is running
      const runningNodes = Array.from(this.runningNodes.values());
      const nodeProcess = runningNodes.find(node => node.rpcUrl === rpcUrl);
      
      if (nodeProcess && nodeProcess.containerName) {
        const isContainerRunning = await DockerUtils.isContainerRunning(nodeProcess.containerName);
        if (!isContainerRunning) {
          logger.debug('Container is not running', { rpcUrl, containerName: nodeProcess.containerName });
          return false;
        }
      }

      // Try multiple health check methods for Nitro nodes
      const healthCheckMethods = [
        { name: 'healthEndpoint', method: () => this.checkHealthEndpoint(rpcUrl) },
        { name: 'rpcEndpoint', method: () => this.checkRpcEndpoint(rpcUrl) },
        { name: 'chainIdEndpoint', method: () => this.checkChainIdEndpoint(rpcUrl) }
      ];

      // Try each health check method
      for (const { name, method } of healthCheckMethods) {
        try {
          const isReady = await method();
          if (isReady) {
            logger.debug('Node is ready', { rpcUrl, method: name });
            return true;
          }
        } catch (error) {
          logger.debug('Health check method failed', { 
            rpcUrl, 
            method: name,
            error: error instanceof Error ? error.message : 'Unknown error' 
          });
        }
      }

      logger.debug('All health check methods failed', { rpcUrl });
      return false;

    } catch (error) {
      logger.debug('Node is not ready', { rpcUrl, error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  private async checkHealthEndpoint(rpcUrl: string): Promise<boolean> {
    // Nitro nodes don't expose a standard /health endpoint
    // Use net_version as a lightweight health check instead
    try {
      const response = await fetch(rpcUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'net_version',
          params: [],
          id: 1
        }),
        signal: AbortSignal.timeout(5000)
      });
      
      if (!response.ok) {
        logger.debug('Health endpoint not available', { rpcUrl, status: response.status });
        return false;
      }
      
      const result = await response.json() as any;
      const isHealthy = !result.error && result.result;
      logger.debug('Health endpoint check result', { rpcUrl, isHealthy, result });
      return isHealthy;
    } catch (error) {
      logger.debug('Health endpoint check failed', { rpcUrl, error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  private async checkRpcEndpoint(rpcUrl: string): Promise<boolean> {
    try {
      const rpcResponse = await fetch(rpcUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_blockNumber',
          params: [],
          id: 1
        }),
        signal: AbortSignal.timeout(5000) // Increased to 5 second timeout
      });
      
      if (!rpcResponse.ok) {
        logger.debug('RPC endpoint not available', { rpcUrl, status: rpcResponse.status });
        return false;
      }
      
      const result = await rpcResponse.json() as any;
      const isReady = !result.error && result.result;
      logger.debug('RPC endpoint check result', { rpcUrl, isReady, result });
      return isReady;
    } catch (error) {
      logger.debug('RPC endpoint check failed', { rpcUrl, error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  private async checkChainIdEndpoint(rpcUrl: string): Promise<boolean> {
    try {
      const rpcResponse = await fetch(rpcUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_chainId',
          params: [],
          id: 2
        }),
        signal: AbortSignal.timeout(5000) // Increased to 5 second timeout
      });
      
      if (!rpcResponse.ok) {
        logger.debug('Chain ID endpoint not available', { rpcUrl, status: rpcResponse.status });
        return false;
      }
      
      const result = await rpcResponse.json() as any;
      const isReady = !result.error && result.result;
      logger.debug('Chain ID endpoint check result', { rpcUrl, isReady, result });
      return isReady;
    } catch (error) {
      logger.debug('Chain ID endpoint check failed', { rpcUrl, error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  /**
   * Get detailed node status
   */
  async getNodeStatus(rpcUrl: string): Promise<NodeStatus> {
    try {
      // Get real node status via RPC calls
      const isRunning = await this.isNodeReady(rpcUrl);
      
      if (!isRunning) {
        return {
          isRunning: false,
          rpcUrl,
          error: 'Node is not responding'
        };
      }

      // Get block number
      let blockNumber: number | undefined;
      let chainId: number | undefined;
      
      try {
        const blockResponse = await fetch(rpcUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_blockNumber',
            params: [],
            id: 1
          }),
          signal: AbortSignal.timeout(5000)
        });
        
        if (blockResponse.ok) {
          const blockResult = await blockResponse.json() as any;
          if (!blockResult.error) {
            blockNumber = parseInt(blockResult.result, 16);
          }
        }
      } catch (blockError) {
        logger.debug('Failed to get block number', { rpcUrl, error: blockError });
      }

      // Get chain ID
      try {
        const chainResponse = await fetch(rpcUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_chainId',
            params: [],
            id: 2
          }),
          signal: AbortSignal.timeout(5000)
        });
        
        if (chainResponse.ok) {
          const chainResult = await chainResponse.json() as any;
          if (!chainResult.error) {
            chainId = parseInt(chainResult.result, 16);
          }
        }
      } catch (chainError) {
        logger.debug('Failed to get chain ID', { rpcUrl, error: chainError });
      }
      
      return {
        isRunning: true,
        rpcUrl,
        explorerUrl: this.generateExplorerUrl(rpcUrl),
        blockNumber,
        chainId
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

  private async saveNodeConfig(nodeConfig: OrbitNodeConfig, chainName: string): Promise<string> {
    await this.ensureNodeConfigDir();
    
    const fileName = `node-config-${this.sanitizeFileName(chainName)}.json`;
    const filePath = join(this.config.nodeConfigDir, fileName);
    
    // The Orbit SDK's prepareNodeConfig already returns a Nitro-compatible config
    // No conversion needed - use the config directly
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

  private async startNodeProcess(nodeConfig: OrbitNodeConfig, nodeId: string): Promise<NodeProcess> {
    try {
      // Check if Docker is available
      const isDockerAvailable = await DockerUtils.isDockerAvailable();
      if (!isDockerAvailable) {
        throw new Error('Docker is not available. Please install Docker and ensure it is running.');
      }

      // 1. Save node configuration to file
      const configPath = await this.saveNodeConfig(nodeConfig, nodeConfig.chain?.name || 'unknown');
      
      // 2. Generate unique container name
      const containerName = `${this.config.containerPrefix || 'orbit-node-'}${nodeId}`;
      
      // 3. Create Docker volume mount points
      const dataDir = join(this.config.nodeConfigDir, `${nodeId}-data`);
      const logsDir = join(this.config.nodeConfigDir, `${nodeId}-logs`);
      
      await mkdir(dataDir, { recursive: true });
      await mkdir(logsDir, { recursive: true });
      
      // 4. Get available ports
      const rpcPort = await DockerUtils.getAvailablePort(
        this.config.defaultRpcPort, 
        this.config.defaultRpcPort + 100
      );
      const explorerPort = await DockerUtils.getAvailablePort(
        this.config.defaultExplorerPort, 
        this.config.defaultExplorerPort + 100
      );
      const metricsPort = await DockerUtils.getAvailablePort(9090, 9190);
      
      // 5. Ensure Docker image is available
      const dockerImage = this.config.dockerImage || 'offchainlabs/nitro-node:v3.6.7-a7c9f1e';
      const imageExists = await DockerUtils.imageExists(dockerImage);
      
      if (!imageExists) {
        logger.info('Docker image not found locally, pulling...', { dockerImage });
        await DockerUtils.pullImage(dockerImage);
      }
      
      // 5.5. Test RPC connectivity before starting container
      const parentChainRpcUrl = nodeConfig['parent-chain']?.connection?.url;
      if (parentChainRpcUrl) {
        const isRpcConnected = await DockerUtils.testRpcConnectivity(parentChainRpcUrl);
        if (!isRpcConnected) {
          logger.warn('RPC connectivity test failed, but proceeding with container start', { 
            parentChainRpcUrl 
          });
        }
      }
      
      // 6. Start Docker container
      await DockerUtils.startContainer(
        containerName,
        dockerImage,
        configPath,
        dataDir,
        logsDir,
        rpcPort,
        explorerPort,
        metricsPort,
        this.config.memoryLimit || '2g',
        this.config.cpuLimit || '1'
      );
      
      // 7. Test DNS resolution in the container
      if (parentChainRpcUrl) {
        try {
          const url = new URL(parentChainRpcUrl);
          const hostname = url.hostname;
          
          // Wait a moment for container to fully start
          await new Promise(resolve => setTimeout(resolve, 2000));
          
          const canResolveDns = await DockerUtils.testContainerDnsResolution(containerName, hostname);
          if (!canResolveDns) {
            logger.warn('DNS resolution failed in container', { 
              containerName, 
              hostname,
              parentChainRpcUrl 
            });
          }
        } catch (error) {
          logger.warn('Failed to test DNS resolution', { 
            containerName, 
            parentChainRpcUrl,
            error: error instanceof Error ? error.message : 'Unknown error' 
          });
        }
      }
      
      // 8. Get container PID
      const pid = await DockerUtils.getContainerPid(containerName);
      
      const nodeProcess: NodeProcess = {
        id: nodeId,
        chainName: nodeConfig.chain?.name || 'unknown',
        configPath,
        rpcUrl: this.generateRpcUrl(rpcPort),
        explorerUrl: this.generateExplorerUrl(this.generateRpcUrl(rpcPort)),
        pid: pid || Date.now(), // Fallback to timestamp if PID not available
        startTime: new Date(),
        containerName,
        dataDir,
        logsDir
      };
      
      logger.info('Real Orbit node started via Docker', {
        nodeId,
        containerName,
        rpcUrl: nodeProcess.rpcUrl,
        explorerUrl: nodeProcess.explorerUrl,
        pid: nodeProcess.pid
      });
      
      return nodeProcess;
      
    } catch (error) {
      logger.error('Failed to start real node process', { nodeId, error });
      throw new Error(`Failed to start real node process: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  private async stopNodeProcess(nodeProcess: NodeProcess): Promise<void> {
    try {
      if (nodeProcess.containerName) {
        // Stop Docker container
        await DockerUtils.stopContainer(nodeProcess.containerName);
        
        logger.info('Real Orbit node stopped via Docker', {
          nodeId: nodeProcess.id,
          containerName: nodeProcess.containerName
        });
      } else if (nodeProcess.process) {
        // Stop binary process
        nodeProcess.process.kill('SIGTERM');
        
        // Wait for graceful shutdown
        await new Promise((resolve) => {
          nodeProcess.process!.on('exit', resolve);
          setTimeout(resolve, 5000); // Force exit after 5 seconds
        });
        
        logger.info('Real Orbit node stopped via binary process', {
          nodeId: nodeProcess.id,
          pid: nodeProcess.pid
        });
      } else {
        logger.warn('No container or process found to stop', { nodeId: nodeProcess.id });
      }
    } catch (error) {
      logger.error('Failed to stop real node process', { 
        nodeId: nodeProcess.id, 
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      throw error;
    }
  }

  private async isNodeProcessRunning(nodeProcess: NodeProcess): Promise<boolean> {
    try {
      if (nodeProcess.containerName) {
        // Check if Docker container is running
        return await DockerUtils.isContainerRunning(nodeProcess.containerName);
      } else if (nodeProcess.process) {
        // Check if binary process is running
        return !nodeProcess.process.killed && nodeProcess.process.exitCode === null;
      } else {
        // Fallback to map check
        return this.runningNodes.has(nodeProcess.id);
      }
    } catch (error) {
      logger.debug('Failed to check node process status', { 
        nodeId: nodeProcess.id, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      return false;
    }
  }

  private async waitForNodeReadyInternal(rpcUrl: string, timeoutMs: number): Promise<boolean> {
    const startTime = Date.now();
    const checkInterval = 3000; // Check every 3 seconds (increased from 2)
    let attemptCount = 0;
    let lastSuccessfulCheck = 0;

    logger.info('Starting node readiness check', { rpcUrl, timeoutMs });

    while (Date.now() - startTime < timeoutMs) {
      attemptCount++;
      const elapsed = Date.now() - startTime;

      // Try multiple health check methods
      const isReady = await this.isNodeReady(rpcUrl);
      
      if (isReady) {
        logger.info('Node readiness check successful', { 
          rpcUrl, 
          attempt: attemptCount, 
          elapsed: `${Math.round(elapsed / 1000)}s` 
        });
        return true;
      }

      // Track if we've had any successful RPC responses
      try {
        const testResponse = await fetch(rpcUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            jsonrpc: '2.0',
            method: 'eth_chainId',
            params: [],
            id: 1
          }),
          signal: AbortSignal.timeout(5000)
        });
        
        if (testResponse.ok) {
          const result = await testResponse.json() as any;
          if (!result.error && result.result) {
            lastSuccessfulCheck = attemptCount;
            logger.debug('RPC responding but node not fully ready', { 
              rpcUrl, 
              attempt: attemptCount,
              chainId: result.result 
            });
          }
        }
      } catch (error) {
        // RPC not responding yet, continue waiting
      }
      await new Promise(resolve => setTimeout(resolve, checkInterval));
    }

    logger.warn('Node readiness check timed out', { 
      rpcUrl, 
      attempt: attemptCount, 
      elapsed: `${Math.round((Date.now() - startTime) / 1000)}s`,
      lastSuccessfulRpc: lastSuccessfulCheck > 0 ? `attempt ${lastSuccessfulCheck}` : 'none'
    });
    return false;
  }
}

export default NodeManagementService;
