import { Router } from 'express';
import healthRoutes from './health';
import deployRoutes from './deploy';

const router = Router();

// Mount route modules
router.use('/', healthRoutes);
router.use('/', deployRoutes);

export default router;
