package queries

const (
	CheckApiKeyQuery = `
			SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
			FROM triggerx.apikeys 
			WHERE owner = ? ALLOW FILTERING`

	InsertApiKeyQuery = `
			INSERT INTO triggerx.apikeys (key, owner, isActive, rateLimit, lastUsed, createdAt)
			VALUES (?, ?, ?, ?, ?, ?)`

	DeleteApiKeyQuery = `
			UPDATE triggerx.apikeys 
			SET isActive = ? 
			WHERE key = ?`

	UpdateApiKeyQuery = `
			UPDATE triggerx.apikeys 
			SET isActive = ?, rateLimit = ? 
			WHERE key = ?`

	SelectApiKeyQuery = `
			SELECT key, owner, isActive, rateLimit, lastUsed, createdAt 
			FROM triggerx.apikeys 
			WHERE key = ? ALLOW FILTERING`
)