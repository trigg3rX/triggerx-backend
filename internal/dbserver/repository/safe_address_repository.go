package repository

import (
	"errors"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

type SafeAddressRepository interface {
	CreateSafeAddress(userAddress string, safeAddress string) error
	GetSafeAddressesByUser(userAddress string) ([]commonTypes.SafeAddress, error)
	CheckSafeAddressExists(userAddress string, safeAddress string) (bool, error)
}

type safeAddressRepository struct {
	db *database.Connection
}

func NewSafeAddressRepository(db *database.Connection) SafeAddressRepository {
	return &safeAddressRepository{
		db: db,
	}
}

func (r *safeAddressRepository) CreateSafeAddress(userAddress string, safeAddress string) error {
	err := r.db.Session().Query(queries.CreateSafeAddressQuery,
		userAddress, safeAddress, time.Now()).Exec()
	if err != nil {
		return errors.New("failed to create safe address")
	}
	return nil
}

func (r *safeAddressRepository) GetSafeAddressesByUser(userAddress string) ([]commonTypes.SafeAddress, error) {
	session := r.db.Session()
	iter := session.Query(queries.GetSafeAddressesByUserQuery, userAddress).Iter()

	var safeAddresses []commonTypes.SafeAddress
	var safeAddress commonTypes.SafeAddress
	for iter.Scan(&safeAddress.SafeAddress, &safeAddress.CreatedAt) {
		safeAddresses = append(safeAddresses, safeAddress)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return safeAddresses, nil
}

func (r *safeAddressRepository) CheckSafeAddressExists(userAddress string, safeAddress string) (bool, error) {
	var foundUserAddress string
	var foundSafeAddress string
	err := r.db.Session().Query(queries.CheckSafeAddressExistsQuery, userAddress, safeAddress).Scan(&foundUserAddress, &foundSafeAddress)
	if err != nil {
		return false, nil // Address doesn't exist
	}
	return true, nil
}
