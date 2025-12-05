import Fastify from 'fastify';
import path from 'path';
import { loadConfig } from './config';
import jwtPlugin from './plugins/jwt';
import redisPlugin from './plugins/redis';
import meiliPlugin from './plugins/meilisearch';

// routes
import { authRoutes } from './routes/auth';
import { leaseRoutes } from './routes/leases';

async function build() {
  const configPath = process.env.CONFIG_PATH || path.join(__dirname, '../config/config.yaml');
  const config = loadConfig(configPath);

  const app = Fastify({ logger: true });
  // attach config for plugins
  (app as any).config = config;

  await app.register(jwtPlugin);
  await app.register(redisPlugin);
  await app.register(meiliPlugin);

  app.get('/health', async () => ({ status: 'ok' }));

  // register routes
  await app.register(authRoutes, { prefix: '/auth' });
  await app.register(leaseRoutes, { prefix: '/api/v1/leases' });

  return app;
}

if (require.main === module) {
  (async () => {
    const app = await build();
    const port = Number(process.env.FASTIFY_PORT || (app as any).config.server.port || 3000);
    await app.listen({ port, host: (app as any).config.server.host || '0.0.0.0' });
    console.log(`core-api listening on ${port}`);
  })();
}

export default build;
