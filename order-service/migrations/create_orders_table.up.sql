CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  number TEXT UNIQUE NOT NULL,
  customer_name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('dine_in', 'takeout', 'delivery')),
  table_number INTEGER,
  delivery_address TEXT,
  total_amount DECIMAL(10, 2) NOT NULL,
  priority INTEGER DEFAULT 1,
  status TEXT DEFAULT 'received',
  processed_by TEXT,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  quantity INTEGER NOT NULL,
  price DECIMAL(8, 2) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  order_id INTEGER NOT NULL,
  FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE TABLE order_status_log (
  id SERIAL PRIMARY KEY,
  status TEXT,
  changed_by TEXT,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  order_id INTEGER NOT NULL,
  FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE
);