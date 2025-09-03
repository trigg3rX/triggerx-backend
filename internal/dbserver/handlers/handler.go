package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/docker"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db                     *database.Connection
	logger                 logging.Logger
	config                 NotificationConfig
	dockerManager          *docker.DockerManager
	jobRepository          repository.JobRepository
	timeJobRepository      repository.TimeJobRepository
	eventJobRepository     repository.EventJobRepository
	conditionJobRepository repository.ConditionJobRepository
	taskRepository         repository.TaskRepository
	userRepository         repository.UserRepository
	keeperRepository       repository.KeeperRepository
	apiKeysRepository      repository.ApiKeysRepository

	// WebSocket components
	hub       *websocket.Hub
	publisher *events.Publisher

	scanNowQuery func(*time.Time) error // for testability
}

func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig, dockerManager *docker.DockerManager, hub *websocket.Hub) *Handler {
	h := &Handler{
		db:                     db,
		logger:                 logger,
		config:                 config,
		dockerManager:          dockerManager,
		jobRepository:          repository.NewJobRepository(db),
		timeJobRepository:      repository.NewTimeJobRepository(db),
		eventJobRepository:     repository.NewEventJobRepository(db),
		conditionJobRepository: repository.NewConditionJobRepository(db),
		taskRepository:         repository.NewTaskRepository(db),
		userRepository:         repository.NewUserRepository(db),
		keeperRepository:       repository.NewKeeperRepository(db),
		apiKeysRepository:      repository.NewApiKeysRepository(db),
		hub:                    hub,
		publisher:              events.NewPublisher(hub, logger),
	}
	h.scanNowQuery = h.defaultScanNowQuery

	// Log Docker manager status
	if dockerManager != nil {
		if dockerManager.IsInitialized() {
			logger.Info("Docker manager is initialized and ready")
			supportedLanguages := dockerManager.GetSupportedLanguages()
			logger.Infof("Supported languages: %v", supportedLanguages)
		} else {
			logger.Warn("Docker manager is not initialized")
		}
	} else {
		logger.Warn("Docker manager is nil")
	}

	return h
}

// NewHandlerWithPublisher creates a new handler with WebSocket-enabled repositories
func NewHandlerWithPublisher(db *database.Connection, logger logging.Logger, config NotificationConfig, dockerManager *docker.DockerManager, hub *websocket.Hub, publisher *events.Publisher) *Handler {
	h := &Handler{
		db:                     db,
		logger:                 logger,
		config:                 config,
		dockerManager:          dockerManager,
		jobRepository:          repository.NewJobRepository(db),
		timeJobRepository:      repository.NewTimeJobRepository(db),
		eventJobRepository:     repository.NewEventJobRepository(db),
		conditionJobRepository: repository.NewConditionJobRepository(db),
		taskRepository:         repository.NewTaskRepositoryWithPublisher(db, publisher),
		userRepository:         repository.NewUserRepository(db),
		keeperRepository:       repository.NewKeeperRepository(db),
		apiKeysRepository:      repository.NewApiKeysRepository(db),
		hub:                    hub,
		publisher:              publisher,
	}
	h.scanNowQuery = h.defaultScanNowQuery

	// Log Docker manager status
	if dockerManager != nil {
		if dockerManager.IsInitialized() {
			logger.Info("Docker manager is initialized and ready")
			supportedLanguages := dockerManager.GetSupportedLanguages()
			logger.Infof("Supported languages: %v", supportedLanguages)
		} else {
			logger.Warn("Docker manager is not initialized")
		}
	} else {
		logger.Warn("Docker manager is nil")
	}

	return h
}

func (h *Handler) defaultScanNowQuery(timestamp *time.Time) error {
	return h.db.Session().Query("SELECT now() FROM system.local").Scan(timestamp)
}

func (h *Handler) getTraceID(c *gin.Context) string {
	traceID, exists := c.Get("trace_id")
	if !exists {
		return ""
	}
	return traceID.(string)
}
