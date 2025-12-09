package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/ds"
	"lab4-swag/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetServices godoc
// @Summary Получить список сервисов
// @Description Возвращает сервисов с возможностью фильтрации по датам и статусу
// @Tags services
// @Produce json
// @Param from-date query string false "Начальная дата (YYYY-MM-DD)"
// @Param to-date query string false "Конечная дата (YYYY-MM-DD)"
// @Param status query string false "Статус сервиса"
// @Success 200 {array} apitypes.ServiceJSON "Список сервиса"
// @Failure 400 {object} map[string]string "Неверный формат даты"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /services [get]
func (h *Handler) GetServices(ctx *gin.Context) {
	fromDate := ctx.Query("from-date")
	var from = time.Time{}
	var to = time.Time{}
	if fromDate != "" {
		from1, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		from = from1
	}

	toDate := ctx.Query("to-date")
	if toDate != "" {
		to1, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		to = to1
	}

	status := ctx.Query("status")

	services, err := h.Repository.GetServices(from, to, status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	services = h.filterServicesByAuth(services, ctx)
	resp := make([]apitypes.ServiceJSON, 0, len(services))
	for _, c := range services {
		creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(c)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		resp = append(resp, apitypes.ServiceToJSON(c, creatorLogin, moderatorLogin))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetServiceCart godoc
// @Summary Получить корзину сервисов
// @Description Возвращает информацию о текущем черновике сервиса пользователя
// @Tags services
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные корзины сервиса"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/service-cart [get]
func (h *Handler) GetServiceCart(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	citesCount := h.Repository.GetServiceCount(userID)

	if citesCount == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status":        "no_draft",
			"planets_count": citesCount,
		})
		return
	}

	service, err := h.Repository.CheckCurrentServiceDraft(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusUnauthorized, err)
		} else if errors.Is(err, repository.ErrNoDraft) {
			ctx.JSON(http.StatusOK, gin.H{
				"status":        "no_draft",
				"planets_count": 0,
			})
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":            service.ID,
		"planets_count": h.Repository.GetServiceCount(service.CreatorID),
	})
}

// GetService godoc
// @Summary Получить сервис по ID
// @Description Возвращает полную информацию об сервисе включая города
// @Tags services
// @Produce json
// @Param id path int true "ID сервиса"
// @Success 200 {object} map[string]interface{} "Данные сервиса с городами"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Сервис не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/{id} [get]
func (h *Handler) GetService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	cites, service, err := h.Repository.GetServiceCites(id)
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

	resp := make([]apitypes.CityJSON, 0, len(cites))
	for _, r := range cites {
		resp = append(resp, apitypes.CityToJSON(r))
	}

	creatorLogin, moderatorLogin, err := h.Repository.GetModeratorAndCreatorLogin(service)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	citesService, _ := h.Repository.GetCitesServices(service.ID)

	resp2 := make([]apitypes.CitesServiceJSON, 0, len(citesService))
	for _, r := range citesService {
		resp2 = append(resp2, apitypes.CitesServiceToJSON(r))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"service":      apitypes.ServiceToJSON(service, creatorLogin, moderatorLogin),
		"cites":        resp,
		"citesService": resp2,
	})
}

// FormService godoc
// @Summary Сформировать сервис
// @Description Переводит сервис в статус "formed"
// @Tags services
// @Produce json
// @Param id path int true "ID cервис"
// @Success 200 {object} apitypes.ServiceJSON "Сформированный сервис"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Сервис не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/{id}/form [put]
func (h *Handler) FormService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := "formed"

	service, err := h.Repository.FormService(id, status)
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

