package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fengxsong/pubmgmt/api"
	"github.com/gorilla/securecookie"
)

type Service struct {
	secret []byte
}

type claims struct {
	UserID   uint64       `json:"id"`
	Username string       `json:"username"`
	Role     pub.UserRole `json:"role"`
	jwt.StandardClaims
}

func NewService() (*Service, error) {
	secret := securecookie.GenerateRandomKey(32)
	if secret == nil {
		return nil, pub.ErrSecretGeneration
	}
	service := &Service{
		secret,
	}
	return service, nil
}

func (service *Service) GenerateToken(data *pub.TokenData) (string, error) {
	expireToken := time.Now().Add(time.Hour * 8).Unix()
	cl := claims{
		data.ID,
		data.Username,
		data.Role,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)

	signedToken, err := token.SignedString(service.secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (service *Service) ParseAndVerifyToken(token string) (*pub.TokenData, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			msg := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			return nil, msg
		}
		return service.secret, nil
	})
	if err == nil && parsedToken != nil {
		if cl, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
			tokenData := &pub.TokenData{
				ID:       cl.UserID,
				Username: cl.Username,
				Role:     pub.UserRole(cl.Role),
			}
			return tokenData, nil
		}
	}
	return nil, pub.ErrInvalidJWTToken
}
