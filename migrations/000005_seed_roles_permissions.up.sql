-- Seed permissions
INSERT INTO permissions (name, resource, action) VALUES
    ('translation:read', 'translation', 'read'),
    ('translation:write', 'translation', 'write'),
    ('translation:delete', 'translation', 'delete'),
    ('user:read', 'user', 'read'),
    ('user:write', 'user', 'write'),
    ('user:delete', 'user', 'delete')
ON CONFLICT (name) DO NOTHING;

-- Seed roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('user', 'Standard user access'),
    ('viewer', 'Read-only access')
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to admin role (all permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- Assign permissions to user role (translation read/write)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('translation:read', 'translation:write')
ON CONFLICT DO NOTHING;

-- Assign permissions to viewer role (translation read only)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'translation:read'
ON CONFLICT DO NOTHING;
