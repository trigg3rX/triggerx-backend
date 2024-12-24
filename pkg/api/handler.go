package api

import (
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/eventbus"
)

// Handler struct update
type Handler struct {
	db       *database.Connection
	eventBus *eventbus.EventBus
}

// NewHandler function update
func NewHandler(db *database.Connection, eb *eventbus.EventBus) *Handler {
	return &Handler{
		db:       db,
		eventBus: eb,
	}
}