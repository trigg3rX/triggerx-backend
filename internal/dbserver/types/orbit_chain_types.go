package types

import "time"

// CreateOrbitChainRequest represents the request structure for creating a new Orbit chain
type CreateOrbitChainRequest struct {
	ChainID     int64  `json:"chain_id" binding:"required"`
	ChainName   string `json:"chain_name" binding:"required"`
	UserAddress string `json:"user_address" binding:"required"`
}

// DeployChainRequest represents the request structure for deploying a new Orbit chain
type DeployChainRequest struct {
	ChainID      int64  `json:"chain_id" binding:"required"`
	ChainName    string `json:"chain_name" binding:"required"`
	OwnerAddress string `json:"owner_address" binding:"required"`
	BatchPoster  string `json:"batch_poster" binding:"required"`
	Validator    string `json:"validator" binding:"required"`
	UserAddress  string `json:"user_address" binding:"required"`

	// ERC20 Token Details (if custom gas token)
	NativeToken   string `json:"native_token,omitempty"` // Address of ERC20 token, empty for ETH
	TokenName     string `json:"token_name,omitempty"`
	TokenSymbol   string `json:"token_symbol,omitempty"`
	TokenDecimals int    `json:"token_decimals,omitempty"`

	// Optional Chain Configuration
	MaxDataSize               int    `json:"max_data_size,omitempty"`
	MaxFeePerGasForRetryables string `json:"max_fee_per_gas_for_retryables,omitempty"`
}

// ChainDeploymentResponse represents the response structure for chain deployment
type ChainDeploymentResponse struct {
	Success      bool   `json:"success"`
	DeploymentID string `json:"deployment_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	ChainAddress string `json:"chain_address,omitempty"`
}

// OrbitChainData represents the orbit chain data structure matching the database schema
type OrbitChainData struct {
	ChainID           int64     `json:"chain_id"`
	ChainName         string    `json:"chain_name"`
	RPCUrl            *string   `json:"rpc_url,omitempty"`
	UserAddress       string    `json:"user_address"`
	DeploymentStatus  string    `json:"deployment_status"`
	OrbitChainAddress string    `json:"orbit_chain_address,omitempty"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// ChainDeploymentStatus represents the status of a chain deployment
type ChainDeploymentStatus struct {
	ChainID           int64     `json:"chain_id"`
	ChainName         string    `json:"chain_name"`
	UserAddress       string    `json:"user_address"`
	DeploymentStatus  string    `json:"deployment_status"`
	OrbitChainAddress string    `json:"orbit_chain_address,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	ErrorMessage      string    `json:"error_message,omitempty"`
}
