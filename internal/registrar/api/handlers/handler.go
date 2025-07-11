package handlers

import (
	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
)

type Handler struct {
	logger logging.Logger
}

func NewHandler(logger logging.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}
