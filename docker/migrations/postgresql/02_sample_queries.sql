-- ============================================================
-- QueryBase — Sample Queries for Testing
-- Database: sample_app (PostgreSQL)
-- Run 01_sample_database.sql first to populate the data.
-- ============================================================

-- ============================================================
-- 1. BASIC SELECT — single-line
-- ============================================================

SELECT * FROM sample_app.customers;

SELECT id, first_name, last_name, email FROM sample_app.customers WHERE is_active = true;

SELECT * FROM sample_app.products WHERE category = 'Electronics';

SELECT * FROM sample_app.orders WHERE status = 'pending';


-- ============================================================
-- 2. MULTI-LINE SELECT — tests newline parsing
-- ============================================================

SELECT
    c.id,
    c.first_name,
    c.last_name,
    c.email,
    c.metadata->>'tier' AS customer_tier
FROM sample_app.customers c
WHERE c.is_active = true
ORDER BY c.last_name ASC;


SELECT
    o.order_number,
    o.status,
    o.total_amount,
    o.shipping_address->>'city'  AS ship_to_city,
    o.shipping_address->>'state' AS ship_to_state
FROM sample_app.orders o
WHERE o.status IN ('pending', 'processing')
ORDER BY o.order_date DESC;


-- ============================================================
-- 3. JSON / JSONB QUERIES — reading JSON fields
-- ============================================================

-- Extract scalar fields from JSONB
SELECT
    id,
    first_name,
    last_name,
    address->>'street' AS street,
    address->>'city'   AS city,
    address->>'state'  AS state,
    address->>'zip'    AS zip
FROM sample_app.customers;


-- Filter by value inside JSONB
SELECT
    id,
    first_name,
    last_name,
    metadata->>'tier' AS tier,
    metadata->'tags'  AS tags
FROM sample_app.customers
WHERE metadata->>'tier' = 'gold';


-- Query JSONB array membership
SELECT
    id,
    name,
    attributes->>'brand'        AS brand,
    (attributes->>'ram_gb')::int AS ram_gb,
    tags
FROM sample_app.products
WHERE tags @> '["laptop"]';


-- Nested JSONB with event logs
SELECT
    id,
    event_type,
    payload->>'email'            AS user_email,
    payload->>'ip'               AS ip_address,
    (payload->>'success')::bool  AS success,
    created_at
FROM sample_app.event_logs
WHERE event_type = 'user_login';


-- JSON array elements from line_items snapshot
SELECT
    o.order_number,
    o.total_amount,
    jsonb_array_elements(o.line_items)->>'name' AS item_name,
    jsonb_array_elements(o.line_items)->>'qty'  AS qty,
    jsonb_array_elements(o.line_items)->>'price' AS unit_price
FROM sample_app.orders o
WHERE o.status = 'delivered';


-- Full payload for payment events
SELECT
    id,
    event_type,
    payload->>'order_id'        AS order_id,
    payload->>'amount'          AS amount,
    payload->>'gateway'         AS gateway,
    payload->>'status'          AS payment_status,
    payload->>'transaction_id'  AS transaction_id,
    created_at
FROM sample_app.event_logs
WHERE event_type IN ('payment_processed', 'payment_failed')
ORDER BY created_at;


-- ============================================================
-- 4. JOIN QUERIES — multi-table
-- ============================================================

SELECT
    o.order_number,
    c.first_name || ' ' || c.last_name AS customer,
    c.email,
    c.metadata->>'tier'                AS tier,
    o.total_amount,
    o.status,
    o.shipping_address->>'city'        AS city
FROM sample_app.orders o
JOIN sample_app.customers c ON c.id = o.customer_id
ORDER BY o.order_date DESC;


SELECT
    o.order_number,
    p.name    AS product,
    p.sku,
    oi.quantity,
    oi.unit_price,
    oi.subtotal,
    p.attributes->>'brand' AS brand
FROM sample_app.order_items oi
JOIN sample_app.orders o   ON o.id = oi.order_id
JOIN sample_app.products p ON p.id = oi.product_id
ORDER BY o.order_number, p.name;


-- ============================================================
-- 5. AGGREGATE QUERIES
-- ============================================================

SELECT
    status,
    COUNT(*)            AS order_count,
    SUM(total_amount)   AS total_revenue,
    AVG(total_amount)   AS avg_order_value,
    MIN(total_amount)   AS min_order,
    MAX(total_amount)   AS max_order
FROM sample_app.orders
GROUP BY status
ORDER BY total_revenue DESC;


SELECT
    c.metadata->>'tier'          AS tier,
    COUNT(DISTINCT c.id)         AS customer_count,
    COUNT(o.id)                  AS total_orders,
    COALESCE(SUM(o.total_amount), 0) AS total_spent
FROM sample_app.customers c
LEFT JOIN sample_app.orders o ON o.customer_id = c.id
GROUP BY tier
ORDER BY total_spent DESC;


