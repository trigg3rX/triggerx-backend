import express from 'express';
import cors from 'cors';
import dotenv from 'dotenv';

// Load environment variables
dotenv.config();

const app = express();
const PORT = process.env.PORT || 3001;

// Middleware
app.use(cors());
app.use(express.json());

// Request logging middleware
app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
  if (req.body && Object.keys(req.body).length > 0) {
    console.log('Request body:', JSON.stringify(req.body, null, 2));
  }
  next();
});

// Handle root route
app.get('/', (req, res) => {
  res.json({
    message: 'Orbit Chain Deployer service is running',
    version: '1.0.0',
    endpoints: {
      health: '/health',
      deployChain: '/deploy-chain',
      deployContracts: '/deploy-contracts'
    }
  });
});

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: '1.0.0'
  });
});

// Deploy Chain endpoint - called by Go backend to deploy a new Orbit chain
app.post('/deploy-chain', (req, res) => {
  console.log('Deploy Chain request received:', req.body);
  
  // TODO: Implement actual orbit chain deployment logic
  // For now, return success response as placeholder
  
  const response = {
    success: true,
    deployment_id: req.body.deployment_id || 'placeholder-deployment-id',
    status: 'pending',
    message: 'Chain deployment initiated (placeholder)',
    chain_address: '0x0000000000000000000000000000000000000000' // Placeholder address
  };
  
  console.log('Deploy Chain response:', response);
  res.status(200).json(response);
});

// Deploy Contracts endpoint - called by Go backend to deploy TriggerX contracts
app.post('/deploy-contracts', (req, res) => {
  console.log('Deploy Contracts request received:', req.body);
  
  // TODO: Implement actual contract deployment logic
  // For now, return success response as placeholder
  
  const response = {
    success: true,
    deployment_id: req.body.deployment_id || 'placeholder-deployment-id',
    status: 'completed',
    message: 'Contracts deployment completed (placeholder)'
  };
  
  console.log('Deploy Contracts response:', response);
  res.status(200).json(response);
});

// Error handling middleware
app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error('Error occurred:', err);
  res.status(500).json({
    success: false,
    message: 'Internal server error',
    error: err.message
  });
});

// Start server
app.listen(PORT, () => {
  console.log(`Orbit Chain Deployer service running on port ${PORT}`);
});
