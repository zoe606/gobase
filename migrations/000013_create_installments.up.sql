-- Create installments table for tracking recurring payments
CREATE TABLE IF NOT EXISTS installments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    merchant VARCHAR(255),
    total_amount NUMERIC(15,2) NOT NULL,
    monthly_amount NUMERIC(15,2) NOT NULL,
    total_terms INTEGER NOT NULL,
    completed_terms INTEGER DEFAULT 0,
    start_date DATE,
    end_date DATE,
    status VARCHAR(20) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_installments_user_id ON installments(user_id);
CREATE INDEX idx_installments_status ON installments(status);
CREATE INDEX idx_installments_deleted_at ON installments(deleted_at);
