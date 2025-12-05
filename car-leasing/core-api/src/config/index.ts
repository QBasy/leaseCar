import fs from 'fs';
import path from 'path';
import yaml from 'js-yaml';

export interface CoreConfig {
  server: { host: string; port: number };
  jwt?: { secret: string; expiration: number };
  session?: { secret: string; expiration: number };
  redis?: { host: string; port: number };
  meilisearch?: { url: string; api_key?: string };
}

export function loadConfig(configPath = path.join(__dirname, '../../config/config.yaml')): CoreConfig {
  const file = fs.readFileSync(configPath, 'utf8');
  const cfg = yaml.load(file) as CoreConfig;
  // Allow env overrides
  if (!cfg.server) cfg.server = { host: '0.0.0.0', port: Number(process.env.FASTIFY_PORT ?? 3000) };
  if (process.env.JWT_SECRET) cfg.jwt = { secret: process.env.JWT_SECRET, expiration: Number(process.env.JWT_EXPIRATION ?? 86400) };
  if (process.env.MEILISEARCH_URL) cfg.meilisearch = { url: process.env.MEILISEARCH_URL, api_key: process.env.MEILISEARCH_API_KEY };
  return cfg;
}
