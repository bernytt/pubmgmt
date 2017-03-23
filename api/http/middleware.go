package http

import (
	"net/http"
	"strings"

	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

type middleWareService struct {
	jwtService   pub.JWTService
	authDisabled bool
}

const (
	contextAuthenticationKey = "auth"
)

func extractTokenDataFromRequestContext(ctx *gin.Context) (*pub.TokenData, error) {
	contextData, ok := ctx.Get(contextAuthenticationKey)
	if !ok {
		return nil, pub.ErrMissingContextData
	}
	tokenData := contextData.(*pub.TokenData)
	return tokenData, nil
}

// mwSecureHeaders provides secure headers middleware for handlers
func mwSecureHeaders(ctx *gin.Context) {
	ctx.Request.Header.Add("X-Content-Type-Options", "nosniff")
	ctx.Request.Header.Add("X-Frame-Options", "DENY")
	ctx.Next()
}

// mwCheckAdministratorRole check the role of the user associated to the request
func (service *middleWareService) mwCheckAdministratorRole() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenData, err := extractTokenDataFromRequestContext(ctx)
		if err != nil {
			ErrorAbort(ctx, pub.ErrResourceAccessDenied, http.StatusForbidden, nil)
			return
		}
		if tokenData.Role != pub.AdministratorRole {
			ErrorAbort(ctx, pub.ErrResourceAccessDenied, http.StatusForbidden, nil)
			return
		}
		ctx.Next()
	}
}

// mwCheckAuthentication provides Authentication middleware for handlers
func (service *middleWareService) mwCheckAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var tokenData *pub.TokenData
		if !service.authDisabled {
			var token string
			// Get token from the Authorization header
			tokens, ok := ctx.Request.Header["Authorization"]
			if ok && len(tokens) >= 1 {
				token = tokens[0]
				token = strings.TrimPrefix(token, "Bearer ")
			}

			if token == "" {
				ErrorAbort(ctx, pub.ErrUnauthorized, http.StatusUnauthorized, nil)
				return
			}
			var err error
			tokenData, err = service.jwtService.ParseAndVerifyToken(token)
			if err != nil {
				ErrorAbort(ctx, err, http.StatusUnauthorized, nil)
				return
			}
		} else {
			tokenData = &pub.TokenData{
				Role: pub.AdministratorRole,
			}
		}
		ctx.Set(contextAuthenticationKey, tokenData)
		ctx.Next()
	}
}
