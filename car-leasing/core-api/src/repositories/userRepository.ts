import { Pool } from 'pg';

const pool = new Pool({
  host: process.env.POSTGRES_HOST || 'postgres',
  port: Number(process.env.POSTGRES_PORT || 5432),
  user: process.env.POSTGRES_USER || 'leasing_user',
  password: process.env.POSTGRES_PASSWORD || 'secure_pass',
  database: process.env.POSTGRES_DB || 'leasing_db'
});

export default class UserRepository {
  async findByEmail(email: string) {
    const res = await pool.query('SELECT id, email, password_hash FROM users WHERE email = $1 LIMIT 1', [email]);
    return res.rows[0];
  }
}
