package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"lab1-design-backend/internal/app/handler"
	"lab1-design-backend/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("../../templates/*")

	r.Static("/static", "../../resources")

	r.GET("/main", handler.GetCites)
	r.GET("/city/:id", handler.GetCity)
	r.GET("/request/:id", handler.RequestHandler)

	r.Run()

	log.Println("Server down")
}
