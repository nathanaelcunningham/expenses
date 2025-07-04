-- Categories for organizing expenses
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT, -- Hex color for UI
    icon TEXT, -- Icon identifier
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Family members (denormalized from master DB for performance)
CREATE TABLE IF NOT EXISTS family_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT, -- Same as user ID in master DB
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('manager', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Expenses table - core expense tracking
CREATE TABLE IF NOT EXISTS expenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER REFERENCES categories(id),
    amount DECIMAL(10, 2) NOT NULL,
    name TEXT NOT NULL,
    day_of_month_due INTEGER NOT NULL,
    is_autopay BOOLEAN  NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Default categories
INSERT OR IGNORE INTO categories (name, description, color, icon) VALUES
('Food & Dining', 'Groceries, restaurants, takeout', '#FF6B6B', 'utensils'),
('Transportation', 'Gas, public transit, rideshare', '#4ECDC4', 'car'),
('Utilities', 'Electricity, water, internet, phone', '#45B7D1', 'bolt'),
('Entertainment', 'Movies, games, subscriptions', '#96CEB4', 'play'),
('Shopping', 'Clothing, household items', '#FFEAA7', 'shopping-bag'),
('Healthcare', 'Medical, dental, prescriptions', '#DDA0DD', 'heart'),
('Education', 'Books, courses, school supplies', '#98D8C8', 'book'),
('Other', 'Miscellaneous expenses', '#A8A8A8', 'more-horizontal');

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_expenses_category_created_at ON expenses(category_id, created_at);
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at);

-- Triggers to update updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_categories_timestamp
    AFTER UPDATE ON categories
    FOR EACH ROW
BEGIN
    UPDATE categories SET updated_at = CURRENT_TIMESTAMP WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS update_expenses_timestamp
    AFTER UPDATE ON expenses
    FOR EACH ROW
BEGIN
    UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = new.id;
END;
