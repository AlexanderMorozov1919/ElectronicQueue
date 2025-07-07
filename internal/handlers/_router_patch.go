package main

import (
	"ElectronicQueue/internal/config"
	"ElectronicQueue/internal/middleware"
)

// ...existing code...
// Вставить в setupRouter после создания ticketHandler
// JWTManager для middleware
cfg, _ := config.LoadConfig()
jwtManager, _ := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)

registrar := r.Group("/api/tickets", middleware.RequireRole(jwtManager, "регистратор"))
{
	registrar.POST(":id/status", ticketHandler.UpdateStatus)
	registrar.DELETE(":id", ticketHandler.DeleteTicket)
}
// ...existing code...
