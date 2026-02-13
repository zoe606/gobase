-- Seed bank statement and installment permissions
INSERT INTO permissions (name, resource, action) VALUES
    ('bank-statement:read', 'bank-statement', 'read'),
    ('bank-statement:write', 'bank-statement', 'write'),
    ('bank-statement:delete', 'bank-statement', 'delete'),
    ('installment:read', 'installment', 'read'),
    ('installment:write', 'installment', 'write'),
    ('installment:delete', 'installment', 'delete')
ON CONFLICT (name) DO NOTHING;

-- Assign all new permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN (
    'bank-statement:read',
    'bank-statement:write',
    'bank-statement:delete',
    'installment:read',
    'installment:write',
    'installment:delete'
)
ON CONFLICT DO NOTHING;
