package handlers

import (
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Handler struct update
type Handler struct {
	db *database.Connection
	logger logging.Logger
}

// NewHandler function update
func NewHandler(db *database.Connection, logger logging.Logger) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
	}
}
