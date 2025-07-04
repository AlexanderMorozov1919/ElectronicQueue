package handlers

import (
	"ElectronicQueue/internal/services"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Словарь для отображения названий услуг
var ServiceNameMap = map[string]string{
	"make_appointment":    "Запись на приём",
	"confirm_appointment": "Приём по записи",
	"lab_tests":           "Анализы",
	"documents":           "Документы",
}

// TicketHandler содержит зависимости для работы с талонами
type TicketHandler struct {
	Service *services.TicketService
}

// NewTicketHandler создает новый TicketHandler
func NewTicketHandler(service *services.TicketService) *TicketHandler {
	return &TicketHandler{Service: service}
}

// GetServicePage - /terminal/service (GET)
func (h *TicketHandler) GetServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "index.html"))
}

// GetSelectServicePage - /terminal/service/select (GET)
func (h *TicketHandler) GetSelectServicePage(c *gin.Context) {
	c.File(filepath.Join("frontend", "select.html"))
}

// HandleService - обработчик для создания талона
func (h *TicketHandler) HandleService(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(302, "/terminal/service/print_ticket?service="+service)
	}
}

// HandlePrintTicketPage - промежуточная страница подтверждения печати
func (h *TicketHandler) HandlePrintTicketPage(c *gin.Context) {
	service := c.Query("service")
	serviceName := ServiceNameMap[service]
	if serviceName == "" {
		serviceName = service
	}
	c.HTML(200, "print_ticket.html", gin.H{
		"Service":     service,
		"ServiceName": serviceName,
	})
}

// HandleDisplayTicketPage - страница отображения талона
func (h *TicketHandler) HandleDisplayTicketPage(c *gin.Context) {
	service := c.Query("service")
	ticketNumber := c.Query("ticket")
	serviceName := ServiceNameMap[service]
	if serviceName == "" {
		serviceName = service
	}

	isPrinted := ticketNumber != ""

	var title string
	if isPrinted {
		title = "Ваш электронный талон"
	} else {
		title = "Возьмите талон"
	}

	c.HTML(200, "display_ticket.html", gin.H{
		"ServiceName":  serviceName,
		"TicketNumber": ticketNumber,
		"IsPrinted":    isPrinted,
		"Title":        title,
	})
}

// HandleDisplayTicketPost обрабатывает POST и делает редирект на GET.
func (h *TicketHandler) HandleDisplayTicketPost(c *gin.Context) {
	service := c.PostForm("service")
	print := c.PostForm("print")

	if print == "yes" {
		// Заглушка для печати талона
		// printStub(service)
	}

	ticket, err := h.Service.CreateTicket(service)
	if err != nil {
		c.String(500, "Ошибка создания талона: %v", err)
		return
	}
	ticketNumber := ticket.TicketNumber
	c.Redirect(302, "/terminal/service/display_ticket?service="+service+"&ticket="+ticketNumber)
}
