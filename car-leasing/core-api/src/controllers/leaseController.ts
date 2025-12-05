import { FastifyInstance, FastifyReply, FastifyRequest } from 'fastify';

export default class LeaseController {
  app: FastifyInstance;
  constructor(app: FastifyInstance) {
    this.app = app;
  }

  async search(req: FastifyRequest, reply: FastifyReply) {
    const q = (req.query as any)?.q || '';
    try {
      const index = this.app.meili.index('leases');
      const res = await index.search(q, { limit: 20 });
      return reply.send(res.hits);
    } catch (err: any) {
      this.app.log.error(err);
      return reply.status(500).send({ error: 'search failed' });
    }
  }

  async getById(req: FastifyRequest, reply: FastifyReply) {
    const id = (req.params as any).id;
    try {
      // Proxy to lease-service
      const leaseServiceUrl = process.env.LEASE_SERVICE_URL || 'http://lease-service:3001';
      const r = await fetch(`${leaseServiceUrl}/leases/${id}`);
      if (!r.ok) return reply.status(r.status).send(await r.text());
      const body = await r.json();
      return reply.send(body);
    } catch (err: any) {
      this.app.log.error(err);
      return reply.status(500).send({ error: 'fetch lease failed' });
    }
  }
}
