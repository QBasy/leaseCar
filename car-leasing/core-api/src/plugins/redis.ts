import fp from 'fastify-plugin';
import Redis from 'ioredis';
import { FastifyInstance } from 'fastify';

export default fp(async (app: FastifyInstance) => {
  const host = process.env.REDIS_HOST || app.config?.redis?.host || '127.0.0.1';
  const port = Number(process.env.REDIS_PORT || app.config?.redis?.port || 6379);
  const client = new Redis({ host, port });

  app.decorate('redis', client);
  app.addHook('onClose', async (instance) => {
    await client.quit();
  });
});
