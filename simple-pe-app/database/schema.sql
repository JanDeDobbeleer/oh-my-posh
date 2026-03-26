-- simple-pe-app | PostgreSQL Schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending_verification');
CREATE TYPE user_role AS ENUM ('user', 'admin', 'support');
CREATE TYPE account_type AS ENUM ('savings', 'current', 'investment');
CREATE TYPE account_status AS ENUM ('active', 'inactive', 'frozen');
CREATE TYPE transaction_type AS ENUM ('transfer', 'deposit', 'withdrawal', 'payment', 'refund');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled', 'reversed');
CREATE TYPE currency_code AS ENUM ('PEN', 'USD');

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION generate_account_number()
RETURNS VARCHAR(20) AS $$
DECLARE new_number VARCHAR(20); exists_flag BOOLEAN;
BEGIN
  LOOP
    new_number := 'SMP' || LPAD(FLOOR(RANDOM() * 9999999999999)::BIGINT::TEXT, 13, '0');
    SELECT EXISTS(SELECT 1 FROM accounts WHERE account_number = new_number) INTO exists_flag;
    EXIT WHEN NOT exists_flag;
  END LOOP;
  RETURN new_number;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  phone VARCHAR(20),
  dni VARCHAR(20) UNIQUE,
  status user_status DEFAULT 'pending_verification',
  role user_role DEFAULT 'user',
  email_verified_at TIMESTAMPTZ,
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE accounts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  account_number VARCHAR(20) UNIQUE NOT NULL DEFAULT generate_account_number(),
  balance DECIMAL(15,2) DEFAULT 0.00 NOT NULL CHECK (balance >= 0),
  available_balance DECIMAL(15,2) DEFAULT 0.00 NOT NULL,
  currency currency_code DEFAULT 'PEN',
  type account_type DEFAULT 'savings',
  status account_status DEFAULT 'active',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  from_account_id UUID REFERENCES accounts(id),
  to_account_id UUID REFERENCES accounts(id),
  amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
  currency currency_code DEFAULT 'PEN',
  type transaction_type NOT NULL,
  status transaction_status DEFAULT 'pending',
  description TEXT,
  reference_code VARCHAR(50) UNIQUE,
  metadata JSONB DEFAULT '{}',
  fee DECIMAL(10,2) DEFAULT 0.00,
  exchange_rate DECIMAL(10,6) DEFAULT 1.000000,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id),
  entity_type VARCHAR(50) NOT NULL,
  entity_id UUID,
  action VARCHAR(50) NOT NULL,
  old_values JSONB,
  new_values JSONB,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE notifications (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL,
  title VARCHAR(255) NOT NULL,
  body TEXT,
  data JSONB DEFAULT '{}',
  read_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_account_number ON accounts(account_number);
CREATE INDEX idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_from_created_at ON transactions(from_account_id, created_at DESC);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

-- Triggers
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_accounts_updated_at BEFORE UPDATE ON accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER trg_transactions_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seeds
INSERT INTO users (id, email, password_hash, first_name, last_name, phone, dni, status, role, email_verified_at)
VALUES
  ('a0000000-0000-0000-0000-000000000001','admin@simple-pe.com','$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK.s5uFDi','Carlos','Ramirez','+51 999 111 000','10000001','active','admin',NOW()),
  ('b0000000-0000-0000-0000-000000000002','ana.flores@gmail.com','$2b$12$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uivHW/yLi','Ana','Flores','+51 987 654 321','45123678','active','user',NOW()),
  ('c0000000-0000-0000-0000-000000000003','luis.quispe@hotmail.com','$2b$12$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uivHW/yLi','Luis','Quispe','+51 956 789 012','72345901','active','user',NOW());

INSERT INTO accounts (id, user_id, account_number, balance, available_balance, currency, type, status)
VALUES
  ('d0000000-0000-0000-0000-000000000001','b0000000-0000-0000-0000-000000000002','SMP0000000000001',1850.50,1850.50,'PEN','savings','active'),
  ('d0000000-0000-0000-0000-000000000002','b0000000-0000-0000-0000-000000000002','SMP0000000000002',320.00,320.00,'USD','savings','active'),
  ('d0000000-0000-0000-0000-000000000003','c0000000-0000-0000-0000-000000000003','SMP0000000000003',5430.75,5430.75,'PEN','current','active');