// ChangeService godoc
// @Summary Изменить сервис
// @Description Обновляет данные сервиса
// @Tags servicees
// @Accept json
// @Produce json
// @Param id path int true "ID сервиса"
// @Param service body apitypes.ServiceJSON true "Новые данные сервиса"
// @Success 200 {object} apitypes.ServiceJSON "Обновленный сервис"
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 404 {object} map[string]string "Сервис не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/{id}/change-service [put]
func (h *Handler) ChangeService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var serviceJSON apitypes.ServiceJSON
	if err := ctx.BindJSON(&serviceJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	service, err := h.Repository.ChangeService(id, serviceJSON)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
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

// DeleteService godoc
// @Summary Удалить сервис
// @Description Выполняет логическое удаление сервиса
// @Tags services
// @Produce json
// @Param id path int true "ID сервиса"
// @Success 200 {object} map[string]string "Статус удаления"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Сервис не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/{id}/delete-service [delete]
func (h *Handler) DeleteService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	serviceId, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	status := "deleted"

	_, err = h.Repository.FormService(serviceId, status)
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

	ctx.JSON(http.StatusOK, gin.H{"message": "Research deleted"})
}

// ModerateService godoc
// @Summary Модерировать сервиса
// @Description Изменяет статус сервиса (только для модераторов)
// @Tags services
// @Accept json
// @Produce json
// @Param id path int true "ID сервиса"
// @Param status body apitypes.StatusJSON true "Новый статус"
// @Success 200 {object} apitypes.ServiceJSON "Результат модерации"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Сервис не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /service/{id}/finish [put]
func (h *Handler) ModerateService(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var statusJSON apitypes.StatusJSON
	if err := ctx.BindJSON(&statusJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	if !user.IsModerator {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("требуются права модератора"))
		return
	}

	service, err := h.Repository.ModerateService(id, statusJSON.Status, userID)
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

	// Если сервис помечен как completed — триггерим асинхронный расчёт
	if statusJSON.Status == "completed" {
		// fire-and-forget POST to async service
		go func(sid int) {
			if h.Config == nil || h.Config.AsyncServiceURL == "" {
				return
			}
			payload := map[string]int{"service_id": sid}
			b, _ := json.Marshal(payload)
			url := h.Config.AsyncServiceURL
			if url[len(url)-1] == '/' {
				url = url[:len(url)-1]
			}
			postURL := url + "/calculate/"
			req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(b))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 3 * time.Second}
			_, _ = client.Do(req)
		}(service.ID)
	}
}

// SetServiceCost godoc
// @Summary Установить стоимость сервиса (callback от async-сервиса)
// @Description Принимает JSON {"cost": <float>} и обновляет поле cost у заявки
// @Accept json
// @Produce json
// @Param id path int true "ID сервиса"
// @Success 200 {object} map[string]string "ok"
// @Failure 400 {object} map[string]string "bad request"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 500 {object} map[string]string "internal error"
// @Router /road/{id}/set-cost [put]
func (h *Handler) SetServiceCost(ctx *gin.Context) {
	// Проверяем токен
	token := ctx.GetHeader("X-Auth-Token")
	if token != "SECRET_KEY_LAB8" {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var body struct {
		Cost float64 `json:"cost"`
	}
	if err := ctx.BindJSON(&body); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.SetServiceCost(id, body.Cost); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) filterServicesByAuth(services []ds.ServiceCalc, ctx *gin.Context) []ds.ServiceCalc {
	userID, err := getUserID(ctx)
	if err != nil {
		return []ds.ServiceCalc{}
	}

	user, err := h.Repository.GetUserByID(userID)
	if err == repository.ErrNotFound {
		return []ds.ServiceCalc{}
	}
	if err != nil {
		return []ds.ServiceCalc{}
	}

	if user.IsModerator {
		return services
	}

	var userServices []ds.ServiceCalc
	for _, service := range services {
		fmt.Println(service.ID)
		if service.CreatorID == userID {
			userServices = append(userServices, service)
		}
	}

	return userServices

}

func (h *Handler) hasAccessToResearch(creatorID uuid.UUID, ctx *gin.Context) bool {
	userID, err := getUserID(ctx)
	if err != nil {
		return false
	}

	user, err := h.Repository.GetUserByID(userID)
	if err == repository.ErrNotFound {
		return false
	}
	if err != nil {
		return false
	}

	return creatorID == userID || user.IsModerator
}
