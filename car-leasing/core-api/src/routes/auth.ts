import { FastifyInstance } from 'fastify';
import AuthController from '../controllers/authController';

export async function authRoutes(app: FastifyInstance) {
  const controller = new AuthController(app);

  app.post('/login', async (req, reply) => controller.login(req, reply));
}
