package queries

// Create Queries
const (
	CreateApiKeyQuery = `
			INSERT INTO triggerx.apikeys (
			key, owner, is_active, rate_limit, last_used, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`
)

// Update Queries
const (
	UpdateApiKeyQuery = `
			UPDATE triggerx.apikeys 
			SET is_active = ?, rate_limit = ? 
			WHERE key = ?`

	UpdateApiKeyStatusQuery = `
			UPDATE triggerx.apikeys 
			SET is_active = ? 
			WHERE key = ?`

	UpdateApiKeyLastUsedQuery = `
			UPDATE triggerx.apikeys 
			SET last_used = ?, success_count = ?, failed_count = ? 
			WHERE key = ?`
)

// Read Queries
const (
	GetApiKeyDataByOwnerQuery = `
			SELECT key, owner, is_active, rate_limit, success_count, failed_count, last_used, created_at 
			FROM triggerx.apikeys 
			WHERE owner = ? ALLOW FILTERING`

	GetApiKeyDataByApiKeyQuery = `
			SELECT key, owner, is_active, rate_limit, success_count, failed_count, last_used, created_at 
			FROM triggerx.apikeys 
			WHERE key = ?`

	GetApiKeyCallCountQuery = `
			SELECT success_count, failed_count 
			FROM triggerx.apikeys 
			WHERE key = ?`

	GetApiKeyByOwnerQuery = `
			SELECT key
			FROM triggerx.apikeys 
			WHERE owner = ? ALLOW FILTERING`

	GetApiOwnerByApiKeyQuery = `
			SELECT owner
			FROM triggerx.apikeys 
			WHERE key = ?`
)
