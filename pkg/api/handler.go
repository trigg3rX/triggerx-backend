package api

import (
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

// Handler struct update
type Handler struct {
	db *database.Connection
}

// NewHandler function update
func NewHandler(db *database.Connection) *Handler {
	return &Handler{
		db: db,
	}
}
