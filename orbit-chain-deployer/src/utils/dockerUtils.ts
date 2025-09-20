import { spawn } from 'child_process';
import { promisify } from 'util';
import * as fs from 'fs/promises';
import * as path from 'path';
import logger from './logger';
import { DockerContainerInfo } from '../types/deployment';

const exec = require('child_process').exec;
const execAsync = promisify(exec);

export class DockerUtils {
  /**
   * Check if Docker is available and running
   */
  static async isDockerAvailable(): Promise<boolean> {
    try {
      const { stdout } = await execAsync('docker --version');
      logger.debug('Docker version check successful', { version: stdout.trim() });
      
      // Test Docker daemon connectivity
      await execAsync('docker info');
      logger.debug('Docker daemon is running');
      
      return true;
    } catch (error) {
      logger.warn('Docker is not available', { error: error instanceof Error ? error.message : 'Unknown error' });
      return false;
    }
  }

  /**
   * Get Docker version information
   */
  static async getDockerVersion(): Promise<string | null> {
    try {
      const { stdout } = await execAsync('docker --version');
      return stdout.trim();
    } catch (error) {
      logger.error('Failed to get Docker version', { error });
      return null;
    }
  }

  /**
   * Pull Docker image if not available locally
   */
  static async pullImage(imageName: string): Promise<boolean> {
    try {
      logger.info('Pulling Docker image', { imageName });
      
      const pullProcess = spawn('docker', ['pull', imageName], {
        stdio: ['ignore', 'pipe', 'pipe']
      });

      let stdout = '';
      let stderr = '';

      pullProcess.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      pullProcess.stderr?.on('data', (data) => {
        stderr += data.toString();
      });

      return new Promise((resolve, reject) => {
        pullProcess.on('close', (code) => {
          if (code === 0) {
            logger.info('Docker image pulled successfully', { imageName });
            resolve(true);
          } else {
            logger.error('Failed to pull Docker image', { 
              imageName, 
              code, 
              stderr: stderr.trim() 
            });
            reject(new Error(`Docker pull failed with code ${code}: ${stderr.trim()}`));
          }
        });

        pullProcess.on('error', (error) => {
          logger.error('Docker pull process error', { imageName, error });
          reject(error);
        });

        // Timeout after 10 minutes
        setTimeout(() => {
          pullProcess.kill();
          reject(new Error('Docker pull timeout'));
        }, 600000);
      });

    } catch (error) {
      logger.error('Docker image pull error', { imageName, error });
      throw error;
    }
  }

