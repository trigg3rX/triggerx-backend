package types

type CreateOrbitChainRequest struct {
	ChainID     int64  `json:"chain_id" binding:"required"`
	ChainName   string `json:"chain_name" binding:"required"`
	UserAddress string `json:"user_address" binding:"required"`
}

type OrbitChainData struct {
	ChainID             int64  `json:"chain_id"`
	ChainName           string `json:"chain_name"`
	RPCUrl              string `json:"rpc_url,omitempty"`
	UserAddress         string `json:"user_address,omitempty"`
	DeploymentStatus    string `json:"deployment_status,omitempty"`
	OrbitChainAddress   string `json:"orbit_chain_address,omitempty"`
}
