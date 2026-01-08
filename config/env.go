package config

// Environment constants.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// IsDevelopment returns true if running in development environment.
func (a *App) IsDevelopment() bool {
	return a.Env == EnvDevelopment || a.Env == ""
}

// IsStaging returns true if running in staging environment.
func (a *App) IsStaging() bool {
	return a.Env == EnvStaging
}

// IsProduction returns true if running in production environment.
func (a *App) IsProduction() bool {
	return a.Env == EnvProduction
}

// IsLocal returns true if running in non-production environment.
func (a *App) IsLocal() bool {
	return !a.IsProduction()
}

// ShouldAutoMigrate returns true if AutoMigrate should be enabled.
// Only enabled in development environment.
func (a *App) ShouldAutoMigrate() bool {
	return a.IsDevelopment()
}
