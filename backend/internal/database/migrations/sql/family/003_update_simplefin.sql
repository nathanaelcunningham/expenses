DROP TABLE simplefin;

CREATE TABLE IF NOT EXISTS family_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    setting_key TEXT UNIQUE NOT NULL,
    setting_value TEXT,
    data_type TEXT NOT NULL
);

