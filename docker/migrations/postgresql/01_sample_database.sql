-- PostgreSQL Sample Database for QueryBase Testing
-- Covers: SELECT, INSERT, UPDATE, DELETE, multi-line queries, JSONB columns, no-match scenarios

-- Create schema
CREATE SCHEMA IF NOT EXISTS sample_app;
SET search_path TO sample_app;

-- ============================================================
-- Tables
-- ============================================================

CREATE TABLE IF NOT EXISTS customers (
    id          SERIAL PRIMARY KEY,
    first_name  VARCHAR(50)  NOT NULL,
    last_name   VARCHAR(50)  NOT NULL,
    email       VARCHAR(100) NOT NULL UNIQUE,
    phone       VARCHAR(20),
    address     JSONB,                          -- structured address as JSON
    metadata    JSONB DEFAULT '{}',             -- arbitrary customer metadata
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(100) NOT NULL,
    description    TEXT,
    category       VARCHAR(50),
    price          NUMERIC(10, 2) NOT NULL,
    stock_quantity INT DEFAULT 0,
    sku            VARCHAR(50) UNIQUE,
    attributes     JSONB DEFAULT '{}',          -- flexible product attributes as JSON
    tags           JSONB DEFAULT '[]',          -- array of tags
    is_active      BOOLEAN DEFAULT TRUE,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id               SERIAL PRIMARY KEY,
    customer_id      INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    order_number     VARCHAR(50) NOT NULL UNIQUE,
    order_date       TIMESTAMPTZ DEFAULT NOW(),
    total_amount     NUMERIC(10, 2) NOT NULL,
    status           VARCHAR(20) DEFAULT 'pending'
                         CHECK (status IN ('pending','processing','shipped','delivered','cancelled')),
    shipping_address JSONB,                     -- shipping destination as JSON
    line_items       JSONB DEFAULT '[]',        -- snapshot of order lines
    notes            TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id         SERIAL PRIMARY KEY,
    order_id   INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity   INT NOT NULL DEFAULT 1,
    unit_price NUMERIC(10, 2) NOT NULL,
    subtotal   NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS event_logs (
    id         BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    payload    JSONB NOT NULL,                  -- full event payload
    source     VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Customers (with JSONB address and metadata)
-- ============================================================

INSERT INTO customers (first_name, last_name, email, phone, address, metadata) VALUES
('John',    'Doe',      'john.doe@example.com',      '555-0101',
 '{"street":"123 Main St","city":"New York","state":"NY","zip":"10001","country":"USA"}',
 '{"tier":"gold","notes":"VIP customer","tags":["loyal","early-adopter"]}'),

('Jane',    'Smith',    'jane.smith@example.com',    '555-0102',
 '{"street":"456 Oak Ave","city":"Los Angeles","state":"CA","zip":"90001","country":"USA"}',
 '{"tier":"silver","referral_code":"JANE20","preferences":{"newsletter":true,"sms":false}}'),

('Bob',     'Johnson',  'bob.johnson@example.com',   '555-0103',
 '{"street":"789 Pine Rd","city":"Chicago","state":"IL","zip":"60601","country":"USA"}',
 '{"tier":"bronze","notes":"New customer"}'),

('Alice',   'Williams', 'alice.williams@example.com','555-0104',
 '{"street":"321 Elm St","city":"Houston","state":"TX","zip":"77001","country":"USA"}',
 '{"tier":"gold","tags":["bulk-buyer"],"credit_limit":5000}'),

('Charlie', 'Brown',    'charlie.brown@example.com', '555-0105',
 '{"street":"654 Maple Dr","city":"Phoenix","state":"AZ","zip":"85001","country":"USA"}',
 '{"tier":"silver","preferences":{"newsletter":false,"sms":true}}'),

('Diana',   'Davis',    'diana.davis@example.com',   '555-0106',
 '{"street":"987 Cedar Ln","city":"Philadelphia","state":"PA","zip":"19101","country":"USA"}',
 '{"tier":"bronze"}'),

('Eve',     'Miller',   'eve.miller@example.com',    '555-0107',
 '{"street":"147 Birch Ct","city":"San Antonio","state":"TX","zip":"78201","country":"USA"}',
 '{"tier":"silver","tags":["repeat-buyer","wholesale"]}'),

('Frank',   'Wilson',   'frank.wilson@example.com',  '555-0108',
 '{"street":"258 Spruce Way","city":"San Diego","state":"CA","zip":"92101","country":"USA"}',
 '{"tier":"gold","credit_limit":10000,"notes":"Corporate account"}'),

('Grace',   'Moore',    'grace.moore@example.com',   '555-0109',
 '{"street":"369 Ash Blvd","city":"Dallas","state":"TX","zip":"75201","country":"USA"}',
 '{"tier":"bronze","tags":["new"]}'),

('Henry',   'Taylor',   'henry.taylor@example.com',  '555-0110',
 '{"street":"741 Walnut St","city":"San Jose","state":"CA","zip":"95101","country":"USA"}',
 '{"tier":"gold","tags":["loyal","high-value"],"credit_limit":20000}'),

-- Dedicated row for DELETE/no-match testing
('Test',    'Customer', 'test.delete@example.com',   '555-9999',
 '{"street":"999 Test St","city":"Testville","state":"TC","zip":"99999","country":"USA"}',
 '{"tier":"test","notes":"Safe to delete — used for DELETE query testing"}');

-- ============================================================
-- Products (with JSONB attributes and tags)
-- ============================================================

INSERT INTO products (name, description, category, price, stock_quantity, sku, attributes, tags) VALUES
('Laptop Pro 15',  'High-performance 15" laptop',             'Electronics', 1299.99,  50, 'ELEC-LAP-001',
 '{"brand":"TechCorp","ram_gb":16,"storage_gb":512,"display_inch":15,"color":"Space Gray"}',
 '["laptop","portable","high-performance"]'),

('Wireless Mouse',  'Ergonomic wireless mouse with USB dongle', 'Electronics',   29.99, 200, 'ELEC-MOU-001',
 '{"brand":"ClickMaster","dpi":1600,"wireless":true,"battery_life_months":12}',
 '["mouse","wireless","ergonomic"]'),

('Office Chair',    'Ergonomic chair with lumbar support',      'Furniture',    249.99,  30, 'FURN-CHA-001',
 '{"brand":"ComfortSeat","max_weight_kg":120,"adjustable_height":true,"material":"mesh"}',
 '["chair","ergonomic","office"]'),

('Standing Desk',   'Motorized height-adjustable desk',         'Furniture',    499.99,  20, 'FURN-DSK-001',
 '{"brand":"ErgoDesk","width_cm":150,"height_range_cm":[70,120],"motor":"dual","color":"White"}',
 '["desk","standing","motorized"]'),

('Coffee Maker',    'Programmable coffee maker, thermal carafe','Appliances',    89.99,  75, 'APPL-COF-001',
 '{"brand":"BrewMaster","capacity_cups":12,"programmable":true,"keep_warm":true}',
 '["coffee","kitchen","programmable"]'),

('Blender Pro',     'High-speed blender for smoothies',         'Appliances',   149.99,  60, 'APPL-BLE-001',
 '{"brand":"BlendKing","watts":1500,"speeds":10,"dishwasher_safe":true}',
 '["blender","kitchen","high-speed"]'),

('Yoga Mat',        'Non-slip yoga mat with carrying strap',    'Sports',        34.99, 100, 'SPRT-YOG-001',
 '{"brand":"ZenFlex","thickness_mm":6,"material":"TPE","length_cm":183,"non_slip":true}',
 '["yoga","fitness","mat"]'),

('Running Shoes',   'Lightweight running shoes, cushioned sole','Sports',       119.99,  80, 'SPRT-SHO-001',
 '{"brand":"SwiftStep","sizes_available":[7,8,9,10,11,12],"drop_mm":8,"waterproof":false}',
 '["shoes","running","lightweight"]'),

('Backpack',        'Durable backpack with laptop compartment', 'Accessories',   59.99, 120, 'ACCS-BAG-001',
 '{"brand":"CarryAll","capacity_liters":30,"laptop_compartment":true,"waterproof":true}',
 '["bag","backpack","travel"]'),

('Water Bottle',    'Insulated stainless steel water bottle',   'Accessories',   24.99, 150, 'ACCS-BOT-001',
 '{"brand":"HydroKeep","capacity_ml":750,"insulated":true,"material":"stainless_steel","colors":["black","silver","blue"]}',
 '["bottle","hydration","insulated"]');

-- ============================================================
-- Orders (with JSONB shipping_address and line_items snapshot)
-- ============================================================

INSERT INTO orders (customer_id, order_number, total_amount, status, shipping_address, line_items) VALUES
(1, 'ORD-2024-001', 1329.98, 'delivered',
 '{"street":"123 Main St","city":"New York","state":"NY","zip":"10001"}',
 '[{"sku":"ELEC-LAP-001","name":"Laptop Pro 15","qty":1,"price":1299.99},{"sku":"ELEC-MOU-001","name":"Wireless Mouse","qty":1,"price":29.99}]'),

(2, 'ORD-2024-002', 279.98, 'shipped',
 '{"street":"456 Oak Ave","city":"Los Angeles","state":"CA","zip":"90001"}',
 '[{"sku":"FURN-CHA-001","name":"Office Chair","qty":1,"price":249.99},{"sku":"ELEC-MOU-001","name":"Wireless Mouse","qty":1,"price":29.99}]'),

(3, 'ORD-2024-003', 749.98, 'processing',
 '{"street":"789 Pine Rd","city":"Chicago","state":"IL","zip":"60601"}',
 '[{"sku":"FURN-DSK-001","name":"Standing Desk","qty":1,"price":499.99},{"sku":"FURN-CHA-001","name":"Office Chair","qty":1,"price":249.99}]'),

(4, 'ORD-2024-004', 89.99, 'delivered',
 '{"street":"321 Elm St","city":"Houston","state":"TX","zip":"77001"}',
 '[{"sku":"APPL-COF-001","name":"Coffee Maker","qty":1,"price":89.99}]'),

(5, 'ORD-2024-005', 199.97, 'pending',
 '{"street":"654 Maple Dr","city":"Phoenix","state":"AZ","zip":"85001"}',
 '[{"sku":"APPL-BLE-001","name":"Blender Pro","qty":1,"price":149.99},{"sku":"ACCS-BOT-001","name":"Water Bottle","qty":2,"price":24.99}]'),

(1, 'ORD-2024-006', 59.99, 'delivered',
 '{"street":"123 Main St","city":"New York","state":"NY","zip":"10001"}',
 '[{"sku":"ACCS-BAG-001","name":"Backpack","qty":1,"price":59.99}]'),

(2, 'ORD-2024-007', 499.99, 'shipped',
 '{"street":"456 Oak Ave","city":"Los Angeles","state":"CA","zip":"90001"}',
 '[{"sku":"FURN-DSK-001","name":"Standing Desk","qty":1,"price":499.99}]'),

(6, 'ORD-2024-008', 119.99, 'delivered',
 '{"street":"987 Cedar Ln","city":"Philadelphia","state":"PA","zip":"19101"}',
 '[{"sku":"SPRT-SHO-001","name":"Running Shoes","qty":1,"price":119.99}]'),

(7, 'ORD-2024-009', 24.99, 'delivered',
 '{"street":"147 Birch Ct","city":"San Antonio","state":"TX","zip":"78201"}',
 '[{"sku":"ACCS-BOT-001","name":"Water Bottle","qty":1,"price":24.99}]'),

(8, 'ORD-2024-010', 1299.99, 'processing',
 '{"street":"258 Spruce Way","city":"San Diego","state":"CA","zip":"92101"}',
 '[{"sku":"ELEC-LAP-001","name":"Laptop Pro 15","qty":1,"price":1299.99}]');

-- ============================================================
-- Order Items
-- ============================================================

INSERT INTO order_items (order_id, product_id, quantity, unit_price, subtotal) VALUES
(1, 1, 1, 1299.99, 1299.99), (1, 2, 1,   29.99,   29.99),
(2, 3, 1,  249.99,  249.99), (2, 2, 1,   29.99,   29.99),
(3, 4, 1,  499.99,  499.99), (3, 3, 1,  249.99,  249.99),
(4, 5, 1,   89.99,   89.99),
(5, 6, 1,  149.99,  149.99), (5,10, 2,   24.99,   49.98),
(6, 9, 1,   59.99,   59.99),
(7, 4, 1,  499.99,  499.99),
(8, 8, 1,  119.99,  119.99),
(9,10, 1,   24.99,   24.99),
(10,1, 1, 1299.99, 1299.99);

-- ============================================================
-- Event Logs (pure JSONB payload, for JSON query testing)
-- ============================================================

INSERT INTO event_logs (event_type, payload, source) VALUES
('user_login',
 '{"user_id":1,"email":"john.doe@example.com","ip":"192.168.1.10","user_agent":"Mozilla/5.0","success":true}',
 'auth-service'),

('order_created',
 '{"order_id":1,"order_number":"ORD-2024-001","customer_id":1,"total":1329.98,"items":[{"sku":"ELEC-LAP-001","qty":1},{"sku":"ELEC-MOU-001","qty":1}]}',
 'order-service'),

('payment_processed',
 '{"order_id":1,"amount":1329.98,"currency":"USD","gateway":"stripe","transaction_id":"ch_3abc123","status":"succeeded","card_last4":"4242"}',
 'payment-service'),

('order_shipped',
 '{"order_id":1,"order_number":"ORD-2024-001","carrier":"FedEx","tracking_number":"FX123456789","estimated_delivery":"2024-02-10"}',
 'fulfillment-service'),

('user_login',
 '{"user_id":2,"email":"jane.smith@example.com","ip":"10.0.0.55","user_agent":"Safari/17.0","success":true}',
 'auth-service'),

('order_created',
 '{"order_id":2,"order_number":"ORD-2024-002","customer_id":2,"total":279.98,"items":[{"sku":"FURN-CHA-001","qty":1},{"sku":"ELEC-MOU-001","qty":1}]}',
 'order-service'),

('login_failed',
 '{"email":"unknown@hacker.io","ip":"203.0.113.77","reason":"user_not_found","attempt":3}',
 'auth-service'),

('inventory_alert',
 '{"product_id":4,"sku":"FURN-DSK-001","name":"Standing Desk","current_stock":20,"threshold":25,"severity":"warning"}',
 'inventory-service'),

('payment_failed',
 '{"order_id":5,"amount":199.97,"currency":"USD","gateway":"stripe","error_code":"card_declined","error_message":"Your card was declined","retry_allowed":true}',
 'payment-service'),

('order_cancelled',
 '{"order_id":5,"order_number":"ORD-2024-005","reason":"payment_failed","refund_issued":false,"cancelled_by":"system"}',
 'order-service');

-- ============================================================
-- View for order summary
-- ============================================================

CREATE OR REPLACE VIEW order_summary AS
SELECT
    o.id                                        AS order_id,
    o.order_number,
    c.first_name || ' ' || c.last_name         AS customer_name,
    c.email                                     AS customer_email,
    o.order_date,
    o.total_amount,
    o.status,
    o.shipping_address->>'city'                 AS shipping_city,
    COUNT(oi.id)                                AS item_count,
    c.metadata->>'tier'                         AS customer_tier
FROM orders o
JOIN customers c ON o.customer_id = c.id
LEFT JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.order_number, c.first_name, c.last_name, c.email,
         o.order_date, o.total_amount, o.status, o.shipping_address, c.metadata;

SELECT 'Sample database ready' AS status;
