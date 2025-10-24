-- ============================================
-- Restaurant Management System - Seed Data
-- Database: PostgreSQL
-- ============================================

-- Insert Restaurants
INSERT INTO restaurants (name, address, phone, email, opening_time, closing_time) VALUES
('The Garden Bistro', '123 Main Street, Downtown', '+1-555-0101', 'info@gardenbistro.com', '11:00:00', '22:00:00'),
('Spice Junction', '456 Oak Avenue, Midtown', '+1-555-0102', 'contact@spicejunction.com', '12:00:00', '23:00:00'),
('Ocean Breeze Cafe', '789 Beach Road, Coastal Area', '+1-555-0103', 'hello@oceanbreeze.com', '09:00:00', '21:00:00');

-- Insert Menu Items for The Garden Bistro (restaurant_id = 1)
INSERT INTO menu_items (restaurant_id, name, description, price, category, available) VALUES
(1, 'Caesar Salad', 'Fresh romaine lettuce with parmesan and croutons', 12.99, 'Appetizer', true),
(1, 'Tomato Soup', 'Creamy tomato soup with basil', 8.99, 'Appetizer', true),
(1, 'Grilled Salmon', 'Atlantic salmon with herbs and lemon butter', 24.99, 'Main Course', true),
(1, 'Beef Tenderloin', 'Prime beef with mushroom sauce', 32.99, 'Main Course', true),
(1, 'Vegetarian Pasta', 'Penne with seasonal vegetables', 16.99, 'Main Course', true),
(1, 'Chocolate Lava Cake', 'Warm chocolate cake with vanilla ice cream', 9.99, 'Dessert', true),
(1, 'Tiramisu', 'Classic Italian dessert', 8.99, 'Dessert', true),
(1, 'House Wine', 'Red or white wine glass', 11.00, 'Beverage', true),
(1, 'Fresh Lemonade', 'Homemade lemonade', 4.99, 'Beverage', true);

-- Insert Menu Items for Spice Junction (restaurant_id = 2)
INSERT INTO menu_items (restaurant_id, name, description, price, category, available) VALUES
(2, 'Samosa', 'Crispy pastry with spiced potatoes', 6.99, 'Appetizer', true),
(2, 'Chicken Tikka', 'Grilled chicken with Indian spices', 14.99, 'Appetizer', true),
(2, 'Butter Chicken', 'Creamy tomato curry with tender chicken', 18.99, 'Main Course', true),
(2, 'Lamb Vindaloo', 'Spicy lamb curry', 21.99, 'Main Course', true),
(2, 'Paneer Tikka Masala', 'Cottage cheese in rich tomato gravy', 16.99, 'Main Course', true),
(2, 'Biryani', 'Fragrant rice with chicken or vegetables', 19.99, 'Main Course', true),
(2, 'Garlic Naan', 'Indian bread with garlic', 3.99, 'Side', true),
(2, 'Mango Lassi', 'Sweet yogurt drink with mango', 5.99, 'Beverage', true),
(2, 'Gulab Jamun', 'Sweet milk dumplings in syrup', 6.99, 'Dessert', true);

-- Insert Menu Items for Ocean Breeze Cafe (restaurant_id = 3)
INSERT INTO menu_items (restaurant_id, name, description, price, category, available) VALUES
(3, 'Shrimp Cocktail', 'Chilled shrimp with cocktail sauce', 13.99, 'Appetizer', true),
(3, 'Clam Chowder', 'Creamy New England style soup', 9.99, 'Appetizer', true),
(3, 'Fish and Chips', 'Beer-battered cod with fries', 17.99, 'Main Course', true),
(3, 'Lobster Roll', 'Fresh lobster in toasted roll', 26.99, 'Main Course', true),
(3, 'Grilled Tuna Steak', 'Seared tuna with wasabi mayo', 28.99, 'Main Course', true),
(3, 'Seafood Paella', 'Spanish rice with mixed seafood', 31.99, 'Main Course', true),
(3, 'Key Lime Pie', 'Tangy and sweet Florida classic', 7.99, 'Dessert', true),
(3, 'Iced Coffee', 'Cold brew coffee', 4.99, 'Beverage', true),
(3, 'Tropical Smoothie', 'Blend of tropical fruits', 6.99, 'Beverage', true);

-- Insert Tables for The Garden Bistro
INSERT INTO tables (restaurant_id, table_number, capacity, status) VALUES
(1, 'T1', 2, 'available'),
(1, 'T2', 2, 'available'),
(1, 'T3', 4, 'available'),
(1, 'T4', 4, 'occupied'),
(1, 'T5', 6, 'available'),
(1, 'T6', 8, 'reserved');

-- Insert Tables for Spice Junction
INSERT INTO tables (restaurant_id, table_number, capacity, status) VALUES
(2, 'A1', 2, 'available'),
(2, 'A2', 4, 'occupied'),
(2, 'A3', 4, 'available'),
(2, 'B1', 6, 'available'),
(2, 'B2', 8, 'available');

-- Insert Tables for Ocean Breeze Cafe
INSERT INTO tables (restaurant_id, table_number, capacity, status) VALUES
(3, '101', 2, 'available'),
(3, '102', 2, 'occupied'),
(3, '103', 4, 'available'),
(3, '104', 4, 'available'),
(3, '105', 6, 'reserved');

