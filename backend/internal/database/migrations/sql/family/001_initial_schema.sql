-- Categories for organizing expenses
CREATE TABLE IF NOT EXISTS categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT, -- Hex color for UI
    icon TEXT, -- Icon identifier
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Family members (denormalized from master DB for performance)
CREATE TABLE IF NOT EXISTS family_members (
    id TEXT PRIMARY KEY, -- Same as user ID in master DB
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('manager', 'member')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Expenses table - core expense tracking
CREATE TABLE IF NOT EXISTS expenses (
    id TEXT PRIMARY KEY,
    category_id TEXT REFERENCES categories(id),
    amount DECIMAL(10, 2) NOT NULL,
    name TEXT NOT NULL,
    day_of_month_due INTEGER NOT NULL,
    is_autopay BOOLEAN  NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Default categories
INSERT OR IGNORE INTO categories (id, name, description, color, icon) VALUES
('cat_food', 'Food & Dining', 'Groceries, restaurants, takeout', '#FF6B6B', 'utensils'),
('cat_transport', 'Transportation', 'Gas, public transit, rideshare', '#4ECDC4', 'car'),
('cat_utilities', 'Utilities', 'Electricity, water, internet, phone', '#45B7D1', 'bolt'),
('cat_entertainment', 'Entertainment', 'Movies, games, subscriptions', '#96CEB4', 'play'),
('cat_shopping', 'Shopping', 'Clothing, household items', '#FFEAA7', 'shopping-bag'),
('cat_health', 'Healthcare', 'Medical, dental, prescriptions', '#DDA0DD', 'heart'),
('cat_education', 'Education', 'Books, courses, school supplies', '#98D8C8', 'book'),
('cat_other', 'Other', 'Miscellaneous expenses', '#A8A8A8', 'more-horizontal');

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_expenses_category_created_at ON expenses(category_id, created_at);
CREATE INDEX IF NOT EXISTS idx_expenses_created_at ON expenses(created_at);

-- Triggers to update updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_categories_timestamp
    AFTER UPDATE ON categories
    FOR EACH ROW
BEGIN
    UPDATE categories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_expenses_timestamp
    AFTER UPDATE ON expenses
    FOR EACH ROW
BEGIN
    UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