  /**
   * Check if Docker image exists locally
   */
  static async imageExists(imageName: string): Promise<boolean> {
    try {
      await execAsync(`docker image inspect ${imageName}`);
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Test RPC connectivity before starting container
   */
  static async testRpcConnectivity(rpcUrl: string): Promise<boolean> {
    try {
      logger.info('Testing RPC connectivity', { rpcUrl });
      
      const response = await fetch(rpcUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'eth_chainId',
          params: [],
          id: 1
        }),
        signal: AbortSignal.timeout(10000) // 10 second timeout
      });
      
      if (!response.ok) {
        logger.warn('RPC connectivity test failed', { rpcUrl, status: response.status });
        return false;
      }
      
      const result = await response.json() as any;
      const isConnected = !result.error && result.result;
      
      logger.info('RPC connectivity test result', { 
        rpcUrl, 
        isConnected,
        chainId: result.result 
      });
      
      return isConnected;
    } catch (error) {
      logger.warn('RPC connectivity test error', { 
        rpcUrl, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      return false;
    }
  }

  /**
   * Test DNS resolution from within a running container
   */
  static async testContainerDnsResolution(containerName: string, hostname: string): Promise<boolean> {
    try {
      logger.info('Testing DNS resolution in container', { containerName, hostname });
      
      const { stdout, stderr } = await execAsync(
        `docker exec ${containerName} nslookup ${hostname}`
      );
      
      const isResolved = stdout.includes('Address:') && !stderr.includes('NXDOMAIN');
      
      logger.info('DNS resolution test result', { 
        containerName, 
        hostname, 
        isResolved,
        output: stdout 
      });
      
      return isResolved;
    } catch (error) {
      logger.warn('DNS resolution test error', { 
        containerName, 
        hostname,
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      return false;
    }
  }

  /**
   * Start Docker container for Orbit node
   */
  static async startContainer(
    containerName: string,
    imageName: string,
    configPath: string,
    dataDir: string,
    logsDir: string,
    rpcPort: number,
    explorerPort: number,
    metricsPort: number,
    memoryLimit: string,
    cpuLimit: string
  ): Promise<void> {
    try {
      logger.info('Starting Docker container', { 
        containerName, 
        imageName, 
        rpcPort, 
        explorerPort 
      });

      // Ensure directories exist
      await fs.mkdir(dataDir, { recursive: true });
      await fs.mkdir(logsDir, { recursive: true });

      // Convert all paths to absolute paths to avoid Docker volume name issues
      const absoluteConfigPath = path.resolve(configPath);
      const absoluteDataDir = path.resolve(dataDir);
      const absoluteLogsDir = path.resolve(logsDir);

      const dockerArgs = [
        'run',
        '--detach',
        '--name', containerName,
        '--restart', 'unless-stopped',
        '--memory', memoryLimit,
        '--cpus', cpuLimit,
        '-p', `${rpcPort}:8449`,      // RPC port
        '-p', `${explorerPort}:8448`, // Explorer port  
        '-p', `${metricsPort}:9090`,  // Metrics port
        '-v', `${absoluteConfigPath}:/config/node-config.json:ro`,
        '-v', `${absoluteDataDir}:/home/user/.arbitrum/nitro`,
        '-v', `${absoluteLogsDir}:/home/user/logs`,
        '--dns', '8.8.8.8',          // Use Google DNS for better resolution
        '--dns', '8.8.4.4',          // Secondary DNS
        '--dns', '1.1.1.1',          // Cloudflare DNS as backup
        '--add-host', 'host.docker.internal:host-gateway', // Add host access
        imageName,
        '--conf.file', '/config/node-config.json'
      ];

      const { stdout, stderr } = await execAsync(`docker ${dockerArgs.join(' ')}`);
      
      if (stderr && !stderr.includes('WARNING')) {
        throw new Error(`Docker container start failed: ${stderr}`);
      }

      logger.info('Docker container started successfully', { 
        containerName, 
        containerId: stdout.trim() 
      });

    } catch (error) {
      logger.error('Failed to start Docker container', { 
        containerName, 
        imageName, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      throw error;
    }
  }

  /**
   * Stop Docker container
   */
  static async stopContainer(containerName: string): Promise<void> {
    try {
      logger.info('Stopping Docker container', { containerName });

      // Stop container
      await execAsync(`docker stop ${containerName} || true`);
      
      // Remove container
      await execAsync(`docker rm ${containerName} || true`);

      logger.info('Docker container stopped and removed', { containerName });

    } catch (error) {
      logger.error('Failed to stop Docker container', { 
        containerName, 
        error: error instanceof Error ? error.message : 'Unknown error' 
      });
      throw error;
    }
  }

  /**
   * Check if Docker container is running
   */
  static async isContainerRunning(containerName: string): Promise<boolean> {
    try {
      const { stdout } = await execAsync(`docker ps --filter name=${containerName} --format "{{.Names}}"`);
      return stdout.trim() === containerName;
    } catch (error) {
      logger.debug('Failed to check container status', { containerName, error });
      return false;
    }
  }

  /**
   * Get container PID
   */
  static async getContainerPid(containerName: string): Promise<number | null> {
    try {
      const { stdout } = await execAsync(`docker inspect -f '{{.State.Pid}}' ${containerName}`);
      const pid = parseInt(stdout.trim());
      return isNaN(pid) ? null : pid;
    } catch (error) {
      logger.debug('Failed to get container PID', { containerName, error });
      return null;
    }
  }

  /**
   * Get container information
   */
  static async getContainerInfo(containerName: string): Promise<DockerContainerInfo | null> {
    try {
      const { stdout } = await execAsync(`docker inspect ${containerName}`);
      const containerInfo = JSON.parse(stdout)[0];
      
      return {
        name: containerInfo.Name.replace('/', ''),
        id: containerInfo.Id,
        status: containerInfo.State.Status,
        ports: Object.keys(containerInfo.NetworkSettings.Ports || {}),
        created: containerInfo.Created,
        image: containerInfo.Config.Image
      };
    } catch (error) {
      logger.debug('Failed to get container info', { containerName, error });
      return null;
    }
  }

  /**
   * Get all running containers with specific prefix
   */
  static async getRunningContainers(prefix: string): Promise<DockerContainerInfo[]> {
    try {
      const { stdout } = await execAsync(`docker ps --filter name=${prefix} --format "{{.Names}}"`);
      const containerNames = stdout.trim().split('\n').filter((name: string) => name.length > 0);
      
      const containers: DockerContainerInfo[] = [];
      for (const name of containerNames) {
        const info = await this.getContainerInfo(name);
        if (info) {
          containers.push(info);
        }
      }
      
      return containers;
    } catch (error) {
      logger.error('Failed to get running containers', { prefix, error });
      return [];
    }
  }

  /**
   * Get container logs
   */
  static async getContainerLogs(containerName: string, tail: number = 100): Promise<string> {
    try {
      const { stdout } = await execAsync(`docker logs --tail ${tail} ${containerName}`);
      return stdout;
    } catch (error) {
      logger.error('Failed to get container logs', { containerName, error });
      return '';
    }
  }

  /**
   * Clean up stopped containers with specific prefix
   */
  static async cleanupStoppedContainers(prefix: string): Promise<void> {
    try {
      logger.info('Cleaning up stopped containers', { prefix });
      
      const { stdout } = await execAsync(`docker ps -a --filter name=${prefix} --filter status=exited --format "{{.Names}}"`);
      const containerNames = stdout.trim().split('\n').filter((name: string) => name.length > 0);
      
      for (const name of containerNames) {
        await execAsync(`docker rm ${name} || true`);
        logger.debug('Removed stopped container', { name });
      }
      
      logger.info('Cleanup completed', { removedContainers: containerNames.length });
    } catch (error) {
      logger.error('Failed to cleanup stopped containers', { prefix, error });
    }
  }

  /**
   * Get available port in range
   */
  static async getAvailablePort(startPort: number, endPort: number): Promise<number> {
    try {
      // Check which ports are in use by Docker containers
      const { stdout } = await execAsync('docker ps --format "{{.Ports}}"');
      const usedPorts = new Set<number>();
      
      // Parse used ports from Docker output
      const portMatches = stdout.match(/(\d+):/g);
      if (portMatches) {
        for (const match of portMatches) {
          const port = parseInt(match.replace(':', ''));
          usedPorts.add(port);
        }
      }
      
      // Find first available port
      for (let port = startPort; port <= endPort; port++) {
        if (!usedPorts.has(port)) {
          return port;
        }
      }
      
      throw new Error(`No available ports in range ${startPort}-${endPort}`);
    } catch (error) {
      logger.error('Failed to get available port', { startPort, endPort, error });
      throw error;
    }
  }
}

export default DockerUtils;
