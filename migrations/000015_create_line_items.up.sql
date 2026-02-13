-- Create line_items table (generic polymorphic line items)
CREATE TABLE IF NOT EXISTS line_items (
    id SERIAL PRIMARY KEY,
    source_type VARCHAR(100) NOT NULL,
    source_id INTEGER NOT NULL,
    date DATE NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100),
    debit NUMERIC(15,2) DEFAULT 0,
    credit NUMERIC(15,2) DEFAULT 0,
    balance NUMERIC(15,2) DEFAULT 0,
    is_installment BOOLEAN DEFAULT FALSE,
    installment_id INTEGER REFERENCES installments(id) ON DELETE SET NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_line_items_source ON line_items(source_type, source_id);
CREATE INDEX idx_line_items_installment ON line_items(installment_id);
CREATE INDEX idx_line_items_date ON line_items(date);
