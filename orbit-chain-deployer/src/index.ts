import express from 'express';
import cors from 'cors';
import { config, isDevelopment } from './utils/config';
import logger from './utils/logger';
import routes from './routes';

const app = express();
const PORT = config.port;

// Middleware
app.use(cors());
app.use(express.json({ limit: '10mb' }));

// Request logging middleware
app.use((req, res, next) => {
  const requestId = req.headers['x-request-id'] as string || 'unknown';
  
  logger.http(`${req.method} ${req.path}`, {
    requestId,
    method: req.method,
    path: req.path,
    ip: req.ip || req.connection.remoteAddress
  });
  
  next();
});

// Handle root route
app.get('/', (req, res) => {
  res.json({
    message: 'Orbit Chain Deployer service is running',
    version: '1.0.0',
    environment: config.nodeEnv,
    endpoints: {
      health: '/health',
      deployChain: '/deploy-chain',
      deployContracts: '/deploy-contracts'
    }
  });
});

// Mount all routes
app.use('/', routes);

// Error handling middleware
app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
  logger.error('Unhandled error occurred', {
    error: err.message,
    path: req.path,
    method: req.method
  });
  
  res.status(500).json({
    success: false,
    message: 'Internal server error',
    error: isDevelopment ? err.message : 'Something went wrong'
  });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    success: false,
    message: 'Route not found',
    path: req.originalUrl
  });
});

// Graceful shutdown handling
const gracefulShutdown = (signal: string) => {
  logger.info(`Received ${signal}, shutting down gracefully`);
  
  process.exit(0);
};

process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));

// Start server
const server = app.listen(Number(PORT), '0.0.0.0', () => {
  logger.info('Orbit Chain Deployer service started', {
    port: PORT,
    environment: config.nodeEnv
  });
});

// Handle server errors
server.on('error', (error: any) => {
  logger.error('Server error occurred', {
    error: error.message,
    code: error.code
  });
  
  if (error.code === 'EADDRINUSE') {
    logger.error(`Port ${PORT} is already in use`);
    process.exit(1);
  }
});
