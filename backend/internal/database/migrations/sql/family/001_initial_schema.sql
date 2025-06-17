-- Description: Create initial family database schema for expenses and categories
-- This defines the core expense tracking tables for each family's dedicated database

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
    member_id TEXT NOT NULL REFERENCES family_members(id),
    category_id TEXT REFERENCES categories(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency TEXT DEFAULT 'USD',
    description TEXT NOT NULL,
    date DATE NOT NULL,
    receipt_url TEXT, -- URL to receipt image if uploaded
    tags TEXT, -- JSON array of tags
    is_recurring BOOLEAN DEFAULT FALSE,
    recurring_interval TEXT, -- 'monthly', 'weekly', etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Expense splits for shared expenses
CREATE TABLE IF NOT EXISTS expense_splits (
    id TEXT PRIMARY KEY,
    expense_id TEXT NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    member_id TEXT NOT NULL REFERENCES family_members(id),
    amount DECIMAL(10, 2) NOT NULL,
    percentage DECIMAL(5, 2), -- For percentage-based splits
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(expense_id, member_id)
);

-- Budgets for expense tracking
CREATE TABLE IF NOT EXISTS budgets (
    id TEXT PRIMARY KEY,
    category_id TEXT REFERENCES categories(id),
    member_id TEXT REFERENCES family_members(id), -- NULL for family-wide budgets
    amount DECIMAL(10, 2) NOT NULL,
    period TEXT NOT NULL CHECK (period IN ('monthly', 'weekly', 'yearly')),
    start_date DATE NOT NULL,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recurring expenses templates
CREATE TABLE IF NOT EXISTS recurring_expenses (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES family_members(id),
    category_id TEXT REFERENCES categories(id),
    amount DECIMAL(10, 2) NOT NULL,
    description TEXT NOT NULL,
    interval_type TEXT NOT NULL CHECK (interval_type IN ('weekly', 'monthly', 'yearly')),
    interval_value INTEGER DEFAULT 1, -- Every N intervals
    next_due_date DATE NOT NULL,
    last_generated_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
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
CREATE INDEX IF NOT EXISTS idx_expenses_member_date ON expenses(member_id, date);
CREATE INDEX IF NOT EXISTS idx_expenses_category_date ON expenses(category_id, date);
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(date);
CREATE INDEX IF NOT EXISTS idx_expense_splits_expense ON expense_splits(expense_id);
CREATE INDEX IF NOT EXISTS idx_expense_splits_member ON expense_splits(member_id);
CREATE INDEX IF NOT EXISTS idx_budgets_category ON budgets(category_id);
CREATE INDEX IF NOT EXISTS idx_budgets_member ON budgets(member_id);
CREATE INDEX IF NOT EXISTS idx_recurring_expenses_member ON recurring_expenses(member_id);
CREATE INDEX IF NOT EXISTS idx_recurring_expenses_due_date ON recurring_expenses(next_due_date);

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

CREATE TRIGGER IF NOT EXISTS update_budgets_timestamp
    AFTER UPDATE ON budgets
    FOR EACH ROW
BEGIN
    UPDATE budgets SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_recurring_expenses_timestamp
    AFTER UPDATE ON recurring_expenses
    FOR EACH ROW
BEGIN
    UPDATE recurring_expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;