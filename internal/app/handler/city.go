package handler

import (
	"net/http"
	"strconv"

	"lab1-design-backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetCites(ctx *gin.Context) {
	var cites []ds.City
	var err error
	creatorID := h.Repository.GetUser()

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		cites, err = h.Repository.GetCites()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			logrus.Error(err)
			return
		}
	} else {
		cites, err = h.Repository.GetCitesByName(searchQuery)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			logrus.Error(err)
			return
		}
	}
	currentService, _ := h.Repository.CheckCurrentServiceDraft(creatorID)

	ctx.HTML(http.StatusOK, "main.html", gin.H{
		"cities":       cites,
		"serviceCount": h.Repository.GetServiceCount(),
		"query":        searchQuery,
		"serviceId":    currentService.ID,
	})
}

func (h *Handler) GetCity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	city, err := h.Repository.GetCity(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "city.html", gin.H{
		"clty": city,
	})
}

func (h *Handler) AddCityToService(ctx *gin.Context) {
	service, err := h.Repository.GetServiceDraft(h.Repository.GetUser())
	serviceId := service.ID
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	cityId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.AddCityToService(int(serviceId), cityId)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/cites")
}
