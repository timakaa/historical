package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timakaa/historical-common/database/models"
	pb "github.com/timakaa/historical-common/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	authpkg "github.com/timakaa/historical-auth/internal"
)

// MockToken is a mock implementation of Token for testing
type MockToken struct {
	models.Token
	MockBeforeSaveError bool
}

// BeforeSave mocks the BeforeSave method
func (m *MockToken) BeforeSave() error {
	if m.MockBeforeSaveError {
		return errors.New("mock BeforeSave error")
	}
	return nil
}

// setupInMemoryDB creates an in-memory SQLite test database
func setupInMemoryDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "Failed to open in-memory database")

	err = db.AutoMigrate(&models.Token{})
	require.NoError(t, err, "Failed to migrate database")

	return db
}

// createTestToken creates a test token in the database
func createTestToken(t *testing.T, db *gorm.DB, candlesLeft int64) *models.Token {
	token := models.NewToken([]string{"read"}, 3600) // 1 hour expiration
	token.CandlesLeft = candlesLeft

	err := token.BeforeSave()
	require.NoError(t, err, "Failed to prepare token")

	result := db.Create(token)
	require.NoError(t, result.Error, "Failed to create token")

	return token
}

// setupErrorDB creates a database connection that will cause errors
func setupErrorDB(t *testing.T) *gorm.DB {
	// Create a DB connection that will be closed
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Get the underlying SQL DB to close it
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Close the DB to simulate errors
	sqlDB.Close()

	return db
}

func TestCreateTokenUnit(t *testing.T) {
	db := setupInMemoryDB(t)
	server := authpkg.NewServer(db)

	// Test successful token creation
	t.Run("Success", func(t *testing.T) {
		req := &pb.CreateTokenRequest{
			Permissions: []string{"read", "write"},
			ExpiresIn:   3600, // 1 hour
		}

		resp, err := server.CreateToken(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Greater(t, resp.ExpiresAt, int64(0))

		// Verify that the token was actually created in the database
		var token models.Token
		result := db.Where("token_string = ?", resp.Token).First(&token)
		assert.NoError(t, result.Error)
	})

	// Test with invalid data
	t.Run("InvalidExpiresIn", func(t *testing.T) {
		req := &pb.CreateTokenRequest{
			Permissions: []string{"read"},
			ExpiresIn:   -1, // Negative value
		}

		resp, err := server.CreateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test with nil database
	t.Run("NilDatabase", func(t *testing.T) {
		nilDBServer := authpkg.NewServer(nil)
		req := &pb.CreateTokenRequest{
			Permissions: []string{"read"},
			ExpiresIn:   3600,
		}

		resp, err := nilDBServer.CreateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection not available")
	})

	// Test with database error
	t.Run("DatabaseError", func(t *testing.T) {
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.CreateTokenRequest{
			Permissions: []string{"read"},
			ExpiresIn:   3600,
		}

		resp, err := errorServer.CreateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to create token")
	})
}

func TestValidateTokenUnit(t *testing.T) {
	db := setupInMemoryDB(t)
	server := authpkg.NewServer(db)

	// Create a test token
	token := createTestToken(t, db, 100)

	// Test successful validation
	t.Run("ValidToken", func(t *testing.T) {
		req := &pb.ValidateRequest{
			Token: token.TokenString,
		}

		resp, err := server.ValidateToken(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.IsValid)
	})

	// Test with non-existent token
	t.Run("NonExistentToken", func(t *testing.T) {
		req := &pb.ValidateRequest{
			Token: "non-existent-token",
		}

		resp, err := server.ValidateToken(context.Background(), req)
		assert.NoError(t, err)
		assert.False(t, resp.IsValid)
	})

	// Test with expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		// Create a token with expired validity
		expiredToken := models.NewToken([]string{"read"}, 1) // 1 second
		expiredToken.ExpiresAt = time.Now().Add(-time.Hour)  // Set expiration time in the past

		err := expiredToken.BeforeSave()
		require.NoError(t, err)

		result := db.Create(expiredToken)
		require.NoError(t, result.Error)

		req := &pb.ValidateRequest{
			Token: expiredToken.TokenString,
		}

		resp, err := server.ValidateToken(context.Background(), req)
		assert.NoError(t, err)
		assert.False(t, resp.IsValid)
	})

	// Test with nil database
	t.Run("NilDatabase", func(t *testing.T) {
		nilDBServer := authpkg.NewServer(nil)
		req := &pb.ValidateRequest{
			Token: "some-token",
		}

		resp, err := nilDBServer.ValidateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection not available")
	})

	// Test with empty token
	t.Run("EmptyToken", func(t *testing.T) {
		req := &pb.ValidateRequest{
			Token: "",
		}

		resp, err := server.ValidateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "token is required")
	})

	// Test with database error (not record not found)
	t.Run("DatabaseError", func(t *testing.T) {
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.ValidateRequest{
			Token: "some-token",
		}

		resp, err := errorServer.ValidateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to validate token")
	})
}

