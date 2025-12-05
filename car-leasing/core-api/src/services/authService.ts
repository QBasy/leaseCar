import { FastifyInstance } from 'fastify';
import UserRepository from '../repositories/userRepository';
import bcrypt from 'bcrypt';

export default class AuthService {
  app: FastifyInstance;
  repo: UserRepository;
  constructor(app: FastifyInstance) {
    this.app = app;
    this.repo = new UserRepository();
  }

  async login(email: string, password: string): Promise<string> {
    const user = await this.repo.findByEmail(email);
    if (!user) throw new Error('User not found');
    const ok = await bcrypt.compare(password, user.password_hash);
    if (!ok) throw new Error('Invalid password');

    const token = this.app.jwt.sign({ userId: user.id, email: user.email });
    return token;
  }
}