-- Insert Customers
INSERT INTO customers (name, phone, email) VALUES
('John Smith', '+1-555-1001', 'john.smith@email.com'),
('Sarah Johnson', '+1-555-1002', 'sarah.j@email.com'),
('Michael Brown', '+1-555-1003', 'mbrown@email.com'),
('Emily Davis', '+1-555-1004', 'emily.davis@email.com'),
('David Wilson', '+1-555-1005', 'dwilson@email.com'),
('Lisa Anderson', '+1-555-1006', 'landerson@email.com');

-- Insert Staff for The Garden Bistro
INSERT INTO staff (restaurant_id, name, role, phone, email) VALUES
(1, 'James Miller', 'Manager', '+1-555-2001', 'james.m@gardenbistro.com'),
(1, 'Maria Garcia', 'Head Chef', '+1-555-2002', 'maria.g@gardenbistro.com'),
(1, 'Tom Harris', 'Waiter', '+1-555-2003', 'tom.h@gardenbistro.com'),
(1, 'Anna Lee', 'Waitress', '+1-555-2004', 'anna.l@gardenbistro.com');

-- Insert Staff for Spice Junction
INSERT INTO staff (restaurant_id, name, role, phone, email) VALUES
(2, 'Raj Patel', 'Manager', '+1-555-2005', 'raj.p@spicejunction.com'),
(2, 'Priya Sharma', 'Head Chef', '+1-555-2006', 'priya.s@spicejunction.com'),
(2, 'Kevin Wong', 'Waiter', '+1-555-2007', 'kevin.w@spicejunction.com');

-- Insert Staff for Ocean Breeze Cafe
INSERT INTO staff (restaurant_id, name, role, phone, email) VALUES
(3, 'Captain Jack Roberts', 'Manager', '+1-555-2008', 'jack.r@oceanbreeze.com'),
(3, 'Sophie Martinez', 'Head Chef', '+1-555-2009', 'sophie.m@oceanbreeze.com'),
(3, 'Chris Taylor', 'Waiter', '+1-555-2010', 'chris.t@oceanbreeze.com');

-- Insert Sample Orders
-- Order 1: Completed order at The Garden Bistro
INSERT INTO orders (restaurant_id, table_id, customer_id, staff_id, total_amount, status, order_time, completed_time) VALUES
(1, 4, 1, 3, 58.96, 'completed', CURRENT_TIMESTAMP - INTERVAL '2 hours', CURRENT_TIMESTAMP - INTERVAL '1 hour');

INSERT INTO order_items (order_id, menu_item_id, quantity, unit_price, subtotal) VALUES
(1, 1, 1, 12.99, 12.99),  -- Caesar Salad
(1, 3, 2, 24.99, 49.98),  -- Grilled Salmon x2
(1, 6, 1, 9.99, 9.99);    -- Chocolate Lava Cake

-- Order 2: Active order at Spice Junction
INSERT INTO orders (restaurant_id, table_id, customer_id, staff_id, total_amount, status, order_time) VALUES
(2, 8, 2, 7, 52.95, 'preparing', CURRENT_TIMESTAMP - INTERVAL '30 minutes');

INSERT INTO order_items (order_id, menu_item_id, quantity, unit_price, subtotal, special_instructions) VALUES
(2, 11, 2, 6.99, 13.98, 'Extra crispy please'),  -- Samosa x2
(2, 13, 1, 18.99, 18.99, NULL),                  -- Butter Chicken
(2, 15, 1, 16.99, 16.99, 'Mild spice level'),   -- Paneer Tikka Masala
(2, 17, 1, 3.99, 3.99, NULL);                    -- Garlic Naan

-- Order 3: New order at Ocean Breeze Cafe
INSERT INTO orders (restaurant_id, table_id, customer_id, staff_id, total_amount, status, order_time) VALUES
(3, 12, 4, 10, 63.96, 'pending', CURRENT_TIMESTAMP - INTERVAL '5 minutes');

INSERT INTO order_items (order_id, menu_item_id, quantity, unit_price, subtotal) VALUES
(3, 20, 1, 13.99, 13.99),  -- Shrimp Cocktail
(3, 23, 1, 26.99, 26.99),  -- Lobster Roll
(3, 24, 1, 28.99, 28.99),  -- Grilled Tuna Steak
(3, 26, 1, 7.99, 7.99);    -- Key Lime Pie

-- Order 4: Another completed order at The Garden Bistro
INSERT INTO orders (restaurant_id, table_id, customer_id, staff_id, total_amount, status, order_time, completed_time) VALUES
(1, 3, 5, 4, 45.96, 'completed', CURRENT_TIMESTAMP - INTERVAL '3 hours', CURRENT_TIMESTAMP - INTERVAL '2 hours');

INSERT INTO order_items (order_id, menu_item_id, quantity, unit_price, subtotal) VALUES
(4, 2, 2, 8.99, 17.98),   -- Tomato Soup x2
(4, 5, 1, 16.99, 16.99),  -- Vegetarian Pasta
(4, 8, 1, 11.00, 11.00);  -- House Wine

-- Display summary
SELECT 'Seed data inserted successfully!' AS status;
SELECT 'Restaurants: ' || COUNT(*) AS count FROM restaurants;
SELECT 'Menu Items: ' || COUNT(*) AS count FROM menu_items;
SELECT 'Tables: ' || COUNT(*) AS count FROM tables;
SELECT 'Customers: ' || COUNT(*) AS count FROM customers;
SELECT 'Staff: ' || COUNT(*) AS count FROM staff;
SELECT 'Orders: ' || COUNT(*) AS count FROM orders;
SELECT 'Order Items: ' || COUNT(*) AS count FROM order_items;