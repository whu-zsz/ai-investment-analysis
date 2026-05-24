package middleware

import (
	"strings"
	"time"

	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/utils"
	"stock-analysis-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey         = "user_id"
	ContextUsernameKey       = "username"
	ContextTokenJTIKey       = "token_jti"
	ContextTokenExpiresAtKey = "token_expires_at"
)

type AuthContext struct {
	UserID         uint64
	Username       string
	TokenJTI       string
	TokenExpiresAt time.Time
}

func GetAuthContext(c *gin.Context) (AuthContext, bool) {
	tokenJTI := c.GetString(ContextTokenJTIKey)
	if tokenJTI == "" {
		return AuthContext{}, false
	}

	tokenExpiresAtValue, ok := c.Get(ContextTokenExpiresAtKey)
	if !ok {
		return AuthContext{}, false
	}

	tokenExpiresAt, ok := tokenExpiresAtValue.(time.Time)
	if !ok {
		return AuthContext{}, false
	}

	return AuthContext{
		UserID:         c.GetUint64(ContextUserIDKey),
		Username:       c.GetString(ContextUsernameKey),
		TokenJTI:       tokenJTI,
		TokenExpiresAt: tokenExpiresAt,
	}, true
}

func abortUnauthorized(c *gin.Context, message string) {
	response.Unauthorized(c, message)
	c.Abort()
}

func setAuthContext(c *gin.Context, authCtx AuthContext) {
	c.Set(ContextUserIDKey, authCtx.UserID)
	c.Set(ContextUsernameKey, authCtx.Username)
	c.Set(ContextTokenJTIKey, authCtx.TokenJTI)
	c.Set(ContextTokenExpiresAtKey, authCtx.TokenExpiresAt)
}

func AuthMiddleware(jwtSecret string, revokedTokenRepo repository.RevokedTokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortUnauthorized(c, "missing authorization header")
			return
		}

		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			abortUnauthorized(c, "invalid authorization format")
			return
		}

		claims, err := utils.ParseToken(tokenString, jwtSecret)
		if err != nil {
			abortUnauthorized(c, "invalid token")
			return
		}

		tokenJTI := strings.TrimSpace(claims.ID)
		if tokenJTI == "" || claims.ExpiresAt == nil {
			abortUnauthorized(c, "invalid token")
			return
		}

		revoked, err := revokedTokenRepo.ExistsByJTI(tokenJTI)
		if err != nil || revoked {
			abortUnauthorized(c, "invalid token")
			return
		}

		setAuthContext(c, AuthContext{
			UserID:         claims.UserID,
			Username:       claims.Username,
			TokenJTI:       tokenJTI,
			TokenExpiresAt: claims.ExpiresAt.Time,
		})

		c.Next()
	}
}
