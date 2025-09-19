package queries

// Create Queries
const (
	CreateOrbitChainQuery = `
		INSERT INTO triggerx.orbit_chain_data (
			chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address
		) VALUES (?, ?, ?, ?, ?, ?)`
)

// Read Queries
const (
	CheckOrbitChainByIDQuery = `
		SELECT chain_id FROM triggerx.orbit_chain_data WHERE chain_id = ?`

	CheckOrbitChainByNameQuery = `
		SELECT chain_name FROM triggerx.orbit_chain_data WHERE chain_name = ? ALLOW FILTERING`

	GetOrbitChainsByUserAddressQuery = `
		SELECT chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address FROM triggerx.orbit_chain_data WHERE user_address = ? ALLOW FILTERING`

	GetAllOrbitChainsQuery = `
		SELECT chain_id, chain_name, rpc_url, user_address, deployment_status, orbit_chain_address FROM triggerx.orbit_chain_data`
)
