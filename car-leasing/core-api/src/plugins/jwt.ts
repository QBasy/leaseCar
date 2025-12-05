import fp from 'fastify-plugin';
import fastifyJwt from '@fastify/jwt';
import { FastifyInstance } from 'fastify';

export default fp(async (app: FastifyInstance) => {
  const secret = process.env.JWT_SECRET || app.config?.jwt?.secret || 'dev-secret';
  app.register(fastifyJwt, { secret });

  app.decorate('authenticate', async (req: any, reply: any) => {
    try {
      await req.jwtVerify();
    } catch (err) {
      reply.send(err);
    }
  });
});
