package handler

import (
	"errors"
	"lab4-swag/docs"
	"lab4-swag/internal/app/config"
	"lab4-swag/internal/app/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, c *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     c,
	}
}

// RegisterHandler godoc
// @title Road Network API
// @version 1.0
// @description API для управления расчетом дорожной сети
// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@astronomy.com
// @license.name MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (h *Handler) RegisterHandler(router *gin.Engine) {

	api := router.Group("/api")

	unauthorized := api.Group("/")
	unauthorized.POST("/users/sign-up", h.SignUp)
	unauthorized.PUT("/road/:id/set-cost", h.SetServiceCost)
	unauthorized.GET("/cites", h.GetCites)
	unauthorized.GET("/city/:id", h.GetCity)
	unauthorized.POST("/users/sign-in", h.SignIn)

	authorized := api.Group("/")
	authorized.Use(h.ModeratorMiddleware(false))
	authorized.POST("/city/:id/add-to-road", h.AddCityToService)
	authorized.GET("/road/road-cart", h.GetServiceCart)
	authorized.GET("/roads", h.GetServices)
	authorized.GET("/road/:id", h.GetService)
	authorized.PUT("/road/:id/change-road", h.ChangeService)
	authorized.PUT("/road/:id/form", h.FormService)
	authorized.DELETE("/road/:id/delete-road", h.DeleteService)
	authorized.DELETE("/cites_road/:city_id/:road_id/delete", h.DeleteCityFromService)
	authorized.PUT("/cites_road/:city_id/:road_id/change", h.ChangeCityService)
	authorized.GET("/users/:login/profile", h.GetProfile)
	authorized.PUT("/users/:login/profile", h.ChangeProfile)
	authorized.POST("/users/sign-out", h.SignOut)

	moderator := api.Group("/")
	moderator.Use(h.ModeratorMiddleware(true))
	moderator.PUT("/road/:id/finish", h.ModerateService)
	moderator.POST("/city/:id/create-image", h.UploadImage)
	moderator.POST("/city/create-city", h.CreateCity)
	moderator.DELETE("/city/:id/delete-city", h.DeleteCity)
	moderator.PUT("/city/:id/change-city", h.ChangeCity)

	// Инициализируем документацию Swagger
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"

	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/static", "/home/muka/Рабочий стол/RIP/LAB1/resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())

	// Добавляем более специфичную обработку ошибок
	var errorMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		errorMessage = "Не найден"
	case errors.Is(err, repository.ErrAlreadyExists):
		errorMessage = "Уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		errorMessage = "Доступ запрещен"
	case errors.Is(err, repository.ErrNoDraft):
		errorMessage = "Черновик не найден"
	default:
		errorMessage = err.Error()
	}

	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}
