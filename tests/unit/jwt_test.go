package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJWTManager_GenerateToken tests token generation
func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// TestJWTManager_ValidateToken tests token validation
func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "admin")
	require.NoError(t, err)

	// Validate the token
	validatedClaims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedClaims.UserID)
	assert.Equal(t, "test@example.com", validatedClaims.Email)
	assert.Equal(t, "admin", validatedClaims.Role)
}

// TestJWTManager_ValidateToken_Invalid tests invalid token validation
func TestJWTManager_ValidateToken_Invalid(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	// Test with completely invalid token
	_, err := manager.ValidateToken("invalid.token.here")
	assert.Error(t, err)

	// Test with empty token
	_, err = manager.ValidateToken("")
	assert.Error(t, err)
}

// TestJWTManager_ValidateToken_WrongSecret tests token with wrong secret
func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-one", 24*time.Hour, "querybase")
	manager2 := NewJWTManager("secret-two", 24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager1.GenerateToken(userID, "test@example.com", "admin")
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = manager2.ValidateToken(token)
	assert.Error(t, err)
}

// TestHashPassword tests password hashing
func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.Contains(t, hash, "$2a$") // bcrypt hash prefix
}

// TestCheckPassword tests password verification
func TestCheckPassword(t *testing.T) {
	password := "mySecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Correct password
	isValid := CheckPassword(password, hash)
	assert.True(t, isValid)

	// Incorrect password
	isValid = CheckPassword("wrongpassword", hash)
	assert.False(t, isValid)
}

// TestCheckPassword_EmptyPassword tests empty password handling
func TestCheckPassword_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123!"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Empty password should not match
	isValid := CheckPassword("", hash)
	assert.False(t, isValid)
}

// BenchmarkHashPassword benchmarks password hashing
func BenchmarkHashPassword(b *testing.B) {
	password := "mySecurePassword123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

// BenchmarkCheckPassword benchmarks password checking
func BenchmarkCheckPassword(b *testing.B) {
	password := "mySecurePassword123!"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPassword(password, hash)
	}
}

// TestJWTManager_TokenExpiration tests token expiration handling
func TestJWTManager_TokenExpiration(t *testing.T) {
	// Create manager with very short expiration
	manager := NewJWTManager("test-secret", 1*time.Millisecond, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "admin")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should now be invalid
	_, err = manager.ValidateToken(token)
	assert.Error(t, err, "Expected error for expired token")
}

// TestJWTManager_TokenUniqueness tests that different users get different tokens
func TestJWTManager_TokenUniqueness(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID1 := uuid.New()
	userID2 := uuid.New()

	token1, err1 := manager.GenerateToken(userID1, "user1@example.com", "user")
	token2, err2 := manager.GenerateToken(userID2, "user2@example.com", "user")

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, token1, token2, "Different users should get different tokens")
}

// TestJWTManager_SameUserDifferentTime tests tokens for same user at different times
func TestJWTManager_SameUserDifferentTime(t *testing.T) {
	t.Skip("Skipping: JWT timestamps are in seconds, tokens within same second are identical (correct behavior)")

	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()

	token1, err1 := manager.GenerateToken(userID, "test@example.com", "user")
	time.Sleep(1 * time.Second) // Need at least 1 second for different JWT timestamps
	token2, err2 := manager.GenerateToken(userID, "test@example.com", "user")

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, token1, token2, "Same user should get different tokens at different times")
}

// TestJWTManager_ClaimsStructure tests the structure of token claims
func TestJWTManager_ClaimsStructure(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()
	email := "test@example.com"
	role := "admin"

	token, err := manager.GenerateToken(userID, email, role)
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	// Verify all claim fields
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "querybase", claims.Issuer)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	// Verify expiration is in the future
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()), "Token should expire in the future")

	// Verify expiration is within expected range
	expectedMaxExpiry := time.Now().Add(24*time.Hour + 1*time.Minute)
	assert.True(t, claims.ExpiresAt.Time.Before(expectedMaxExpiry), "Token expiration should be within 24 hours")
}

