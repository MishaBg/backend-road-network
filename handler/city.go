package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/ds"
	"lab4-swag/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// GetCites godoc
// @Summary Получить список городов
// @Description Возвращает все города или фильтрует по названию
// @Tags cites
// @Produce json
// @Param city_name query string false "Название горда для поиска"
// @Success 200 {array} apitypes.CityJSON "Список городов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /cites [get]
func (h *Handler) GetCites(ctx *gin.Context) {
	var cites []ds.City
	var err error

	searchQuery := ctx.Query("name")
	if searchQuery == "" {
		cites, err = h.Repository.GetCites()
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	} else {
		cites, err = h.Repository.GetCitesByName(searchQuery)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	resp := make([]apitypes.CityJSON, 0, len(cites))
	for _, r := range cites {
		resp = append(resp, apitypes.CityToJSON(r))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetCity godoc
// @Summary Получить город по ID
// @Description Возвращает информацию о планете по её идентификатору
// @Tags cites
// @Produce json
// @Param id path int true "ID города"
// @Success 200 {object} apitypes.CityJSON "Данные города"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "Город не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /city/{id} [get]
func (h *Handler) GetCity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	city, err := h.Repository.GetCity(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.CityToJSON(*city))
}

// CreateCity godoc
// @Summary Создать новый город
// @Description Создает новый город и возвращает её данные
// @Tags cites
// @Accept json
// @Produce json
// @Param city body apitypes.CityJSON true "Данные нового города"
// @Success 201 {object} apitypes.CityJSON "Созданный город"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /city/create-city [post]
func (h *Handler) CreateCity(ctx *gin.Context) {
	var cityJSON apitypes.CityJSON
	if err := ctx.BindJSON(&cityJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	city, err := h.Repository.CreateCity(cityJSON)
	fmt.Println(err)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Header("Location", fmt.Sprintf("/cites/%v", city.ID))
	ctx.JSON(http.StatusCreated, apitypes.CityToJSON(city))
}

// DeleteCity godoc
// @Summary Удалить город
// @Description Выполняет логическое удаление города по ID
// @Tags cites
// @Produce json
// @Param id path int true "ID города"
// @Success 200 {object} map[string]string "Статус удаления"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "город не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /city/{id}/delete-city [delete]
func (h *Handler) DeleteCity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteCity(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "deleted",
	})
}

// ChangeCity godoc
// @Summary Изменить данные города
// @Description Обновляет информацию о городе по ID
// @Tags cites
// @Accept json
// @Produce json
// @Param id path int true "ID города"
// @Param city body apitypes.CityJSON true "Новые данные города"
// @Success 200 {object} apitypes.CityJSON "Обновленный город"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 404 {object} map[string]string "Город не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /city/{id}/change-city [put]
func (h *Handler) ChangeCity(ctx *gin.Context) {
	var cityJSON apitypes.CityJSON
	if err := ctx.BindJSON(&cityJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	city, err := h.Repository.ChangeCity(id, cityJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, apitypes.CityToJSON(city))
}

// AddCityToService godoc
// @Summary Добавить город в исследование
// @Description Добавляет город в черновик исследования пользователя
// @Tags cites
// @Produce json
// @Param id path int true "ID города"
// @Success 200 {object} apitypes.ServiceJSON "Cервис с добавленным городом"
// @Success 201 {object} apitypes.ServiceJSON "Создано новый сервис"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 404 {object} map[string]string "Город не найден"
// @Failure 409 {object} map[string]string "город уже в cервисе"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /city/{id}/add-to-service [post]
func (h *Handler) AddCityToService(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	service, created, err := h.Repository.GetServiceDraft(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	serviceId := service.ID

	cityId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.AddCityToService(int(serviceId), cityId)
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

	status := http.StatusOK

	if created {
		ctx.Header("Location", fmt.Sprintf("/research/%v", service.ID))
		status = http.StatusCreated
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(service)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(status, apitypes.ServiceToJSON(service, creatorLogin, moderatorLogin))
}

// UploadImage godoc
// @Summary Загрузить изображение для города
// @Description Загружает изображение для города и возвращает обновленные данные
// @Tags cites
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID города"
// @Param image formData file true "Изображение города"
// @Success 200 {object} map[string]interface{} "Статус загрузки и данные города"
// @Failure 400 {object} map[string]string "Неверный запрос или файл"
// @Failure 404 {object} map[string]string "Город не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /city/{id}/create-image [post]
func (h *Handler) UploadImage(ctx *gin.Context) {
	cityId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	city, err := h.Repository.UploadImage(ctx, cityId, file)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "uploaded",
		"planet": apitypes.CityToJSON(city),
	})
}
