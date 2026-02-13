-- Remove role_permissions for bank statement permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN (
        'bank-statement:read',
        'bank-statement:write',
        'bank-statement:delete',
        'installment:read',
        'installment:write',
        'installment:delete'
    )
);

-- Remove bank statement permissions
DELETE FROM permissions WHERE name IN (
    'bank-statement:read',
    'bank-statement:write',
    'bank-statement:delete',
    'installment:read',
    'installment:write',
    'installment:delete'
);
