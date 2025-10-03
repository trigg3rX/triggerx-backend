package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/datastore/interfaces"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	logger                 logging.Logger
	config                 NotificationConfig
	dockerExecutor         dockerexecutor.DockerExecutorAPI
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
	config NotificationConfig,
	dockerExecutor dockerexecutor.DockerExecutorAPI,
	hub *websocket.Hub,
	publisher *events.Publisher,
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
		config:                 config,
		dockerExecutor:         dockerExecutor,
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

func (h *Handler) getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}
