-- MySQL Sample Database for QueryBase Testing
-- This script creates a sample application database with test data

-- Create sample_app database
CREATE DATABASE IF NOT EXISTS sample_app;
USE sample_app;

-- Customers table
CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    phone VARCHAR(20),
    address VARCHAR(200),
    city VARCHAR(50),
    state VARCHAR(50),
    zip_code VARCHAR(10),
    country VARCHAR(50) DEFAULT 'USA',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_city (city)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INT DEFAULT 0,
    sku VARCHAR(50) UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_sku (sku)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_amount DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'processing', 'shipped', 'delivered', 'cancelled') DEFAULT 'pending',
    shipping_address VARCHAR(200),
    shipping_city VARCHAR(50),
    shipping_state VARCHAR(50),
    shipping_zip VARCHAR(10),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
    INDEX idx_customer (customer_id),
    INDEX idx_status (status),
    INDEX idx_order_date (order_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Order Items table
CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    INDEX idx_order (order_id),
    INDEX idx_product (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Insert sample customers
INSERT INTO customers (first_name, last_name, email, phone, address, city, state, zip_code) VALUES
('John', 'Doe', 'john.doe@example.com', '555-0101', '123 Main St', 'New York', 'NY', '10001'),
('Jane', 'Smith', 'jane.smith@example.com', '555-0102', '456 Oak Ave', 'Los Angeles', 'CA', '90001'),
('Bob', 'Johnson', 'bob.johnson@example.com', '555-0103', '789 Pine Rd', 'Chicago', 'IL', '60601'),
('Alice', 'Williams', 'alice.williams@example.com', '555-0104', '321 Elm St', 'Houston', 'TX', '77001'),
('Charlie', 'Brown', 'charlie.brown@example.com', '555-0105', '654 Maple Dr', 'Phoenix', 'AZ', '85001'),
('Diana', 'Davis', 'diana.davis@example.com', '555-0106', '987 Cedar Ln', 'Philadelphia', 'PA', '19101'),
('Eve', 'Miller', 'eve.miller@example.com', '555-0107', '147 Birch Ct', 'San Antonio', 'TX', '78201'),
('Frank', 'Wilson', 'frank.wilson@example.com', '555-0108', '258 Spruce Way', 'San Diego', 'CA', '92101'),
('Grace', 'Moore', 'grace.moore@example.com', '555-0109', '369 Ash Blvd', 'Dallas', 'TX', '75201'),
('Henry', 'Taylor', 'henry.taylor@example.com', '555-0110', '741 Walnut St', 'San Jose', 'CA', '95101'),
-- Add a test customer with ID 999 for deletion testing
('Test', 'Customer', 'test.customer@example.com', '555-9999', '999 Test St', 'Test City', 'TC', '99999');

-- Set the last customer ID to 999 for the test customer
UPDATE customers SET id = 999 WHERE email = 'test.customer@example.com';

-- Insert sample products
INSERT INTO products (name, description, category, price, stock_quantity, sku) VALUES
('Laptop Pro 15', 'High-performance laptop with 15-inch display', 'Electronics', 1299.99, 50, 'ELEC-LAP-001'),
('Wireless Mouse', 'Ergonomic wireless mouse with USB receiver', 'Electronics', 29.99, 200, 'ELEC-MOU-001'),
('Office Chair', 'Ergonomic office chair with lumbar support', 'Furniture', 249.99, 30, 'FURN-CHA-001'),
('Standing Desk', 'Adjustable height standing desk', 'Furniture', 499.99, 20, 'FURN-DSK-001'),
('Coffee Maker', 'Programmable coffee maker with thermal carafe', 'Appliances', 89.99, 75, 'APPL-COF-001'),
('Blender Pro', 'High-speed blender for smoothies and more', 'Appliances', 149.99, 60, 'APPL-BLE-001'),
('Yoga Mat', 'Non-slip yoga mat with carrying strap', 'Sports', 34.99, 100, 'SPRT-YOG-001'),
('Running Shoes', 'Lightweight running shoes with cushioned sole', 'Sports', 119.99, 80, 'SPRT-SHO-001'),
('Backpack', 'Durable backpack with laptop compartment', 'Accessories', 59.99, 120, 'ACCS-BAG-001'),
('Water Bottle', 'Insulated stainless steel water bottle', 'Accessories', 24.99, 150, 'ACCS-BOT-001');

-- Insert sample orders
INSERT INTO orders (customer_id, order_number, total_amount, status, shipping_address, shipping_city, shipping_state, shipping_zip) VALUES
(1, 'ORD-2024-001', 1329.98, 'delivered', '123 Main St', 'New York', 'NY', '10001'),
(2, 'ORD-2024-002', 279.98, 'shipped', '456 Oak Ave', 'Los Angeles', 'CA', '90001'),
(3, 'ORD-2024-003', 749.98, 'processing', '789 Pine Rd', 'Chicago', 'IL', '60601'),
(4, 'ORD-2024-004', 89.99, 'delivered', '321 Elm St', 'Houston', 'TX', '77001'),
(5, 'ORD-2024-005', 154.98, 'pending', '654 Maple Dr', 'Phoenix', 'AZ', '85001'),
(1, 'ORD-2024-006', 59.99, 'delivered', '123 Main St', 'New York', 'NY', '10001'),
(2, 'ORD-2024-007', 499.99, 'shipped', '456 Oak Ave', 'Los Angeles', 'CA', '90001'),
(6, 'ORD-2024-008', 119.99, 'delivered', '987 Cedar Ln', 'Philadelphia', 'PA', '19101'),
(7, 'ORD-2024-009', 24.99, 'delivered', '147 Birch Ct', 'San Antonio', 'TX', '78201'),
(8, 'ORD-2024-010', 1299.99, 'processing', '258 Spruce Way', 'San Diego', 'CA', '92101');

-- Insert sample order items
INSERT INTO order_items (order_id, product_id, quantity, unit_price, subtotal) VALUES
-- Order 1
(1, 1, 1, 1299.99, 1299.99),
(1, 2, 1, 29.99, 29.99),
-- Order 2
(2, 3, 1, 249.99, 249.99),
(2, 2, 1, 29.99, 29.99),
-- Order 3
(3, 4, 1, 499.99, 499.99),
(3, 3, 1, 249.99, 249.99),
-- Order 4
(4, 5, 1, 89.99, 89.99),
-- Order 5
(5, 6, 1, 149.99, 149.99),
(5, 10, 2, 24.99, 49.98),
-- Order 6
(6, 9, 1, 59.99, 59.99),
-- Order 7
(7, 4, 1, 499.99, 499.99),
-- Order 8
(8, 8, 1, 119.99, 119.99),
-- Order 9
(9, 10, 1, 24.99, 24.99),
-- Order 10
(10, 1, 1, 1299.99, 1299.99);

-- Create a view for order summaries
CREATE OR REPLACE VIEW order_summary AS
SELECT 
    o.id AS order_id,
    o.order_number,
    CONCAT(c.first_name, ' ', c.last_name) AS customer_name,
    c.email AS customer_email,
    o.order_date,
    o.total_amount,
    o.status,
    COUNT(oi.id) AS item_count
FROM orders o
JOIN customers c ON o.customer_id = c.id
LEFT JOIN order_items oi ON o.id = oi.order_id
GROUP BY o.id, o.order_number, c.first_name, c.last_name, c.email, o.order_date, o.total_amount, o.status;

-- Grant permissions to querybase user
GRANT ALL PRIVILEGES ON sample_app.* TO 'querybase'@'%';
FLUSH PRIVILEGES;

-- Display summary
SELECT 'Database setup complete!' AS status;
SELECT COUNT(*) AS customer_count FROM customers;
SELECT COUNT(*) AS product_count FROM products;
SELECT COUNT(*) AS order_count FROM orders;
SELECT COUNT(*) AS order_item_count FROM order_items;
