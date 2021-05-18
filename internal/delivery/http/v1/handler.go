package v1

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/peleng-meteo/meteo-go/internal/service"
	"gitlab.com/peleng-meteo/meteo-go/pkg/auth"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewHandler(services *service.Services, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initUsersRoutes(v1)
		h.initCallbackRoutes(v1)
		h.initAdminRoutes(v1)

		// TODO: check this
		/*
			v1.GET("/settings", h.setSchoolFromRequest, h.getSchoolSettings)
			v1.GET("/promocodes/:code", h.setSchoolFromRequest, h.getPromo)
		*/
	}
}
