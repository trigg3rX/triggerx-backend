package handlers

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/events"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/redis"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/database/repository"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/rpc/clients/conditionscheduler"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/websocket"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/dockerexecutor"
	"github.com/trigg3rX/triggerx-backend/pkg/http"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db                       *database.Connection
	logger                   observability.Logger
	tracer                   observability.Tracer
	config                   NotificationConfig
	dockerExecutor           dockerexecutor.DockerExecutorAPI
	jobRepository            repository.JobRepository
	timeJobRepository        repository.TimeJobRepository
	eventJobRepository       repository.EventJobRepository
	conditionJobRepository   repository.ConditionJobRepository
	taskRepository           repository.TaskRepository
	userRepository           repository.UserRepository
	keeperRepository         repository.KeeperRepository
	apiKeysRepository        repository.ApiKeysRepository
	safeAddressRepository    repository.SafeAddressRepository
	httpClient               http.HTTPClientInterface
	redisClient              *redis.Client
	conditionSchedulerClient *conditionscheduler.Client
	// WebSocket components
	hub       *websocket.Hub
	publisher *events.Publisher

	scanNowQuery func(*time.Time) error // for testability
}

func NewHandler(db *database.Connection, logger observability.Logger, tracer observability.Tracer, config NotificationConfig, dockerExecutor dockerexecutor.DockerExecutorAPI, hub *websocket.Hub, publisher *events.Publisher, httpClient http.HTTPClientInterface, redisClient *redis.Client, conditionSchedulerClient *conditionscheduler.Client) *Handler {
	h := &Handler{
		db:                       db,
		logger:                   logger,
		tracer:                   tracer,
		config:                   config,
		dockerExecutor:           dockerExecutor,
		jobRepository:            repository.NewJobRepository(db),
		timeJobRepository:        repository.NewTimeJobRepository(db),
		eventJobRepository:       repository.NewEventJobRepository(db),
		conditionJobRepository:   repository.NewConditionJobRepository(db),
		taskRepository:           repository.NewTaskRepository(db),
		userRepository:           repository.NewUserRepository(db),
		keeperRepository:         repository.NewKeeperRepository(db),
		apiKeysRepository:        repository.NewApiKeysRepository(db),
		safeAddressRepository:    repository.NewSafeAddressRepository(db),
		hub:                      hub,
		publisher:                publisher,
		httpClient:               httpClient,
		redisClient:              redisClient,
		conditionSchedulerClient: conditionSchedulerClient,
	}
	h.scanNowQuery = h.defaultScanNowQuery

	return h
}

func (h *Handler) defaultScanNowQuery(timestamp *time.Time) error {
	return h.db.Session().Query("SELECT now() FROM system.local").Scan(timestamp)
}
