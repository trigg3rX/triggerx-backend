package connection

import (
	"context"

	"github.com/gocql/gocql"
)

// ConnectionManager defines the interface for managing the database session.
// Services will depend on this to get a session for repositories.
type ConnectionManager interface {
	GetSession() Sessioner
	Close()
	HealthCheck(ctx context.Context) error
}

// Sessioner defines an interface for gocql.Session for easier mocking.
type Sessioner interface {
	Query(stmt string, values ...interface{}) *gocql.Query
	Close()
}
