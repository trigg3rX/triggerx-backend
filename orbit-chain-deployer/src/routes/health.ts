import { Router, Request, Response } from 'express';
import { HealthCheckResponse } from '../types/deployment';
import { config } from '../utils/config';
import logger from '../utils/logger';

const router = Router();

// Health check endpoint
router.get('/health', async (req: Request, res: Response) => {
  try {
    // Check service connectivity (simplified)
    const databaseStatus = config.dbServerUrl ? 'connected' : 'disconnected';
    const arbitrumStatus = config.arbitrumRpcUrl ? 'connected' : 'disconnected';
    const overallStatus = (databaseStatus === 'connected' && arbitrumStatus === 'connected') ? 'healthy' : 'unhealthy';
    
    const response: HealthCheckResponse = {
      status: overallStatus,
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      environment: config.nodeEnv,
      services: {
        database: {
          status: databaseStatus,
          url: config.dbServerUrl ? '***configured***' : undefined
        },
        arbitrum: {
          status: arbitrumStatus,
          rpc_url: config.arbitrumRpcUrl ? '***configured***' : undefined
        }
      },
      configuration: {
        port: config.port,
        log_level: config.logLevel,
        deployment_timeout: config.deploymentTimeout,
        max_retries: config.maxRetries
      },
      uptime: Math.floor(process.uptime()),
      memory_usage: process.memoryUsage(),
      node_version: process.version
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
