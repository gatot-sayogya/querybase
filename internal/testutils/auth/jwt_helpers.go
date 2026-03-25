package auth

import (
	"time"

	"github.com/yourorg/querybase/internal/auth"
)

func CreateAdminToken() (string, error) {
	return CreateTestJWTToken(TestAdminUserID, TestAdminEmail, "admin")
}

func CreateUserToken() (string, error) {
	return CreateTestJWTToken(TestUserID, TestUserEmail, "user")
}

func CreateViewerToken() (string, error) {
	return CreateTestJWTToken(TestViewerUserID, TestViewerEmail, "viewer")
}

func CreateAdminTokenWithManager(manager *auth.JWTManager) (string, error) {
	return CreateTestJWTTokenWithManager(manager, TestAdminUserID, TestAdminEmail, "admin")
}

func CreateUserTokenWithManager(manager *auth.JWTManager) (string, error) {
	return CreateTestJWTTokenWithManager(manager, TestUserID, TestUserEmail, "user")
}

func CreateViewerTokenWithManager(manager *auth.JWTManager) (string, error) {
	return CreateTestJWTTokenWithManager(manager, TestViewerUserID, TestViewerEmail, "viewer")
}

func CreateExpiredToken() (string, error) {
	manager := auth.NewJWTManager(TestJWTSecret, -1*time.Hour, TestJWTIssuer)
	uid := auth.MustParseUUID(TestUserID)
	return manager.GenerateToken(uid, TestUserEmail, "user")
}
