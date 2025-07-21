package handlers

import (
	"ElectronicQueue/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateRegistrarRequest struct {
	WindowNumber int    `json:"window_number" binding:"required"`
	Login       string `json:"login" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// LoginRegistrar обрабатывает аутентификацию регистратора
// @Summary      Аутентификация регистратора
// @Description  Принимает логин и пароль, возвращает JWT токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "Учетные данные"
// @Success      200 {object} map[string]string "Успешный ответ с токеном"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      401 {object} map[string]string "Ошибка: неверные учетные данные"
// @Router       /api/auth/login/registrar [post]
func (h *AuthHandler) LoginRegistrar(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	token, err := h.authService.AuthenticateRegistrar(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// CreateRegistrar создает нового пользователя-регистратора.
// @Summary      Создать нового регистратора (Админ)
// @Description  Создает нового пользователя с ролью "регистратор". Требует INTERNAL_API_KEY.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body CreateRegistrarRequest true "Данные нового регистратора"
// @Success      201 {object} map[string]interface{} "Регистратор успешно создан"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      409 {object} map[string]string "Ошибка: логин уже занят"
// @Security     ApiKeyAuth
// @Router       /api/auth/create/registrar [post]
func (h *AuthHandler) CreateRegistrar(c *gin.Context) {
	var req CreateRegistrarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	registrar, err := h.authService.CreateRegistrar(req.WindowNumber, req.Login, req.Password)
	if err != nil {
		// Проверяем, является ли ошибка конфликтом (логин занят)
		if err.Error() == "логин '"+req.Login+"' уже занят" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Регистратор успешно создан",
		"registrar_id": registrar.RegistrarID,
		"login":        registrar.Login,
		"window_number": registrar.WindowNumber,
	})
}
