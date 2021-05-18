package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/peleng-meteo/meteo-go/internal/domain"
	"gitlab.com/peleng-meteo/meteo-go/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) initUsersRoutes(api *gin.RouterGroup) {
	students := api.Group("/users", h.setSchoolFromRequest)
	{
		students.POST("/sign-up", h.userSignUp)
		students.POST("/sign-in", h.userSignIn)
		students.POST("/auth/refresh", h.userRefresh)
		students.POST("/verify/:code", h.userVerify)

		authenticated := students.Group("/", h.userIdentity)
		{
			authenticated.GET("/account", h.studentGetAccount)
		}
	}
}

type userSignUpInput struct {
	Name     string `json:"name" binding:"required,min=2,max=64"`
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

// @Summary Student SignUp
// @Tags students-auth
// @Description create student account
// @ModuleID studentSignUp
// @Accept  json
// @Produce  json
// @Param input body userSignUpInput true "sign up info"
// @Success 201 {string} string "ok"
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/sign-up [post]
func (h *Handler) userSignUp(c *gin.Context) {
	var inp userSignUpInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	if err := h.services.Users.SignUp(c.Request.Context(), service.SignUpInput{
		Name:     inp.Name,
		Email:    inp.Email,
		Password: inp.Password,
	}); err != nil {
		if err == service.ErrUserAlreadyExists {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusCreated)
}

type signInInput struct {
	Email    string `json:"email" binding:"required,email,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// @Summary Student SignIn
// @Tags students-auth
// @Description student sign in
// @ModuleID studentSignIn
// @Accept  json
// @Produce  json
// @Param input body signInInput true "sign up info"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/sign-in [post]
func (h *Handler) userSignIn(c *gin.Context) {
	var inp signInInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	res, err := h.services.Users.SignIn(c.Request.Context(), service.SignInInput{
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		if err == service.ErrUserNotFound {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

type refreshInput struct {
	Token string `json:"token" binding:"required"`
}

// @Summary Student Refresh Tokens
// @Tags students-auth
// @Description student refresh tokens
// @Accept  json
// @Produce  json
// @Param input body refreshInput true "sign up info"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/auth/refresh [post]
func (h *Handler) userRefresh(c *gin.Context) {
	var inp refreshInput
	if err := c.BindJSON(&inp); err != nil {
		newResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	res, err := h.services.Users.RefreshTokens(c.Request.Context(), inp.Token)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

// @Summary Student Verify Registration
// @Tags students-auth
// @Description student verify registration
// @ModuleID studentVerify
// @Accept  json
// @Produce  json
// @Param code path string true "verification code"
// @Success 200 {object} tokenResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /students/verify/{code} [post]
func (h *Handler) userVerify(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		newResponse(c, http.StatusBadRequest, "code is empty")
		return
	}

	if err := h.services.Students.Verify(c.Request.Context(), code); err != nil {
		if err == service.ErrVerificationCodeInvalid {
			newResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, response{"success"})
}

// @Summary Student Get Account Info
// @Security StudentsAuth
// @Tags students-account
// @Description student get account info
// @ModuleID userGetAccount
// @Accept  json
// @Produce  json
// @Success 200 {object} userAccountResponse
// @Failure 400,404 {object} response
// @Failure 500 {object} response
// @Failure default {object} response
// @Router /users/account [get]
func (h *Handler) userGetAccount(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := h.services.Users.GetById(c.Request.Context(), userId)
	if err != nil {
		newResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, userAccountResponse{
		Name:  user.Name,
		Email: user.Email,
	})
}
