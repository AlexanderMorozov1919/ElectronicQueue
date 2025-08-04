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
	Login        string `json:"login" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

type CreateDoctorRequest struct {
	FullName       string `json:"full_name" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
	Login          string `json:"login" binding:"required"`
	Password       string `json:"password" binding:"required"`
}

type CreateAdministratorRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
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
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        credentials body CreateRegistrarRequest true "Данные нового регистратора"
// @Success      201 {object} map[string]interface{} "Регистратор успешно создан"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      409 {object} map[string]string "Ошибка: логин уже занят"
// @Security     ApiKeyAuth
// @Router       /api/admin/create/registrar [post]
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
		"message":       "Регистратор успешно создан",
		"registrar_id":  registrar.RegistrarID,
		"login":         registrar.Login,
		"window_number": registrar.WindowNumber,
	})
}

// LoginDoctor обрабатывает аутентификацию врача
// @Summary      Аутентификация врача
// @Description  Принимает логин и пароль, возвращает JWT токен и информацию о враче.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "Учетные данные"
// @Success      200 {object} map[string]interface{} "Успешный ответ с токеном и данными врача"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      401 {object} map[string]string "Ошибка: неверные учетные данные"
// @Router       /api/auth/login/doctor [post]
func (h *AuthHandler) LoginDoctor(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	token, doctor, err := h.authService.AuthenticateDoctor(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"doctor": doctor,
	})
}

// CreateDoctor создает нового пользователя-врача.
// @Summary      Создать нового врача (Админ)
// @Description  Создает нового пользователя с ролью "врач". Требует INTERNAL_API_KEY.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        credentials body CreateDoctorRequest true "Данные нового врача"
// @Success      201 {object} map[string]interface{} "Врач успешно создан"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      409 {object} map[string]string "Ошибка: логин уже занят"
// @Security     ApiKeyAuth
// @Router       /api/admin/create/doctor [post]
func (h *AuthHandler) CreateDoctor(c *gin.Context) {
	var req CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	doctor, err := h.authService.CreateDoctor(req.FullName, req.Specialization, req.Login, req.Password)
	if err != nil {
		if err.Error() == "логин '"+req.Login+"' уже занят" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Врач успешно создан",
		"doctor_id":      doctor.ID,
		"login":          doctor.Login,
		"full_name":      doctor.FullName,
		"specialization": doctor.Specialization,
	})
}

// LoginAdministrator обрабатывает аутентификацию администратора
// @Summary      Аутентификация администратора
// @Description  Принимает логин и пароль, возвращает JWT токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginRequest true "Учетные данные"
// @Success      200 {object} map[string]string "Успешный ответ с токеном"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      401 {object} map[string]string "Ошибка: неверные учетные данные"
// @Router       /api/auth/login/administrator [post]
func (h *AuthHandler) LoginAdministrator(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	token, err := h.authService.AuthenticateAdministrator(req.Login, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// CreateAdministrator создает нового пользователя-администратора.
// @Summary      Создать нового администратора (Админ)
// @Description  Создает нового пользователя с ролью "администратор". Требует INTERNAL_API_KEY.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        credentials body CreateAdministratorRequest true "Данные нового администратора"
// @Success      201 {object} map[string]interface{} "Администратор успешно создан"
// @Failure      400 {object} map[string]string "Ошибка: неверный запрос"
// @Failure      409 {object} map[string]string "Ошибка: логин уже занят"
// @Security     ApiKeyAuth
// @Router       /api/admin/create/administrator [post]
func (h *AuthHandler) CreateAdministrator(c *gin.Context) {
	var req CreateAdministratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса: " + err.Error()})
		return
	}

	admin, err := h.authService.CreateAdministrator(req.FullName, req.Login, req.Password)
	if err != nil {
		if err.Error() == "логин '"+req.Login+"' уже занят" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "Администратор успешно создан",
		"administrator_id": admin.AdministratorID,
		"login":            admin.Login,
		"full_name":        admin.FullName,
	})
}
