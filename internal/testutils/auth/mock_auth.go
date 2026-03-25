package auth

import (
	"time"

	"github.com/yourorg/querybase/internal/auth"
)

const (
	TestJWTSecret     = "test-secret-key-for-testing-only-min-32-chars"
	TestJWTExpireTime = 24 * time.Hour
	TestJWTIssuer     = "querybase-test"
)

var (
	TestAdminUserID  = "00000000-0000-0000-0000-000000000001"
	TestUserID       = "00000000-0000-0000-0000-000000000002"
	TestViewerUserID = "00000000-0000-0000-0000-000000000003"
)

var (
	TestAdminEmail  = "admin@test.local"
	TestUserEmail   = "user@test.local"
	TestViewerEmail = "viewer@test.local"
)

func CreateTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager(TestJWTSecret, TestJWTExpireTime, TestJWTIssuer)
}

func CreateTestJWTToken(userID, email, role string) (string, error) {
	manager := CreateTestJWTManager()
	uid := auth.MustParseUUID(userID)
	return manager.GenerateToken(uid, email, role)
}

func CreateTestJWTTokenWithManager(manager *auth.JWTManager, userID, email, role string) (string, error) {
	uid := auth.MustParseUUID(userID)
	return manager.GenerateToken(uid, email, role)
}
