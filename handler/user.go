package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	apitypes "lab4-swag/internal/app/api_types"
	"lab4-swag/internal/app/repository"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// SignUp
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя. Возвращает URL созданного ресурса в Location и тело созданного пользователя.
// @Tags users
// @Accept json
// @Produce json
// @Param user body apitypes.UserJSON true "Параметры нового пользователя"
// @Success 201 {object} apitypes.UserJSON "Пользователь создан"
// @Failure 400 {object} map[string]string "Ошибка валидации или входных данных"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /users/sign-up [post]
func (h *Handler) SignUp(ctx *gin.Context) {
	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.CreateUser(userJSON)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.Header("Location", fmt.Sprintf("/users/%v", user.Login))
	ctx.JSON(http.StatusCreated, apitypes.UserToJSON(user))
}

// SignIn
// @Summary Вход (получение токена)
// @Description Принимает логин/пароль, возвращает jwt-токен в формате {"token":"..."}.
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body apitypes.UserJSON true "Логин и пароль"
// @Success 200 {object} map[string]string "token"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /users/sign-in [post]
func (h *Handler) SignIn(ctx *gin.Context) {
	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	token, err := h.Repository.SignIn(userJSON)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// GetProfile
// @Summary Получить профиль пользователя
// @Description Возвращает данные профиля (доступен только тот, чей UUID совпадает с user_id в токене).
// @Tags users
// @Produce json
// @Param login path string true "Логин пользователя"
// @Success 200 {object} apitypes.UserJSON "Профиль пользователя"
// @Failure 400 {object} map[string]string "Проблема с авторизацией/получением user_id"}
// @Failure 403 {object} map[string]string "Пользователи не совпадают"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/{login}/profile [get]
func (h *Handler) GetProfile(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	login := ctx.Param("login")

	user, err := h.Repository.GetUserByLogin(login)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if user.ID != userID {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("users do not match"))
		return
	}

	user.Password = ""

	ctx.JSON(http.StatusOK, apitypes.UserToJSON(user))
}

// ChangeProfile
// @Summary Изменить профиль пользователя
// @Description Обновляет профиль пользователя (может делать только сам пользователь).
// @Tags users
// @Accept json
// @Produce json
// @Param login path string true "Логин пользователя"
// @Param user body apitypes.UserJSON true "Новые данные профиля"
// @Success 200 {object} apitypes.UserJSON "Обновлённый профиль"
// @Failure 400 {object} map[string]string "Ошибка запроса или авторизации"
// @Failure 403 {object} map[string]string "Доступ запрещён"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /users/{login}/profile [put]
func (h *Handler) ChangeProfile(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	login := ctx.Param("login")

	var userJSON apitypes.UserJSON
	if err := ctx.BindJSON(&userJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByLogin(login)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if user.ID != userID {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	user, err = h.Repository.ChangeProfile(login, userJSON)
	if err == repository.ErrNotFound {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, apitypes.UserToJSON(user))
}

// SignOut
// @Summary Выход (удаление токена)
// @Description Удаляет токен текущего пользователя из хранилища. Возвращает {"status":"signed_out"}.
// @Tags users
// @Produce json
// @Success 200 {object} map[string]string "status"
// @Failure 400 {object} map[string]string "Проблема с получением user_id"
// @Failure 500 {object} map[string]string "Внутренняя ошибка при удалении токена"
// @Security ApiKeyAuth
// @Router /users/sign-out [post]
func (h *Handler) SignOut(ctx *gin.Context) {
	tokenString := extractTokenFromHeader(ctx.Request)
	if tokenString == "" {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("no token provided"))
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || token == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("invalid token claims"))
		return
	}

	ttl, err := tokenTTLFromClaims(claims)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
		return
	}

	err = h.Repository.AddTokenToBlacklist(context.Background(), tokenString, ttl)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
}

func getUserID(ctx *gin.Context) (uuid.UUID, error) {
	userIDStr, exits := ctx.Get("user_id")
	if !exits {
		return uuid.UUID{}, errors.New("user_id not found")
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.UUID{}, err
	}
	return userID, nil
}

func tokenTTLFromClaims(claims jwt.MapClaims) (time.Duration, error) {
	expVal, ok := claims["exp"]
	if !ok {
		return 0, errors.New("exp not present")
	}

	var expUnix int64
	switch v := expVal.(type) {
	case float64:
		expUnix = int64(v)
	case int64:
		expUnix = v
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, err
		}
		expUnix = i
	default:
		return 0, fmt.Errorf("unsupported exp type %T", v)
	}

	expTime := time.Unix(expUnix, 0)
	ttl := time.Until(expTime)
	if ttl < 0 {
		return 0, errors.New("token already expired")
	}
	return ttl, nil
}
