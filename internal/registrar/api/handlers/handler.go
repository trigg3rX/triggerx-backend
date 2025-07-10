package handlers

import (
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Handler struct {
	logger logging.Logger
}

func NewHandler(logger logging.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}
