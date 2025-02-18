package handlers

import (
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler struct update
type Handler struct {
	db *config.Connection
	logger logging.Logger
}

// NewHandler function update
func NewHandler(db *config.Connection, logger logging.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}
