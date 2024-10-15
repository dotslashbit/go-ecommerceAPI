-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    categories TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on categories for faster searches
CREATE INDEX idx_products_categories ON products USING GIN (categories);

-- Create index on name for faster text searches
CREATE INDEX idx_products_name ON products USING GIN (to_tsvector('english', name));

-- Create index on description for faster text searches
CREATE INDEX idx_products_description ON products USING GIN (to_tsvector('english', description));
