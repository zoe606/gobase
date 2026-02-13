package app

import (
	"errors"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/hasher"
	"go-boilerplate/pkg/logger"
)

// defaultPermissions defines the permissions to seed.
//
//nolint:gochecknoglobals // seed data configuration
var defaultPermissions = []entity.Permission{
	// Translation permissions
	{Name: "translation:read", Resource: "translation", Action: "read"},
	{Name: "translation:write", Resource: "translation", Action: "write"},
	{Name: "translation:delete", Resource: "translation", Action: "delete"},
	// User management permissions
	{Name: "users:read", Resource: "users", Action: "read"},
	{Name: "users:write", Resource: "users", Action: "write"},
	{Name: "users:delete", Resource: "users", Action: "delete"},
	// Role management permissions
	{Name: "roles:read", Resource: "roles", Action: "read"},
	{Name: "roles:write", Resource: "roles", Action: "write"},
	{Name: "roles:delete", Resource: "roles", Action: "delete"},
	// Permission management
	{Name: "permissions:read", Resource: "permissions", Action: "read"},
	{Name: "permissions:write", Resource: "permissions", Action: "write"},
	{Name: "permissions:delete", Resource: "permissions", Action: "delete"},
	// Bank statement permissions
	{Name: "bank-statement:read", Resource: "bank-statement", Action: "read"},
	{Name: "bank-statement:write", Resource: "bank-statement", Action: "write"},
	{Name: "bank-statement:delete", Resource: "bank-statement", Action: "delete"},
	// Installment permissions
	{Name: "installment:read", Resource: "installment", Action: "read"},
	{Name: "installment:write", Resource: "installment", Action: "write"},
	{Name: "installment:delete", Resource: "installment", Action: "delete"},
}

// allPermissionNames returns all permission names for superadmin.
func allPermissionNames() []string {
	names := make([]string, len(defaultPermissions))
	for i, p := range defaultPermissions {
		names[i] = p.Name
	}

	return names
}

// defaultRoles defines the roles to seed with their permission assignments.
//
//nolint:gochecknoglobals // seed data configuration
var defaultRoles = []struct {
	Role        entity.Role
	Permissions []string // permission names to assign
}{
	{
		Role:        entity.Role{Name: "superadmin", Description: "Super administrator with all permissions"},
		Permissions: nil, // Will be set to all permissions in seedRoles
	},
	{
		Role: entity.Role{Name: "admin", Description: "Full system access"},
		Permissions: []string{
			"translation:read", "translation:write", "translation:delete",
			"users:read", "users:write", "users:delete",
			"roles:read", "roles:write", "roles:delete",
			"permissions:read", "permissions:write", "permissions:delete",
			"bank-statement:read", "bank-statement:write", "bank-statement:delete",
			"installment:read", "installment:write", "installment:delete",
		},
	},
	{
		Role: entity.Role{Name: "user", Description: "Standard user access"},
		Permissions: []string{
			"translation:read", "translation:write",
			"users:read",
			"roles:read",
		},
	},
	{
		Role: entity.Role{Name: "viewer", Description: "Read-only access"},
		Permissions: []string{
			"translation:read",
			"users:read",
			"roles:read",
		},
	},
}

// defaultUsers defines the users to seed.
//
//nolint:gochecknoglobals // seed data configuration
var defaultUsers = []struct {
	Email    string
	Password string
	Name     string
	RoleName string
}{
	{
		Email:    "superadmin@example.com",
		Password: "superadmin123",
		Name:     "Super Admin",
		RoleName: "superadmin",
	},
	{
		Email:    "admin@example.com",
		Password: "admin123",
		Name:     "Admin User",
		RoleName: "admin",
	},
}

// runSeeder seeds default data in development mode.
func runSeeder(db *gorm.DB, l *logger.Logger) {
	l.Info("Running database seeder (development mode)")

	seedPermissions(db, l)
	seedRoles(db, l)
	seedUsers(db, l)

	l.Info("Database seeding completed")
}

// seedPermissions creates default permissions if they don't exist.
func seedPermissions(db *gorm.DB, l *logger.Logger) {
	for i := range defaultPermissions {
		perm := &defaultPermissions[i]

		var existing entity.Permission

		result := db.Where("name = ?", perm.Name).First(&existing)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(perm).Error; err != nil {
				l.Error("Failed to seed permission %s: %v", perm.Name, err)
			}
		}
	}
}

// seedRoles creates default roles and assigns permissions.
func seedRoles(db *gorm.DB, l *logger.Logger) {
	for i := range defaultRoles {
		r := &defaultRoles[i]

		// Superadmin gets all permissions
		perms := r.Permissions
		if r.Role.Name == "superadmin" {
			perms = allPermissionNames()
		}

		var existing entity.Role

		result := db.Where("name = ?", r.Role.Name).First(&existing)

		var role entity.Role
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new role
			role = r.Role
			if err := db.Create(&role).Error; err != nil {
				l.Error("Failed to seed role %s: %v", r.Role.Name, err)

				continue
			}

			l.Info("Created role: %s", role.Name)
		} else {
			role = existing
		}

		// Assign permissions to role
		assignPermissionsToRole(db, &role, perms, l)
	}
}

// assignPermissionsToRole links permissions to a role.
func assignPermissionsToRole(db *gorm.DB, role *entity.Role, permNames []string, l *logger.Logger) {
	var permissions []entity.Permission

	if err := db.Where("name IN ?", permNames).Find(&permissions).Error; err != nil {
		l.Error("Failed to find permissions for role %s: %v", role.Name, err)

		return
	}

	// Use Association to replace permissions (handles duplicates automatically)
	if err := db.Model(role).Association("Permissions").Replace(permissions); err != nil {
		l.Error("Failed to assign permissions to role %s: %v", role.Name, err)
	}
}

// seedUsers creates default users if they don't exist.
func seedUsers(db *gorm.DB, l *logger.Logger) {
	for _, u := range defaultUsers {
		var existing entity.User

		result := db.Where("email = ?", u.Email).First(&existing)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Find role
			var role entity.Role
			if err := db.Where("name = ?", u.RoleName).First(&role).Error; err != nil {
				l.Error("Failed to find role %s for user %s: %v", u.RoleName, u.Email, err)

				continue
			}

			// Hash password
			passwordHash, err := hasher.Hash(u.Password)
			if err != nil {
				l.Error("Failed to hash password for user %s: %v", u.Email, err)

				continue
			}

			// Create user
			user := entity.User{
				Email:    u.Email,
				Password: passwordHash,
				Name:     u.Name,
				RoleID:   role.ID,
				Active:   true,
			}

			if err := db.Create(&user).Error; err != nil {
				l.Error("Failed to seed user %s: %v", u.Email, err)

				continue
			}

			l.Info("Created user: %s (role: %s)", user.Email, u.RoleName)
		}
	}
}
