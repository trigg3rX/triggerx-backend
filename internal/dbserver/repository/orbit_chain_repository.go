package repository

import (
	"errors"
	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type OrbitChainRepository interface {
	CreateOrbitChain(chain *types.CreateOrbitChainRequest) error
	GetOrbitChainsByUserAddress(userAddress string) ([]types.OrbitChainData, error)
	GetAllOrbitChains() ([]types.OrbitChainData, error)
}

type orbitChainRepository struct {
	db *database.Connection
}

func NewOrbitChainRepository(db *database.Connection) OrbitChainRepository {
	return &orbitChainRepository{db: db}
}

func (r *orbitChainRepository) CreateOrbitChain(chain *types.CreateOrbitChainRequest) error {
	// Check for duplicate chain_id
	var existingID int64
	err := r.db.Session().Query(queries.CheckOrbitChainByIDQuery, chain.ChainID).Scan(&existingID)
	if err != nil && err != gocql.ErrNotFound {
		return err
	}
	if err == nil {
		return errors.New("chain_id already exists")
	}
	// Check for duplicate chain_name
	var existingName string
	err = r.db.Session().Query(queries.CheckOrbitChainByNameQuery, chain.ChainName).Scan(&existingName)
	if err != nil && err != gocql.ErrNotFound {
		return err
	}
	if err == nil {
		return errors.New("chain_name already exists")
	}
	// Insert new chain with rpc_url as nil, user_address from request, deployment_status and orbit_chain_address as empty string
	return r.db.Session().Query(queries.CreateOrbitChainQuery, chain.ChainID, chain.ChainName, nil, chain.UserAddress, "", "").Exec()
}

func (r *orbitChainRepository) GetOrbitChainsByUserAddress(userAddress string) ([]types.OrbitChainData, error) {
	iter := r.db.Session().Query(queries.GetOrbitChainsByUserAddressQuery, userAddress).Iter()
	var chains []types.OrbitChainData
	var chain types.OrbitChainData
	for iter.Scan(&chain.ChainID, &chain.ChainName, &chain.RPCUrl, &chain.UserAddress, &chain.DeploymentStatus, &chain.OrbitChainAddress) {
		chains = append(chains, chain)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return chains, nil
}

func (r *orbitChainRepository) GetAllOrbitChains() ([]types.OrbitChainData, error) {
	iter := r.db.Session().Query(queries.GetAllOrbitChainsQuery).Iter()
	var chains []types.OrbitChainData
	var chain types.OrbitChainData
	for iter.Scan(&chain.ChainID, &chain.ChainName, &chain.RPCUrl, &chain.UserAddress, &chain.DeploymentStatus, &chain.OrbitChainAddress) {
		chains = append(chains, chain)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return chains, nil
}