SELECT
    category,
    COUNT(*)                     AS product_count,
    AVG(price)::NUMERIC(10,2)    AS avg_price,
    SUM(stock_quantity)          AS total_stock
FROM sample_app.products
WHERE is_active = true
GROUP BY category
ORDER BY avg_price DESC;


-- ============================================================
-- 6. VIEW QUERY
-- ============================================================

SELECT * FROM sample_app.order_summary ORDER BY order_id;


-- ============================================================
-- 7. QUERY WITH SEMICOLON INSIDE STRING LITERAL
-- (tests parser — semicolon must NOT split the statement)
-- ============================================================

SELECT
    id,
    first_name,
    CASE
        WHEN metadata->>'tier' = 'gold'   THEN 'Gold; Priority Support'
        WHEN metadata->>'tier' = 'silver' THEN 'Silver; Standard Support'
        ELSE                                   'Bronze; Self Service'
    END AS support_level
FROM sample_app.customers;


-- ============================================================
-- 8. EMPTY RESULT — no_match test
-- These queries execute successfully but return 0 rows.
-- ============================================================

SELECT * FROM sample_app.customers WHERE email = 'nobody@doesnotexist.com';

SELECT * FROM sample_app.orders WHERE status = 'processing' AND total_amount > 99999;

SELECT * FROM sample_app.event_logs WHERE event_type = 'nonexistent_event';


-- ============================================================
-- 9. INSERT — requires approval
-- ============================================================

INSERT INTO sample_app.customers (
    first_name, last_name, email, phone, address, metadata
) VALUES (
    'New',
    'Customer',
    'new.customer@example.com',
    '555-1234',
    '{"street":"10 Cloud St","city":"Seattle","state":"WA","zip":"98101","country":"USA"}',
    '{"tier":"bronze","notes":"Added via QueryBase test","tags":["new"]}'
);


INSERT INTO sample_app.event_logs (event_type, payload, source) VALUES
(
    'test_event',
    '{"message":"Hello from QueryBase","timestamp":"2024-02-01T12:00:00Z","data":{"key1":"value1","key2":42,"nested":{"flag":true}}}',
    'querybase-test'
);


-- ============================================================
-- 10. UPDATE — requires approval, affects known rows
-- ============================================================

UPDATE sample_app.customers
SET
    metadata = metadata || '{"tier":"gold","upgraded_at":"2024-02-01"}',
    updated_at = NOW()
WHERE email = 'bob.johnson@example.com';


UPDATE sample_app.products
SET
    stock_quantity = stock_quantity + 50,
    updated_at     = NOW()
WHERE category = 'Electronics'
  AND stock_quantity < 100;


-- ============================================================
-- 11. UPDATE — no_match test (WHERE matches nothing)
-- Backend should return status "no_match", not create approval.
-- ============================================================

UPDATE sample_app.customers
SET metadata = metadata || '{"flag":"test"}'
WHERE email = 'ghost@doesnotexist.com';

UPDATE sample_app.orders
SET notes = 'Flagged for review'
WHERE order_number = 'ORD-9999-999';


-- ============================================================
-- 12. DELETE — requires approval, affects the test row
-- ============================================================

DELETE FROM sample_app.customers
WHERE email = 'test.delete@example.com';


DELETE FROM sample_app.event_logs
WHERE event_type = 'test_event';


-- ============================================================
-- 13. DELETE — no_match test (WHERE matches nothing)
-- Backend should return status "no_match".
-- ============================================================

DELETE FROM sample_app.customers WHERE id = 99999;

DELETE FROM sample_app.orders WHERE order_number = 'ORD-0000-000';


-- ============================================================
-- 14. MULTI-QUERY — multiple statements separated by semicolons
-- Paste the block below as a single run in the query editor.
-- ============================================================

SELECT COUNT(*) AS total_customers FROM sample_app.customers;
SELECT COUNT(*) AS total_orders    FROM sample_app.orders;
SELECT COUNT(*) AS total_products  FROM sample_app.products;


-- Multi-query with write statements (each goes through approval)
UPDATE sample_app.products SET updated_at = NOW() WHERE sku = 'ELEC-LAP-001';
UPDATE sample_app.products SET updated_at = NOW() WHERE sku = 'ELEC-MOU-001';


-- ============================================================
-- 15. COMPLEX JSON BUILD — constructing JSON in SELECT
-- ============================================================

SELECT
    c.id,
    c.email,
    jsonb_build_object(
        'name',    c.first_name || ' ' || c.last_name,
        'tier',    c.metadata->>'tier',
        'address', c.address,
        'orders',  jsonb_agg(
            jsonb_build_object(
                'order_number', o.order_number,
                'total',        o.total_amount,
                'status',       o.status
            )
        )
    ) AS customer_profile
FROM sample_app.customers c
LEFT JOIN sample_app.orders o ON o.customer_id = c.id
GROUP BY c.id, c.email, c.first_name, c.last_name, c.metadata, c.address
ORDER BY c.id;
