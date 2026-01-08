-- Remove seeded data
DELETE FROM role_permissions;
DELETE FROM roles WHERE name IN ('admin', 'user', 'viewer');
DELETE FROM permissions WHERE name IN (
    'translation:read',
    'translation:write',
    'translation:delete',
    'user:read',
    'user:write',
    'user:delete'
);
