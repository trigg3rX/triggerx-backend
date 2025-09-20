import { Router, Request, Response } from 'express';
import { HealthCheckResponse } from '../types/deployment';
import { config } from '../utils/config';
import { DockerUtils } from '../utils/dockerUtils';
import logger from '../utils/logger';

const router = Router();

// Health check endpoint
router.get('/health', async (req: Request, res: Response) => {
  try {
    // Check service connectivity
    const goBackendStatus = config.goBackendUrl ? 'connected' : 'disconnected';
    const arbitrumStatus = config.arbitrumRpcUrl ? 'connected' : 'disconnected';
    
    // Check Docker availability
    const isDockerAvailable = await DockerUtils.isDockerAvailable();
    const dockerVersion = isDockerAvailable ? await DockerUtils.getDockerVersion() : undefined;
    const dockerStatus = isDockerAvailable ? 'available' : 'unavailable';
    
    const overallStatus = (goBackendStatus === 'connected' && arbitrumStatus === 'connected' && dockerStatus === 'available') ? 'healthy' : 'unhealthy';
    
    const response: HealthCheckResponse = {
      status: overallStatus,
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      environment: config.nodeEnv,
      services: {
        goBackend: {
          status: goBackendStatus,
          url: config.goBackendUrl ? '***configured***' : undefined
        },
        arbitrum: {
          status: arbitrumStatus,
          rpc_url: config.arbitrumRpcUrl ? '***configured***' : undefined
        },
        docker: {
          status: dockerStatus,
          version: dockerVersion || undefined
        }
      },
      configuration: {
        port: config.port,
        deployment_timeout: config.deploymentTimeout,
        max_retries: config.maxRetries
      },
      uptime: Math.floor(process.uptime()),
      memory_usage: process.memoryUsage(),
      node_version: process.version,
      running_nodes: 0 // TODO: Get actual count from NodeManagementService
    };
    
    const statusCode = overallStatus === 'healthy' ? 200 : 503;
    res.status(statusCode).json(response);
    
  } catch (error) {
    logger.error('Health check failed', error);
    
    res.status(503).json({
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      environment: config.nodeEnv,
      error: error instanceof Error ? error.message : 'Unknown error'
    });
  }
});

export default router;
