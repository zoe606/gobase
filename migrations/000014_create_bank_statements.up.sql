-- Create bank_statements table for uploaded statement metadata
CREATE TABLE IF NOT EXISTS bank_statements (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    bank_id INTEGER NOT NULL REFERENCES banks(id),
    media_id INTEGER REFERENCES media(id),
    period_start DATE,
    period_end DATE,
    password VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    error_message TEXT,
    total_debit NUMERIC(15,2) DEFAULT 0,
    total_credit NUMERIC(15,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_bank_statements_user_id ON bank_statements(user_id);
CREATE INDEX idx_bank_statements_bank_id ON bank_statements(bank_id);
CREATE INDEX idx_bank_statements_status ON bank_statements(status);
CREATE INDEX idx_bank_statements_deleted_at ON bank_statements(deleted_at);
