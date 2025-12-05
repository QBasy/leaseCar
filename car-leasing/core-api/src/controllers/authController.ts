import { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';
import AuthService from '../services/authService';
import { z } from 'zod';

const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6)
});

export default class AuthController {
  app: FastifyInstance;
  service: AuthService;
  constructor(app: FastifyInstance) {
    this.app = app;
    this.service = new AuthService(app);
  }

  async login(req: FastifyRequest, reply: FastifyReply) {
    try {
      const body = LoginSchema.parse(req.body);
      const token = await this.service.login(body.email, body.password);
      return reply.send({ token });
    } catch (err: any) {
      this.app.log.error(err);
      return reply.status(400).send({ error: err?.message || 'Invalid credentials' });
    }
  }
}
