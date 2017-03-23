package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

type UserHandler struct {
	Logger        logger
	CryptoService pub.CryptoService
	JWTService    pub.JWTService
	UserService   pub.UserService
	adminExists   bool
}

// url: /users  method: PUT  body: pub.User
// use for create `StandardUserRole` user
func (u *UserHandler) createUser(ctx *gin.Context) {
	var req pub.User
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	user, err := u.UserService.UserByUsername(req.Username)
	if err != nil && err != pub.ErrUserNotFound {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if user != nil {
		Error(ctx, pub.ErrUsernameAlreadyExists, http.StatusConflict, nil)
		return
	}
	user, err = u.UserService.UserByEmail(req.Email)
	if err != nil && err != pub.ErrUserNotFound {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if user != nil {
		Error(ctx, pub.ErrEmailAlreadyExists, http.StatusConflict, nil)
		return
	}

	// only accept role=pub.StandardUserRole
	// after create a user, admin can change user's role with POST method in handler(handlePostUser).
	user = &pub.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     pub.StandardUserRole,
		Avatar:   req.Avatar,
		IsActive: true,
	}
	user.Password, err = u.CryptoService.Hash(req.Password)
	if err != nil {
		Error(ctx, pub.ErrCryptoHashFailure, http.StatusInternalServerError, u.Logger)
		return
	}
	err = u.UserService.CreateUser(user)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "Put user success"})
}

// url: /users?role=:role method: GET
// only for admin user to get all users
func (u *UserHandler) getUsers(ctx *gin.Context) {
	role, err := strconv.Atoi(ctx.DefaultQuery("role", "0"))
	if err != nil || role < 0 || role > 2 {
		Error(ctx, pub.Error("role must equal to 1(admin) or 2(standard)"), http.StatusBadRequest, nil)
		return
	}
	var users []pub.User

	if role == 0 {
		users, err = u.UserService.Users()
	} else {
		users, err = u.UserService.UsersByRole(pub.UserRole(role))
	}
	if err == pub.ErrModelSetEmpty {
		Error(ctx, pub.Error("Not users found"), http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	for i := range users {
		users[i].Password = ""
	}
	ctx.IndentedJSON(http.StatusOK, users)
}

// url: /users/profile/:id  method: GET
func (u *UserHandler) getUserProfileByID(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	tokenData, err := extractTokenDataFromRequestContext(ctx)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if tokenData.ID == userID || tokenData.Role == pub.AdministratorRole {
		user, err := u.UserService.User(userID)
		if err == pub.ErrObjNotFound {
			Error(ctx, pub.ErrUserNotFound, http.StatusNotFound, nil)
			return
		} else if err != nil {
			Error(ctx, err, http.StatusInternalServerError, u.Logger)
			return
		}
		user.Password = ""
		ctx.IndentedJSON(http.StatusOK, user)
	} else {
		Error(ctx, pub.Error("Current user is not an admin user"), http.StatusForbidden, nil)
	}
}

// url: /users/profile/:id  method: POST
func (u *UserHandler) updateUserProfileByID(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	tokenData, err := extractTokenDataFromRequestContext(ctx)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if tokenData.Role != pub.AdministratorRole && tokenData.ID != userID {
		Error(ctx, pub.ErrUnauthorized, http.StatusUnauthorized, nil)
		return
	}
	var req postUserRequest
	if err = ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.ID == 0 || req.ID != userID {
		Error(ctx, pub.Error("ID is a hidden filed, equal to ID in request uri"), http.StatusBadRequest, nil)
		return
	}
	user, err := u.UserService.User(userID)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrUserNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if req.Password != "" {
		if user.Password, err = u.CryptoService.Hash(req.Password); err != nil {
			Error(ctx, pub.ErrCryptoHashFailure, http.StatusInternalServerError, u.Logger)
			return
		}
	}
	if req.Role != 0 {
		if tokenData.Role != pub.AdministratorRole {
			Error(ctx, pub.ErrUnauthorized, http.StatusUnauthorized, nil)
			return
		}
		if req.Role == 1 {
			user.Role = pub.AdministratorRole
		} else {
			user.Role = pub.StandardUserRole
		}
	}
	err = u.UserService.UpdateUser(user.ID, user)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Update user success"})
}

type postUserRequest struct {
	ID       uint64
	Email    string `binding:"required"`
	Password string
	Role     pub.UserRole
	Avatar   string `json:"Avatar"`
	IsActive bool   `json:"IsActive"`
}

// url: /users/profile/:id  method: DELETE
func (u *UserHandler) deleteUserByID(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	tokenData, err := extractTokenDataFromRequestContext(ctx)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	if tokenData.Role != pub.AdministratorRole && tokenData.ID != userID {
		Error(ctx, pub.ErrUnauthorized, http.StatusUnauthorized, nil)
		return
	}

	_, err = u.UserService.User(userID)

	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrUserNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	err = u.UserService.DeleteUser(userID)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Delete user success"})
}

// url: /users/admin/init  method: POST  body:postAdminPassword
// it's forbidden when admin user is exists
func (u *UserHandler) initAdmin(ctx *gin.Context) {
	if u.adminExists {
		Error(ctx, pub.ErrAdminAlreadyInitialized, http.StatusForbidden, nil)
		return
	}
	var (
		req postAdminPassword
		err error
	)
	if err = ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	admin := &pub.User{
		Username: "admin",
		Role:     pub.AdministratorRole,
		IsActive: true,
	}
	admin.Password, err = u.CryptoService.Hash(req.Password)
	if err != nil {
		Error(ctx, pub.ErrCryptoHashFailure, http.StatusInternalServerError, u.Logger)
		return
	}
	err = u.UserService.CreateUser(admin)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, u.Logger)
		return
	}
	u.adminExists = true
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "Initial admin user success"})
}

// use a go routine to check admin user is exists, set the handler.adminExists to true
func (u *UserHandler) checkAdminExists() {
	ticker := time.Tick(10 * time.Second)
	for {
		select {
		case <-ticker:
			user, _ := u.UserService.UserByUsername("admin")
			if user != nil {
				u.adminExists = true
				return
			}
		}
	}
}

type postAdminPassword struct {
	Password string `json:"Password" binding:"required,min=6,max=32"`
}
