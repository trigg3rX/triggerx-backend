package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type Handler struct {
	logger                 logging.Logger
	dockerExecutor         dockerexecutor.DockerExecutorAPI
	datastore              datastore.DatastoreService
	jobRepository          interfaces.GenericRepository[types.JobDataEntity]
	timeJobRepository      interfaces.GenericRepository[types.TimeJobDataEntity]
	eventJobRepository     interfaces.GenericRepository[types.EventJobDataEntity]
	conditionJobRepository interfaces.GenericRepository[types.ConditionJobDataEntity]
	taskRepository         interfaces.GenericRepository[types.TaskDataEntity]
	userRepository         interfaces.GenericRepository[types.UserDataEntity]
	keeperRepository       interfaces.GenericRepository[types.KeeperDataEntity]
	apiKeysRepository      interfaces.GenericRepository[types.ApiKeyDataEntity]

	// WebSocket components
	hub       *websocket.Hub
	publisher *events.Publisher
}

func NewHandler(
	logger logging.Logger,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
	hub *websocket.Hub,
	publisher *events.Publisher,
	datastore datastore.DatastoreService,
	jobRepository interfaces.GenericRepository[types.JobDataEntity],
	timeJobRepository interfaces.GenericRepository[types.TimeJobDataEntity],
	eventJobRepository interfaces.GenericRepository[types.EventJobDataEntity],
	conditionJobRepository interfaces.GenericRepository[types.ConditionJobDataEntity],
	taskRepository interfaces.GenericRepository[types.TaskDataEntity],
	userRepository interfaces.GenericRepository[types.UserDataEntity],
	keeperRepository interfaces.GenericRepository[types.KeeperDataEntity],
	apiKeysRepository interfaces.GenericRepository[types.ApiKeyDataEntity],
) *Handler {
	h := &Handler{
		logger:                 logger,
		dockerExecutor:         dockerExecutor,
		datastore:              datastore,
		jobRepository:          jobRepository,
		timeJobRepository:      timeJobRepository,
		eventJobRepository:     eventJobRepository,
		conditionJobRepository: conditionJobRepository,
		taskRepository:         taskRepository,
		userRepository:         userRepository,
		keeperRepository:       keeperRepository,
		apiKeysRepository:      apiKeysRepository,
		hub:                    hub,
		publisher:              publisher,
	}

	return h
}

// getLogger retrieves the traced logger from context
// This logger already has the traceID attached
func (h *Handler) getLogger(c *gin.Context) logging.Logger {
	logger, exists := c.Get("logger")
	if !exists {
		// Fallback to base logger if context logger not found
		return h.logger
	}
	return logger.(logging.Logger)
}
