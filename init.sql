-- ============================================================================
-- SCHÉMA DE BASE DE DONNÉES NORMALISÉ (3NF+)
-- Projet d'évaluation : Comparaison V1 (non optimisé) vs V2 (optimisé)
-- ============================================================================

-- ============================================================================
-- 1. TABLE CATEGORIES - Catégories de produits
-- ============================================================================
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_categories_name ON categories(name);

-- ============================================================================
-- 2. TABLE SUPPLIERS - Fournisseurs
-- ============================================================================
CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    contact_name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_suppliers_name ON suppliers(name);

-- ============================================================================
-- 3. TABLE PRODUCTS - Produits
-- ============================================================================
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    base_price NUMERIC(10, 2) NOT NULL CHECK (base_price > 0),
    stock_quantity INTEGER DEFAULT 0 CHECK (stock_quantity >= 0),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_supplier ON products(supplier_id);
CREATE INDEX idx_products_price ON products(base_price);

-- ============================================================================
-- 4. TABLE PRODUCT_CATEGORIES - Relation N-N Produits ↔ Catégories
-- ============================================================================
CREATE TABLE IF NOT EXISTS product_categories (
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

CREATE INDEX idx_product_categories_product ON product_categories(product_id);
CREATE INDEX idx_product_categories_category ON product_categories(category_id);

-- ============================================================================
-- 5. TABLE CUSTOMERS - Clients
-- ============================================================================
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE,
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'France',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_name ON customers(last_name, first_name);

-- ============================================================================
-- 6. TABLE STORES - Magasins / Points de vente
-- ============================================================================
CREATE TABLE IF NOT EXISTS stores (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    city VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    country VARCHAR(100) DEFAULT 'France',
    address TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_stores_city ON stores(city);
CREATE INDEX idx_stores_region ON stores(region);

-- ============================================================================
-- 7. TABLE PAYMENT_METHODS - Méthodes de paiement
-- ============================================================================
CREATE TABLE IF NOT EXISTS payment_methods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_payment_methods_active ON payment_methods(active);

-- ============================================================================
-- 8. TABLE PROMOTIONS - Promotions / Remises
-- ============================================================================
CREATE TABLE IF NOT EXISTS promotions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    discount_percent NUMERIC(5, 2) CHECK (discount_percent >= 0 AND discount_percent <= 100),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    CHECK (end_date >= start_date)
);

CREATE INDEX idx_promotions_code ON promotions(code);
CREATE INDEX idx_promotions_dates ON promotions(start_date, end_date);
CREATE INDEX idx_promotions_active ON promotions(active);

-- ============================================================================
-- 9. TABLE ORDERS - Commandes (Header)
-- ============================================================================
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    store_id INTEGER NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    payment_method_id INTEGER NOT NULL REFERENCES payment_methods(id),
    promotion_id INTEGER REFERENCES promotions(id) ON DELETE SET NULL,
    order_date DATE NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_store ON orders(store_id);
CREATE INDEX idx_orders_date ON orders(order_date DESC);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_promotion ON orders(promotion_id);

-- ============================================================================
-- 10. TABLE ORDER_ITEMS - Lignes de commande (Détails)
-- ============================================================================
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10, 2) NOT NULL CHECK (unit_price >= 0),
    subtotal NUMERIC(12, 2) NOT NULL CHECK (subtotal >= 0),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);

-- ============================================================================
-- VUES UTILES POUR L'ANALYSE
-- ============================================================================

-- Vue : Ventes complètes avec toutes les informations
CREATE OR REPLACE VIEW v_sales_complete AS
SELECT
    oi.id as sale_id,
    o.order_date,
    o.id as order_id,
    c.id as customer_id,
    c.first_name || ' ' || c.last_name as customer_name,
    c.email as customer_email,
    p.id as product_id,
    p.name as product_name,
    s.name as store_name,
    s.city as store_city,
    pm.name as payment_method,
    pr.code as promotion_code,
    pr.discount_percent,
    oi.quantity,
    oi.unit_price,
    oi.subtotal,
    o.total_amount as order_total
FROM order_items oi
INNER JOIN orders o ON oi.order_id = o.id
INNER JOIN customers c ON o.customer_id = c.id
INNER JOIN products p ON oi.product_id = p.id
INNER JOIN stores s ON o.store_id = s.id
INNER JOIN payment_methods pm ON o.payment_method_id = pm.id
LEFT JOIN promotions pr ON o.promotion_id = pr.id;

-- Vue : Statistiques par catégorie
CREATE OR REPLACE VIEW v_stats_by_category AS
SELECT
    cat.id as category_id,
    cat.name as category_name,
    COUNT(DISTINCT oi.id) as nb_ventes,
    SUM(oi.quantity) as total_quantity,
    SUM(oi.subtotal) as total_ca,
    AVG(oi.subtotal) as avg_sale
FROM categories cat
INNER JOIN product_categories pc ON cat.id = pc.category_id
INNER JOIN products p ON pc.product_id = p.id
INNER JOIN order_items oi ON p.id = oi.product_id
GROUP BY cat.id, cat.name;

-- ============================================================================
-- DONNÉES INITIALES (RÉFÉRENTIELLES)
-- ============================================================================

-- Méthodes de paiement
INSERT INTO payment_methods (name, description) VALUES
    ('Carte Bancaire', 'Paiement par carte bancaire'),
    ('Espèces', 'Paiement en espèces'),
    ('Virement', 'Virement bancaire'),
    ('PayPal', 'Paiement via PayPal'),
    ('Chèque', 'Paiement par chèque')
ON CONFLICT (name) DO NOTHING;

-- Catégories
INSERT INTO categories (name, description) VALUES
    ('Électronique', 'Appareils électroniques et high-tech'),
    ('Vêtements', 'Vêtements et accessoires de mode'),
    ('Alimentation', 'Produits alimentaires et boissons'),
    ('Maison', 'Articles pour la maison et décoration'),
    ('Sport', 'Équipements et vêtements de sport'),
    ('Livres', 'Livres et publications'),
    ('Jouets', 'Jouets et jeux pour enfants'),
    ('Beauté', 'Produits de beauté et cosmétiques')
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- ANALYSE ET OPTIMISATION
-- ============================================================================
ANALYZE categories;
ANALYZE suppliers;
ANALYZE products;
ANALYZE product_categories;
ANALYZE customers;
ANALYZE stores;
ANALYZE payment_methods;
ANALYZE promotions;
ANALYZE orders;
ANALYZE order_items;