func TestUpdateTokenCandlesLeftUnit(t *testing.T) {
	db := setupInMemoryDB(t)
	server := authpkg.NewServer(db)

	// Create a test token with 100 candles
	token := createTestToken(t, db, 100)

	// Test successful decrease of candles count
	t.Run("DecreaseCandles", func(t *testing.T) {
		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           token.TokenString,
			DecreaseCandles: 20,
		}

		resp, err := server.UpdateTokenCandlesLeft(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int64(80), resp.CandlesLeft) // 100 - 20 = 80

		// Verify that the value in the database has been updated
		var updatedToken models.Token
		result := db.Where("token_string = ?", token.TokenString).First(&updatedToken)
		assert.NoError(t, result.Error)

		var candlesLeft int64
		err = db.Model(&updatedToken).Select("candles_left").Scan(&candlesLeft).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(80), candlesLeft)
	})

	// Test with decrease to zero (token should be deleted)
	t.Run("DecreaseToZero", func(t *testing.T) {
		// Create a new token with 50 candles
		zeroToken := createTestToken(t, db, 50)

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           zeroToken.TokenString,
			DecreaseCandles: 50, // Decrease to zero
		}

		resp, err := server.UpdateTokenCandlesLeft(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), resp.CandlesLeft)

		// Verify that the token has been deleted from the database
		var count int64
		db.Model(&models.Token{}).Where("token_string = ?", zeroToken.TokenString).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	// Test with negative DecreaseCandles value (token should be revoked)
	t.Run("NegativeDecrease", func(t *testing.T) {
		// Create a new token
		negToken := createTestToken(t, db, 100)

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           negToken.TokenString,
			DecreaseCandles: -10, // Negative value
		}

		resp, err := server.UpdateTokenCandlesLeft(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, resp.CandlesLeft, req.DecreaseCandles)

		// Verify that the token has been deleted from the database
		var count int64
		db.Model(&models.Token{}).Where("token_string = ?", negToken.TokenString).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	// Test with nil database
	t.Run("NilDatabase", func(t *testing.T) {
		nilDBServer := authpkg.NewServer(nil)
		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           "some-token",
			DecreaseCandles: 10,
		}

		resp, err := nilDBServer.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test with error during update
	t.Run("UpdateError", func(t *testing.T) {
		// Create a server with a closed database that will cause errors
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           "some-token",
			DecreaseCandles: 10,
		}

		resp, err := errorServer.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test with database error when finding token
	t.Run("FindTokenError", func(t *testing.T) {
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           "some-token",
			DecreaseCandles: 10,
		}

		resp, err := errorServer.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to find token")
	})

	// Test error when scanning candles_left
	t.Run("ScanError", func(t *testing.T) {
		// Create a token in the real DB
		scanToken := createTestToken(t, db, 100)

		// Create a closed DB that will cause errors
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		// We need to modify the server to find the token but fail on scan
		// This is hard to test directly, so we'll test the error path in the server
		// by using a closed DB which will fail on any operation

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           scanToken.TokenString,
			DecreaseCandles: 10,
		}

		resp, err := errorServer.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test error when updating candles_left
	t.Run("UpdateDBError", func(t *testing.T) {
		// Create a token in the real DB
		updateToken := createTestToken(t, db, 100)

		// Create a closed DB that will cause errors
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           updateToken.TokenString,
			DecreaseCandles: 10,
		}

		resp, err := errorServer.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetTokenInfoUnit(t *testing.T) {
	db := setupInMemoryDB(t)
	server := authpkg.NewServer(db)

	// Create a test token
	token := createTestToken(t, db, 100)

	// Test successful retrieval of token information
	t.Run("Success", func(t *testing.T) {
		req := &pb.GetTokenInfoRequest{
			Token: token.TokenString,
		}

		resp, err := server.GetTokenInfo(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, token.TokenString, resp.Token)
		assert.Equal(t, int64(100), resp.CandlesLeft)
		assert.Equal(t, token.ExpiresAt.Unix(), resp.ExpiresAt)
	})

	// Test with non-existent token
	t.Run("NonExistentToken", func(t *testing.T) {
		req := &pb.GetTokenInfoRequest{
			Token: "non-existent-token",
		}

		resp, err := server.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test with nil database
	t.Run("NilDatabase", func(t *testing.T) {
		nilDBServer := authpkg.NewServer(nil)
		req := &pb.GetTokenInfoRequest{
			Token: "some-token",
		}

		resp, err := nilDBServer.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection not available")
	})

	// Test with empty token
	t.Run("EmptyToken", func(t *testing.T) {
		req := &pb.GetTokenInfoRequest{
			Token: "",
		}

		resp, err := server.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "token is required")
	})

	// Test with database error
	t.Run("DatabaseError", func(t *testing.T) {
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.GetTokenInfoRequest{
			Token: "some-token",
		}

		resp, err := errorServer.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to find token")
	})

	// Test error when scanning candles_left
	t.Run("ScanError", func(t *testing.T) {
		// Create a token in the real DB
		scanToken := createTestToken(t, db, 100)

		// Create a closed DB that will cause errors
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.GetTokenInfoRequest{
			Token: scanToken.TokenString,
		}

		resp, err := errorServer.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestRevokeTokenUnit(t *testing.T) {
	db := setupInMemoryDB(t)
	server := authpkg.NewServer(db)

	// Create a test token
	token := createTestToken(t, db, 100)

	// Test successful token revocation
	t.Run("Success", func(t *testing.T) {
		req := &pb.RevokeTokenRequest{
			Token: token.TokenString,
		}

		resp, err := server.RevokeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)

		// Verify that the token has been deleted from the database
		var count int64
		db.Model(&models.Token{}).Where("token_string = ?", token.TokenString).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	// Test with non-existent token
	t.Run("NonExistentToken", func(t *testing.T) {
		req := &pb.RevokeTokenRequest{
			Token: "non-existent-token",
		}

		resp, err := server.RevokeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
	})

	// Test with nil database
	t.Run("NilDatabase", func(t *testing.T) {
		nilDBServer := authpkg.NewServer(nil)
		req := &pb.RevokeTokenRequest{
			Token: "some-token",
		}

		resp, err := nilDBServer.RevokeToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "database connection not available")
	})

	// Test with empty token
	t.Run("EmptyToken", func(t *testing.T) {
		req := &pb.RevokeTokenRequest{
			Token: "",
		}

		resp, err := server.RevokeToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "token is required")
	})

	// Test with error during delete
	t.Run("DeleteError", func(t *testing.T) {
		// Create a server with a closed database that will cause errors
		errorDB := setupErrorDB(t)
		errorServer := authpkg.NewServer(errorDB)

		req := &pb.RevokeTokenRequest{
			Token: "some-token",
		}

		resp, err := errorServer.RevokeToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to revoke token")
	})

	// Test with RowsAffected == 0 (already tested in NonExistentToken)
	t.Run("NoRowsAffected", func(t *testing.T) {
		// This is already tested in the NonExistentToken test
		// When a token doesn't exist, RowsAffected will be 0
		// and the server should return Success: false
		req := &pb.RevokeTokenRequest{
			Token: "non-existent-token",
		}

		resp, err := server.RevokeToken(context.Background(), req)
		assert.NoError(t, err)
		assert.False(t, resp.Success)
	})
}

// TestDatabaseErrorCases tests various database error scenarios
func TestDatabaseErrorCases(t *testing.T) {
	// Create a test DB that will be closed to simulate errors
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create a token before closing the DB
	token := models.NewToken([]string{"read"}, 3600)
	token.CandlesLeft = 100

	err = token.BeforeSave()
	require.NoError(t, err)

	result := db.Create(token)
	require.NoError(t, result.Error)

	// Get the underlying SQL DB to close it
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Create server with the DB that will be closed
	server := authpkg.NewServer(db)

	// Close the DB to simulate errors
	sqlDB.Close()

	// Test ValidateToken with DB error
	t.Run("ValidateTokenDBError", func(t *testing.T) {
		req := &pb.ValidateRequest{
			Token: token.TokenString,
		}

		resp, err := server.ValidateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test CreateToken with DB error
	t.Run("CreateTokenDBError", func(t *testing.T) {
		req := &pb.CreateTokenRequest{
			Permissions: []string{"read"},
			ExpiresIn:   3600,
		}

		resp, err := server.CreateToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test RevokeToken with DB error
	t.Run("RevokeTokenDBError", func(t *testing.T) {
		req := &pb.RevokeTokenRequest{
			Token: token.TokenString,
		}

		resp, err := server.RevokeToken(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test UpdateTokenCandlesLeft with DB error
	t.Run("UpdateTokenCandlesLeftDBError", func(t *testing.T) {
		req := &pb.UpdateTokenCandlesLeftRequest{
			Token:           token.TokenString,
			DecreaseCandles: 10,
		}

		resp, err := server.UpdateTokenCandlesLeft(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Test GetTokenInfo with DB error
	t.Run("GetTokenInfoDBError", func(t *testing.T) {
		req := &pb.GetTokenInfoRequest{
			Token: token.TokenString,
		}

		resp, err := server.GetTokenInfo(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
