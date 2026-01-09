package app

import (
	"errors"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/logger"
)

// defaultPermissions defines the permissions to seed.
//
//nolint:gochecknoglobals // seed data configuration
var defaultPermissions = []entity.Permission{
	{Name: "translation:read", Resource: "translation", Action: "read"},
	{Name: "translation:write", Resource: "translation", Action: "write"},
	{Name: "translation:delete", Resource: "translation", Action: "delete"},
	{Name: "user:read", Resource: "user", Action: "read"},
	{Name: "user:write", Resource: "user", Action: "write"},
	{Name: "user:delete", Resource: "user", Action: "delete"},
}

// defaultRoles defines the roles to seed with their permission assignments.
//
//nolint:gochecknoglobals // seed data configuration
var defaultRoles = []struct {
	Role        entity.Role
	Permissions []string // permission names to assign
}{
	{
		Role:        entity.Role{Name: "admin", Description: "Full system access"},
		Permissions: []string{"translation:read", "translation:write", "translation:delete", "user:read", "user:write", "user:delete"},
	},
	{
		Role:        entity.Role{Name: "user", Description: "Standard user access"},
		Permissions: []string{"translation:read", "translation:write"},
	},
	{
		Role:        entity.Role{Name: "viewer", Description: "Read-only access"},
		Permissions: []string{"translation:read"},
	},
}

// runSeeder seeds default data in development mode.
func runSeeder(db *gorm.DB, l *logger.Logger) {
	l.Info("Running database seeder (development mode)")

	seedPermissions(db, l)
	seedRoles(db, l)

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
		} else {
			role = existing
		}

		// Assign permissions to role
		assignPermissionsToRole(db, &role, r.Permissions, l)
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
