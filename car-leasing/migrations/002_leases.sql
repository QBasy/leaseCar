-- 002_leases.sql - Leases and vehicles schema

CREATE TYPE lease_status AS ENUM ('DRAFT', 'ACTIVE', 'PENDING_APPROVAL', 'REJECTED', 'COMPLETED', 'TERMINATED');
CREATE TYPE vehicle_type AS ENUM ('SEDAN', 'SUV', 'TRUCK', 'VAN', 'SPORTS', 'HYBRID', 'ELECTRIC');

CREATE TABLE IF NOT EXISTS vehicles (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  make VARCHAR(100) NOT NULL,
  model VARCHAR(100) NOT NULL,
  year INTEGER NOT NULL,
  vin VARCHAR(17) UNIQUE NOT NULL,
  license_plate VARCHAR(20) UNIQUE,
  vehicle_type vehicle_type NOT NULL,
  color VARCHAR(50),
  mileage INTEGER DEFAULT 0,
  price_per_month DECIMAL(10, 2) NOT NULL,
  deposit_amount DECIMAL(10, 2),
  description TEXT,
  image_url VARCHAR(500),
  available BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicles_available ON vehicles(available);
CREATE INDEX idx_vehicles_created_at ON vehicles(created_at);

CREATE TABLE IF NOT EXISTS leases (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
  status lease_status DEFAULT 'DRAFT',
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  monthly_payment DECIMAL(10, 2) NOT NULL,
  deposit_paid DECIMAL(10, 2) DEFAULT 0,
  total_cost DECIMAL(10, 2),
  mileage_limit INTEGER DEFAULT 50000,
  notes TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  approved_at TIMESTAMP,
  started_at TIMESTAMP,
  ended_at TIMESTAMP
);

CREATE INDEX idx_leases_user_id ON leases(user_id);
CREATE INDEX idx_leases_vehicle_id ON leases(vehicle_id);
CREATE INDEX idx_leases_status ON leases(status);
CREATE INDEX idx_leases_created_at ON leases(created_at);

CREATE TABLE IF NOT EXISTS lease_payments (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  lease_id UUID NOT NULL REFERENCES leases(id) ON DELETE CASCADE,
  payment_number INTEGER NOT NULL,
  due_date DATE NOT NULL,
  amount DECIMAL(10, 2) NOT NULL,
  paid_amount DECIMAL(10, 2) DEFAULT 0,
  paid_at TIMESTAMP,
  status VARCHAR(20) DEFAULT 'PENDING',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_lease_payments_lease_id ON lease_payments(lease_id);
CREATE INDEX idx_lease_payments_status ON lease_payments(status);
CREATE INDEX idx_lease_payments_due_date ON lease_payments(due_date);