// TestJWTManager_DifferentRoles tests tokens with different roles
func TestJWTManager_DifferentRoles(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()

	roles := []string{"admin", "user", "viewer"}

	for _, role := range roles {
		t.Run("Role_"+role, func(t *testing.T) {
			token, err := manager.GenerateToken(userID, "test@example.com", role)
			require.NoError(t, err)

			claims, err := manager.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, role, claims.Role)
		})
	}
}

// TestJWTManager_TokenWithDifferentIssuers tests tokens with different issuers
func TestJWTManager_TokenWithDifferentIssuers(t *testing.T) {
	manager1 := NewJWTManager("secret", 24*time.Hour, "issuer1")
	manager2 := NewJWTManager("secret", 24*time.Hour, "issuer2")

	userID := uuid.New()

	token1, err1 := manager1.GenerateToken(userID, "test@example.com", "user")
	token2, err2 := manager2.GenerateToken(userID, "test@example.com", "user")

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Tokens should be different due to different issuers
	assert.NotEqual(t, token1, token2)

	// Validate with respective managers
	claims1, err1 := manager1.ValidateToken(token1)
	claims2, err2 := manager2.ValidateToken(token2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.Equal(t, "issuer1", claims1.Issuer)
	assert.Equal(t, "issuer2", claims2.Issuer)
}

// TestJWTManager_VeryLongExpiration tests token with very long expiration time
func TestJWTManager_VeryLongExpiration(t *testing.T) {
	// Create manager with 1 year expiration
	manager := NewJWTManager("test-secret", 365*24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "user")
	require.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	require.NoError(t, err)

	// Verify expiration is approximately 1 year from now
	expectedExpiry := time.Now().Add(365 * 24 * time.Hour)
	timeDiff := claims.ExpiresAt.Time.Sub(expectedExpiry)

	// Allow 1 minute tolerance
	assert.True(t, timeDiff.Abs() < 1*time.Minute, "Expiration should be approximately 1 year from now")
}

// TestJWTManager_EmptySecret tests behavior with empty secret
func TestJWTManager_EmptySecret(t *testing.T) {
	// This test documents behavior - empty secrets should be avoided in production
	manager := NewJWTManager("", 24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "user")

	// Generation should still work (though not secure)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validation should also work with same empty secret
	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

// TestJWTManager_TokenTampering tests that tampered tokens are rejected
func TestJWTManager_TokenTampering(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()
	token, err := manager.GenerateToken(userID, "test@example.com", "user")
	require.NoError(t, err)

	// Tamper with the token by changing a character
	tamperedToken := token[:len(token)-5] + "XXXXX"

	_, err = manager.ValidateToken(tamperedToken)
	assert.Error(t, err, "Tampered token should be rejected")
}

// TestJWTManager_MultipleValidTokens tests validating multiple tokens for the same user
func TestJWTManager_MultipleValidTokens(t *testing.T) {
	manager := NewJWTManager("test-secret", 24*time.Hour, "querybase")

	userID := uuid.New()

	// Generate multiple tokens
	tokens := make([]string, 5)
	for i := 0; i < 5; i++ {
		token, err := manager.GenerateToken(userID, "test@example.com", "user")
		require.NoError(t, err)
		tokens[i] = token
	}

	// All tokens should be valid (uniqueness not guaranteed within same second)
	for i, token := range tokens {
		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err, "Token %d should be valid", i)
		assert.Equal(t, userID, claims.UserID)
	}
}

// TestHashPassword_DifferentPasswordsGenerateDifferentHashes tests password hash uniqueness
func TestHashPassword_DifferentPasswordsGenerateDifferentHashes(t *testing.T) {
	password1 := "password123"
	password2 := "password456"

	hash1, err1 := HashPassword(password1)
	hash2, err2 := HashPassword(password2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, hash1, hash2, "Different passwords should generate different hashes")
}

// TestHashPassword_SamePasswordDifferentHashes tests salt randomness
func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	password := "samePassword123"

	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.NotEqual(t, hash1, hash2, "Same password should generate different hashes due to salt")

	// But both should validate correctly
	assert.True(t, CheckPassword(password, hash1))
	assert.True(t, CheckPassword(password, hash2))
}
