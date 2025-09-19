import axios from 'axios';
import logger from '../utils/logger';
import { DeploymentStatus, ChainDeploymentStatus, ContractInfo } from '../types/deployment';

export interface StatusServiceConfig {
  goBackendUrl: string;
  apiKey: string;
}

export interface DeploymentProgress {
  deploymentId: string;
  status: DeploymentStatus;
  progress: number; // 0-100
  currentStep: string;
  chainAddress?: string;
  contracts?: ContractInfo[];
  error?: string;
  logs: string[];
  startedAt: Date;
  updatedAt: Date;
}

class StatusService {
  private config: StatusServiceConfig;
  private activeDeployments: Map<string, DeploymentProgress> = new Map();

  constructor(config: StatusServiceConfig) {
    this.config = config;
  }

  /**
   * Update deployment status in the Go backend database
   * Implements FR-009, FR-010: Track deployment progress and provide updates
   */
  async updateDeploymentStatus(
    deploymentId: string,
    status: DeploymentStatus,
    additionalData?: {
      chainAddress?: string;
      errorMessage?: string;
      deploymentLogs?: string;
      contracts?: ContractInfo[];
    }
  ): Promise<boolean> {
    try {
      logger.info('Updating deployment status', { deploymentId, status });

      const updateData = {
        deployment_id: deploymentId,
        status: status,
        orbit_chain_address: additionalData?.chainAddress,
        error_message: additionalData?.errorMessage,
        deployment_logs: additionalData?.deploymentLogs,
        contracts: additionalData?.contracts
      };

      const response = await axios.put(
        `${this.config.goBackendUrl}/api/orbit-chain/update-status`,
        updateData,
        {
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.config.apiKey}`
          },
          timeout: 10000
        }
      );

      if (response.status === 200) {
        logger.info('Deployment status updated successfully', { deploymentId, status });
        return true;
      } else {
        logger.error('Failed to update deployment status', { 
          deploymentId, 
          status, 
          responseStatus: response.status 
        });
        return false;
      }

    } catch (error) {
      logger.error('Error updating deployment status', { 
        deploymentId, 
        status, 
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      return false;
    }
  }

  /**
   * Get deployment status from the Go backend
   */
  async getDeploymentStatus(deploymentId: string): Promise<ChainDeploymentStatus | null> {
    try {
      logger.info('Getting deployment status', { deploymentId });

      const response = await axios.get(
        `${this.config.goBackendUrl}/api/orbit-chains/${deploymentId}/status`,
        {
          headers: {
            'Authorization': `Bearer ${this.config.apiKey}`
          },
          timeout: 10000
        }
      );

      if (response.status === 200) {
        return response.data;
      } else {
        logger.error('Failed to get deployment status', { 
          deploymentId, 
          responseStatus: response.status 
        });
        return null;
      }

    } catch (error) {
      logger.error('Error getting deployment status', { 
        deploymentId, 
        error: error instanceof Error ? error.message : 'Unknown error'
      });
      return null;
    }
  }

  /**
   * Start tracking a new deployment
   */
  startDeploymentTracking(deploymentId: string, chainName: string): DeploymentProgress {
    const progress: DeploymentProgress = {
      deploymentId,
      status: DeploymentStatus.PENDING,
      progress: 0,
      currentStep: 'Initializing deployment',
      logs: [`Deployment started for chain: ${chainName}`],
      startedAt: new Date(),
      updatedAt: new Date()
    };

    this.activeDeployments.set(deploymentId, progress);
    logger.info('Started deployment tracking', { deploymentId, chainName });

    return progress;
  }

  /**
   * Update local deployment progress
   */
  updateDeploymentProgress(
    deploymentId: string,
    updates: {
      status?: DeploymentStatus;
      progress?: number;
      currentStep?: string;
      chainAddress?: string;
      contracts?: ContractInfo[];
      error?: string;
      log?: string;
    }
  ): DeploymentProgress | null {
    const progress = this.activeDeployments.get(deploymentId);
    if (!progress) {
      logger.warn('Deployment progress not found', { deploymentId });
      return null;
    }

    // Update progress fields
    if (updates.status !== undefined) progress.status = updates.status;
    if (updates.progress !== undefined) progress.progress = updates.progress;
    if (updates.currentStep !== undefined) progress.currentStep = updates.currentStep;
    if (updates.chainAddress !== undefined) progress.chainAddress = updates.chainAddress;
    if (updates.contracts !== undefined) progress.contracts = updates.contracts;
    if (updates.error !== undefined) progress.error = updates.error;
    if (updates.log !== undefined) progress.logs.push(updates.log);

    progress.updatedAt = new Date();

    this.activeDeployments.set(deploymentId, progress);

    logger.info('Updated deployment progress', { 
      deploymentId, 
      status: progress.status,
      progress: progress.progress,
      currentStep: progress.currentStep
    });

    return progress;
  }

  /**
   * Get local deployment progress
   */
  getDeploymentProgress(deploymentId: string): DeploymentProgress | null {
    return this.activeDeployments.get(deploymentId) || null;
  }

  /**
   * Complete deployment tracking
   */
  completeDeploymentTracking(deploymentId: string, success: boolean, error?: string): void {
    const progress = this.activeDeployments.get(deploymentId);
    if (!progress) {
      logger.warn('Deployment progress not found for completion', { deploymentId });
      return;
    }

    progress.status = success ? DeploymentStatus.COMPLETED : DeploymentStatus.FAILED;
    progress.progress = 100;
    progress.currentStep = success ? 'Deployment completed successfully' : 'Deployment failed';
    progress.updatedAt = new Date();

    if (error) {
      progress.error = error;
      progress.logs.push(`Error: ${error}`);
    } else {
      progress.logs.push('Deployment completed successfully');
    }

    this.activeDeployments.set(deploymentId, progress);

    logger.info('Completed deployment tracking', { 
      deploymentId, 
      success, 
      duration: progress.updatedAt.getTime() - progress.startedAt.getTime()
    });
  }

  /**
   * Get all active deployments
   */
  getActiveDeployments(): DeploymentProgress[] {
    return Array.from(this.activeDeployments.values());
  }

  /**
   * Clean up completed deployments (older than 1 hour)
   */
  cleanupCompletedDeployments(): void {
    const oneHourAgo = new Date(Date.now() - 60 * 60 * 1000);
    const toDelete: string[] = [];

    for (const [deploymentId, progress] of this.activeDeployments.entries()) {
      if (
        (progress.status === DeploymentStatus.COMPLETED || 
         progress.status === DeploymentStatus.FAILED) &&
        progress.updatedAt < oneHourAgo
      ) {
        toDelete.push(deploymentId);
      }
    }

    for (const deploymentId of toDelete) {
      this.activeDeployments.delete(deploymentId);
      logger.info('Cleaned up completed deployment', { deploymentId });
    }

    if (toDelete.length > 0) {
      logger.info('Cleaned up completed deployments', { count: toDelete.length });
    }
  }

  /**
   * Get deployment statistics
   */
  getDeploymentStatistics(): {
    total: number;
    active: number;
    completed: number;
    failed: number;
    averageDuration: number;
  } {
    const deployments = Array.from(this.activeDeployments.values());
    const total = deployments.length;
    const active = deployments.filter(d => 
      d.status === DeploymentStatus.PENDING || 
      d.status === DeploymentStatus.DEPLOYING_ORBIT ||
      d.status === DeploymentStatus.ORBIT_DEPLOYED ||
      d.status === DeploymentStatus.DEPLOYING_CONTRACTS ||
      d.status === DeploymentStatus.CONFIGURING_CONTRACTS
    ).length;
    const completed = deployments.filter(d => d.status === DeploymentStatus.COMPLETED).length;
    const failed = deployments.filter(d => d.status === DeploymentStatus.FAILED).length;

    const completedDeployments = deployments.filter(d => d.status === DeploymentStatus.COMPLETED);
    const averageDuration = completedDeployments.length > 0
      ? completedDeployments.reduce((sum, d) => 
          sum + (d.updatedAt.getTime() - d.startedAt.getTime()), 0
        ) / completedDeployments.length
      : 0;

    return {
      total,
      active,
      completed,
      failed,
      averageDuration
    };
  }

  /**
   * Health check for status service
   */
  async healthCheck(): Promise<{
    status: 'healthy' | 'unhealthy';
    goBackendConnected: boolean;
    activeDeployments: number;
    error?: string;
  }> {
    try {
      // Test connection to Go backend
      const response = await axios.get(
        `${this.config.goBackendUrl}/health`,
        { timeout: 5000 }
      );

      const goBackendConnected = response.status === 200;
      const activeDeployments = this.getActiveDeployments().length;

      return {
        status: goBackendConnected ? 'healthy' : 'unhealthy',
        goBackendConnected,
        activeDeployments
      };

    } catch (error) {
      logger.error('Status service health check failed', { error });
      return {
        status: 'unhealthy',
        goBackendConnected: false,
        activeDeployments: this.getActiveDeployments().length,
        error: error instanceof Error ? error.message : 'Unknown error'
      };
    }
  }
}

export default StatusService;
