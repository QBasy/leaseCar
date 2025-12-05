import fp from 'fastify-plugin';
import { MeiliSearch } from 'meilisearch';
import { FastifyInstance } from 'fastify';

export default fp(async (app: FastifyInstance) => {
  const url = process.env.MEILISEARCH_URL || app.config?.meilisearch?.url || 'http://127.0.0.1:7700';
  const apiKey = process.env.MEILISEARCH_API_KEY || app.config?.meilisearch?.api_key;
  const client = new MeiliSearch({ host: url, apiKey: apiKey });

  app.decorate('meili', client);
});
