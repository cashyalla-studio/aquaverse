package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyUserRole = "user_role"
	ContextKeyLocale   = "locale"
)

type AquaClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			token, err := jwt.ParseWithClaims(parts[1], &AquaClaims{}, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			claims, ok := token.Claims.(*AquaClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyUserRole, claims.Role)
			return next(c)
		}
	}
}

// OptionalJWTAuth - 인증 선택적 (공개 API지만 로그인 시 사용자 정보 활용)
func OptionalJWTAuth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return next(c)
			}

			token, err := jwt.ParseWithClaims(parts[1], &AquaClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*AquaClaims); ok {
					c.Set(ContextKeyUserID, claims.UserID)
					c.Set(ContextKeyUserRole, claims.Role)
				}
			}
			return next(c)
		}
	}
}

// MustGetUserID JWT claims에서 userID를 추출하여 UUID로 반환한다.
// JWTAuth 미들웨어가 설정한 claims를 사용하며, 파싱 실패 시 uuid.Nil을 반환한다.
func MustGetUserID(c echo.Context) uuid.UUID {
	claims, ok := c.Get(ContextKeyUserID).(string)
	if !ok {
		return uuid.Nil
	}
	id, _ := uuid.Parse(claims)
	return id
}

// RequireRole - 특정 역할 요구
func RequireRole(roles ...string) echo.MiddlewareFunc {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(ContextKeyUserRole).(string)
			if !ok || !roleSet[role] {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
			}
			return next(c)
		}
	}
}
