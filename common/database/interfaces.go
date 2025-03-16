package database

import "gorm.io/gorm"

// DBProvider defines an interface for database access
type DBProvider interface {
	GetDB() *gorm.DB
}

// DefaultDBProvider uses the global DB variable
type DefaultDBProvider struct{}

// GetDB returns the global DB variable
func (p *DefaultDBProvider) GetDB() *gorm.DB {
	return DB
}

// Provider - global provider that can be replaced in tests
var Provider DBProvider = &DefaultDBProvider{}
