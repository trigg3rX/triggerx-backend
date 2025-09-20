package queries

// Create Queries
const (
	CreateOrbitChainQuery = `
		INSERT INTO triggerx.orbit_chain_data (
			chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
)

// Read Queries
const (
	CheckOrbitChainByIDQuery = `
		SELECT chain_id FROM triggerx.orbit_chain_data WHERE chain_id = ?`

	CheckOrbitChainByNameQuery = `
		SELECT chain_name FROM triggerx.orbit_chain_data WHERE chain_name = ? ALLOW FILTERING`

	GetOrbitChainsByUserAddressQuery = `
		SELECT chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address, created_at, updated_at 
		FROM triggerx.orbit_chain_data WHERE user_address = ? ALLOW FILTERING`

	GetAllOrbitChainsQuery = `
		SELECT chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address, created_at, updated_at 
		FROM triggerx.orbit_chain_data`

	GetOrbitChainByIDQuery = `
		SELECT chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address, created_at, updated_at 
		FROM triggerx.orbit_chain_data WHERE chain_id = ?`
)

// Update Queries
const (
	UpdateOrbitChainStatusQuery = `
		UPDATE triggerx.orbit_chain_data SET 
			deployment_status = ?, orbit_chain_address = ?, updated_at = ?
			WHERE chain_id = ?`

	UpdateOrbitChainRPCUrlQuery = `
		UPDATE triggerx.orbit_chain_data SET 
			rpc_url = ?, updated_at = ?
			WHERE chain_id = ?`
)
