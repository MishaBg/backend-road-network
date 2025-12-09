package handler

import (
	"errors"
	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteCityFromService godoc
// @Summary Удалить город из сервиса
// @Description Удаляет связь города и сервиса
// @Tags cites-service
// @Produce json
// @Param city_id path int true "ID города"
// @Param service_id path int true "ID сервис"
// @Success 200 {object} apitypes.ServiceJSON "Обновленный сервис"
// @Failure 400 {object} map[string]string "Неверные ID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /cites_road/{cityt_id}/{road_id}/delete [delete]
func (h *Handler) DeleteCityFromService(ctx *gin.Context) {
	serviceId, err := strconv.Atoi(ctx.Param("road_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	cityId, err := strconv.Atoi(ctx.Param("city_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.DeleteCityFromService(serviceId, cityId)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(service)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, apitypes.ServiceToJSON(service, creatorLogin, moderatorLogin))
}

// ChangeCityService godoc
// @Summary Изменить данные города в сервисе
// @Description Обновляет параметры города в конкретном сервисе
// @Tags cites-service
// @Accept json
// @Produce json
// @Param city_id path int true "ID города"
// @Param road_id path int true "ID сервиса"
// @Param data body apitypes.CitesServiceJSON true "Новые данные"
// @Success 200 {object} apitypes.CitesServiceJSON "Обновленные данные"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 409 {object} map[string]string "Конфликт - только один центральный город"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /cites_road/{city_id}/{road_id}/change [put]
func (h *Handler) ChangeCityService(ctx *gin.Context) {
	serviceId, err := strconv.Atoi(ctx.Param("road_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	cityId, err := strconv.Atoi(ctx.Param("city_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var cityServiceJSON apitypes.CitesServiceJSON
	if err := ctx.BindJSON(&cityServiceJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	cityService, err := h.Repository.ChangeСityService(serviceId, cityId, cityServiceJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.CitesServiceToJSON(cityService))
}
