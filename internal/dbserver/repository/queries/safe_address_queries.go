package queries

// Create Query
const (
	CreateSafeAddressQuery = `
		INSERT INTO triggerx.safe_addresses (
			user_address, safe_address, created_at
		) VALUES (?, ?, ?)`
)

// Read Queries
const (
	GetSafeAddressesByUserQuery = `
		SELECT safe_address, created_at
		FROM triggerx.safe_addresses
		WHERE user_address = ?`

	CheckSafeAddressExistsQuery = `
		SELECT user_address, safe_address
		FROM triggerx.safe_addresses
		WHERE user_address = ? AND safe_address = ?`
)
