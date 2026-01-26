package repository

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"

	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type UserRepository interface {
	CreateNewUser(user *types.CreateUserDataRequest) error
	UpdateUserJobIDs(userAddress string, jobIDs []string) error
	UpdateUserPoints(userAddress string, userPoints string) error
	UpdateUserEmail(userAddress string, email string) error
	GetUserData(userAddress string) (*types.UserDataDTO, error)
	GetUserPoints(userAddress string) (string, error)
	GetUserJobIDs(userAddress string) ([]string, error)
	GetUserLeaderboard() ([]types.UserLeaderboardEntry, error)
}

type userRepository struct {
	db *database.Connection
}

func NewUserRepository(db *database.Connection) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateNewUser(user *types.CreateUserDataRequest) error {
	err := r.db.Session().Query(CreateUserDataQuery,
		strings.ToLower(user.UserAddress), user.EmailID, []string{}, user.UserPoints,
		0, 0, time.Now().UTC(), time.Now().UTC()).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserJobIDs(userAddress string, jobIDs []string) error {
	err := r.db.Session().Query(UpdateUserJobIDsQuery,
		jobIDs, len(jobIDs), time.Now().UTC(), strings.ToLower(userAddress)).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserEmail(userAddress string, email string) error {
	err := r.db.Session().Query(UpdateUserEmailQuery,
		strings.ToLower(email), time.Now().UTC(), strings.ToLower(userAddress)).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserPoints(userAddress string, userPoints string) error {
	err := r.db.Session().Query(UpdateUserPointsQuery,
		userPoints, time.Now().UTC(), strings.ToLower(userAddress)).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserData(userAddress string) (*types.UserDataDTO, error) {
	var entity types.UserDataEntity

	err := r.db.Session().Query(GetUserDataQuery, strings.ToLower(userAddress)).Scan(
		&entity.UserAddress, &entity.EmailID, &entity.JobIDs, &entity.UserPoints,
		&entity.TotalJobs, &entity.TotalTasks, &entity.CreatedAt, &entity.LastUpdatedAt)

	if err == gocql.ErrNotFound {
		return nil, gocql.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	dto := types.UserDataEntityToDTO(&entity)
	return dto, nil
}

func (r *userRepository) GetUserPoints(userAddress string) (string, error) {
	var userPoints string
	err := r.db.Session().Query(GetUserPointsQuery, strings.ToLower(userAddress)).Scan(&userPoints)
	if err == gocql.ErrNotFound {
		return "0", nil
	}
	if err != nil {
		return "0", err
	}
	return userPoints, nil
}

func (r *userRepository) GetUserJobIDs(userAddress string) ([]string, error) {
	var jobIDs []string
	err := r.db.Session().Query(GetUserJobIDsQuery, strings.ToLower(userAddress)).Scan(&jobIDs)
	if err == gocql.ErrNotFound {
		return nil, errors.New("user address not found")
	}
	if err != nil {
		return nil, err
	}

	return jobIDs, nil
}

func (r *userRepository) GetUserLeaderboard() ([]types.UserLeaderboardEntry, error) {
	iter := r.db.Session().Query(GetUserLeaderboardQuery).Iter()

	var leaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	for iter.Scan(
		&userEntry.UserAddress,
		&userEntry.TotalJobs,
		&userEntry.TotalTasks,
		&userEntry.UserPoints,
	) {
		leaderboard = append(leaderboard, userEntry)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	// Sort leaderboard by UserPoints (desc), TotalJobs (desc), TotalTasks (desc), UserAddress (asc)
	sort.Slice(leaderboard, func(i, j int) bool {
		// First compare UserPoints using big integer comparison (descending)
		if !types.IsEqual(leaderboard[i].UserPoints, leaderboard[j].UserPoints) {
			return types.IsGreater(leaderboard[i].UserPoints, leaderboard[j].UserPoints)
		}
		// If UserPoints equal, compare TotalJobs
		if leaderboard[i].TotalJobs != leaderboard[j].TotalJobs {
			return leaderboard[i].TotalJobs > leaderboard[j].TotalJobs
		}
		// If TotalJobs equal, compare TotalTasks
		if leaderboard[i].TotalTasks != leaderboard[j].TotalTasks {
			return leaderboard[i].TotalTasks > leaderboard[j].TotalTasks
		}
		// If all else equal, sort by UserAddress ascending
		return leaderboard[i].UserAddress < leaderboard[j].UserAddress
	})
	return leaderboard, nil
}
