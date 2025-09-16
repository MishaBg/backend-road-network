package handler

import (
	"lab1-design-backend/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetCites(ctx *gin.Context) {
	var cites []repository.City
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		cites, err = h.Repository.GetProtocols()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		cites, err = h.Repository.GetCitesByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "main.html", gin.H{
		"cites":        cites,
		"query":        searchQuery,
		"requestId":    h.Repository.GetRequestId(),
		"requestCount": h.Repository.GetRequestsCount(1),
	})
}

func (h *Handler) GetCity(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}

	city, err := h.Repository.GetCity(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "city.html", gin.H{
		"city": city,
	})
}

func (h *Handler) RequestHandler(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
	}
	calculationProtocols := h.Repository.GetRequestCites(id)
	ctx.HTML(http.StatusOK, "app.html", gin.H{
		"requestCites": calculationProtocols,
		"count":        len(calculationProtocols),
	})
}
