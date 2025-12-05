import { FastifyInstance } from 'fastify';
import LeaseController from '../controllers/leaseController';

export async function leaseRoutes(app: FastifyInstance) {
  const controller = new LeaseController(app);

  app.get('/', async (req, reply) => controller.search(req, reply));
  app.get('/:id', async (req, reply) => controller.getById(req, reply));
}
