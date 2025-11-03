package handler

import (
	"log/slog"

	"github.com/histopathai/auth-service/internal/service"
)

type SessionHandler struct {
	sessionService *service.ScopedSessionService
	BaseHandler    // Embed the BaseHandler
}

func NewSessionHandler(sessionService *service.ScopedSessionService, logger *slog.Logger) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		BaseHandler:    BaseHandler{logger: logger, response: &ResponseHelper{}},
	}
}
