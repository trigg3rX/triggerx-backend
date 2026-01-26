package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type SafeAddressRepository interface {
	CreateSafeAddress(safeAddress *types.SafeAddressDataEntity) error
	GetSafeAddressesByUser(userAddress string) ([]types.SafeAddressDataDTO, error)
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

func (r *safeAddressRepository) CreateSafeAddress(safeAddress *types.SafeAddressDataEntity) error {
	err := r.db.Session().Query(CreateSafeAddressQuery,
		strings.ToLower(safeAddress.UserAddress), strings.ToLower(safeAddress.SafeAddress), safeAddress.SafeName, time.Now().UTC()).Exec()
	if err != nil {
		return errors.New("failed to create safe address")
	}
	return nil
}

func (r *safeAddressRepository) GetSafeAddressesByUser(userAddress string) ([]types.SafeAddressDataDTO, error) {
	session := r.db.Session()
	iter := session.Query(GetSafeAddressesByUserQuery, strings.ToLower(userAddress)).Iter()

	var safeAddresses []types.SafeAddressDataDTO
	var safeAddress types.SafeAddressDataDTO
	for iter.Scan(&safeAddress.SafeAddress, &safeAddress.SafeName, &safeAddress.CreatedAt) {
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
	err := r.db.Session().Query(CheckSafeAddressExistsQuery, strings.ToLower(userAddress), strings.ToLower(safeAddress)).Scan(&foundUserAddress, &foundSafeAddress)
	if err != nil {
		return false, nil // Address doesn't exist
	}
	return true, nil
}
