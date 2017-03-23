package http

import (
	"net/http"

	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/helper"
	"gopkg.in/gin-gonic/gin.v1"
)

type AuthHandler struct {
	Logger        logger
	CryptoService pub.CryptoService
	JWTService    pub.JWTService
	UserService   pub.UserService
}

const (
	// ErrInvalidCredentialsFormat is an error raised when credentials format is not valid
	errInvalidCredentialsFormat = pub.Error("Invalid credentials format")
	// ErrInvalidCredentials is an error raised when credentials for a user are invalid
	errInvalidCredentials = pub.Error("Username/Email with wrong password")
)

func (h *AuthHandler) signIn(ctx *gin.Context) {
	var req postAuthRequest
	if err := ctx.BindJSON(&req); err != nil {
		// return ErrInvalidJSON or ErrInvalidRequestFormat
		Error(ctx, ErrInvalidJSON, http.StatusBadRequest, h.Logger)
		return
	}
	var (
		u   *pub.User
		err error
	)
	if helper.IsEmail(req.User) {
		u, err = h.UserService.UserByEmail(req.User)
	} else {
		u, err = h.UserService.UserByUsername(req.User)
	}
	if err == pub.ErrUserNotFound {
		Error(ctx, err, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	if !u.IsActive {
		Error(ctx, pub.ErrUserInactive, http.StatusForbidden, nil)
		return
	}
	if err := h.CryptoService.CompareHashAndData(u.Password, req.Password); err != nil {
		Error(ctx, errInvalidCredentials, http.StatusUnprocessableEntity, nil)
		return
	}
	tokenData := &pub.TokenData{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
	}
	token, err := h.JWTService.GenerateToken(tokenData)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &postAuthResponse{JWT: token})
}

// `user` field = username/email
type postAuthRequest struct {
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type postAuthResponse struct {
	JWT string `json:"jwt"`
}
