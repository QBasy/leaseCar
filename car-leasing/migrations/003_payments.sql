-- 003_payments.sql - Payment processing schema

CREATE TYPE payment_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'REFUNDED', 'CANCELLED');
CREATE TYPE payment_method AS ENUM ('CARD', 'BANK_TRANSFER', 'CRYPTO', 'WALLET');
CREATE TYPE payment_provider AS ENUM ('STRIPE', 'BANK_API', 'TON_BLOCKCHAIN');

CREATE TABLE IF NOT EXISTS payments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  lease_id UUID NOT NULL REFERENCES leases(id) ON DELETE CASCADE,
  lease_payment_id UUID REFERENCES lease_payments(id) ON DELETE SET NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount DECIMAL(10, 2) NOT NULL,
  currency VARCHAR(3) DEFAULT 'USD',
  status payment_status DEFAULT 'PENDING',
  method payment_method NOT NULL,
  provider payment_provider NOT NULL,
  transaction_id VARCHAR(255),
  blockchain_tx_hash VARCHAR(255),
  error_message TEXT,
  metadata JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP
);

CREATE INDEX idx_payments_lease_id ON payments(lease_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);

CREATE TABLE IF NOT EXISTS payment_webhooks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
  provider payment_provider NOT NULL,
  event_type VARCHAR(100) NOT NULL,
  event_data JSONB NOT NULL,
  processed BOOLEAN DEFAULT false,
  processed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payment_webhooks_payment_id ON payment_webhooks(payment_id);
CREATE INDEX idx_payment_webhooks_processed ON payment_webhooks(processed);

CREATE TABLE IF NOT EXISTS blockchain_transactions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
  tx_hash VARCHAR(255) UNIQUE NOT NULL,
  from_address VARCHAR(255) NOT NULL,
  to_address VARCHAR(255) NOT NULL,
  amount VARCHAR(100) NOT NULL,
  status VARCHAR(50) DEFAULT 'PENDING',
  confirmed BOOLEAN DEFAULT false,
  confirmed_at TIMESTAMP,
  error TEXT,
  raw_response JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_blockchain_transactions_tx_hash ON blockchain_transactions(tx_hash);
CREATE INDEX idx_blockchain_transactions_payment_id ON blockchain_transactions(payment_id);
CREATE INDEX idx_blockchain_transactions_status ON blockchain_transactions(status);
