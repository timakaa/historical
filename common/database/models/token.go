package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Token represents an authentication token
type Token struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TokenString     string    `json:"tokenString" gorm:"uniqueIndex"`
	ExpiresAt       time.Time `json:"expiresAt"`
	CandlesLeft     int64
	Permissions     []string  `json:"permissions" gorm:"-"` // Stored as JSON in PermissionsJSON
	PermissionsJSON string    `json:"-" gorm:"column:permissions"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName specifies the table name for the Token model
func (Token) TableName() string {
	return "tokens"
}

// NewToken creates a new token with the given parameters
func NewToken(permissions []string, expiresIn int64) *Token {
	now := time.Now()
	return &Token{
		TokenString: generateTokenString(),
		Permissions: permissions,
		CandlesLeft: 5000,
		ExpiresAt:   now.Add(time.Duration(expiresIn) * time.Second),
		CreatedAt:   now,
	}
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// TimeUntilExpiration returns the duration until the token expires
func (t *Token) TimeUntilExpiration() time.Duration {
	return time.Until(t.ExpiresAt)
}

// generateTokenString generates a unique token string
func generateTokenString() string {
	// Generate a UUID v4
	id, err := uuid.NewRandom()
	if err != nil {
		// Fallback to timestamp if UUID generation fails
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	return id.String()
}

// BeforeSave hook to handle JSON serialization of Permissions
func (t *Token) BeforeSave() error {
	// Convert permissions slice to JSON string
	if len(t.Permissions) > 0 {
		permissionsJSON, err := json.Marshal(t.Permissions)
		if err != nil {
			return err
		}
		t.PermissionsJSON = string(permissionsJSON)
	} else {
		t.PermissionsJSON = "[]"
	}
	return nil
}

// AfterFind hook to handle JSON deserialization of Permissions
func (t *Token) AfterFind() error {
	// Convert JSON string to permissions slice
	if t.PermissionsJSON != "" {
		return json.Unmarshal([]byte(t.PermissionsJSON), &t.Permissions)
	}
	return nil
}
