package repository

import (
	"errors"
	"sort"
	"time"

	"github.com/gocql/gocql"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository/queries"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type UserRepository interface {
	CheckUserExists(address string) (int64, error)
	CreateNewUser(user *types.CreateUserDataRequest) (types.UserData, error)
	UpdateUserBalance(user *types.UpdateUserBalanceRequest) error
	UpdateUserJobIDs(userID int64, jobIDs []int64) error
	UpdateUserTasksAndPoints(userID int64, tasksCompleted int64, userPoints float64) error
	GetUserDataByAddress(address string) (int64, types.UserData, error)
	GetUserPointsByID(id int64) (float64, error)
	GetUserPointsByAddress(address string) (float64, error)
	GetUserJobIDsByAddress(address string) (int64, []int64, error)
	GetUserLeaderboard() ([]types.UserLeaderboardEntry, error)
	GetUserLeaderboardByAddress(address string) (types.UserLeaderboardEntry, error)
}

type userRepository struct {
	db *database.Connection
}

func NewUserRepository(db *database.Connection) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CheckUserExists(address string) (int64, error) {
	var userID int64
	err := r.db.Session().Query(queries.GetUserIDByAddressQuery, address).Scan(&userID)
	if err == gocql.ErrNotFound {
		return -1, errors.New("user not found")
	}
	if err != nil {
		return -1, err
	}
	return userID, nil
}

func (r *userRepository) CreateNewUser(user *types.CreateUserDataRequest) (types.UserData, error) {
	var maxUserID int64
	err := r.db.Session().Query(queries.GetMaxUserIDQuery).Scan(&maxUserID)
	if err != nil {
		return types.UserData{}, err
	}
	err = r.db.Session().Query(queries.CreateUserDataQuery, maxUserID + 1, user.UserAddress, user.EtherBalance, user.TokenBalance, user.UserPoints, 0, 0, time.Now()).Exec()
	if err != nil {
		return types.UserData{}, err
	}
	return types.UserData{
		UserID: maxUserID + 1,
		UserAddress: user.UserAddress,
		EtherBalance: user.EtherBalance,
		TokenBalance: user.TokenBalance,
		UserPoints: user.UserPoints,
	}, nil
}

func (r *userRepository) UpdateUserBalance(user *types.UpdateUserBalanceRequest) error {
	err := r.db.Session().Query(queries.UpdateUserBalanceQuery, user.EtherBalance, user.TokenBalance, user.UserID).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserJobIDs(userID int64, jobIDs []int64) error {
	err := r.db.Session().Query(queries.UpdateUserJobIDsQuery, jobIDs, len(jobIDs), time.Now(), userID).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserTasksAndPoints(userID int64, tasksCompleted int64, userPoints float64) error {
	err := r.db.Session().Query(queries.UpdateUserTasksAndPointsQuery, tasksCompleted, userPoints, userID).Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserDataByAddress(address string) (int64, types.UserData, error) {
	var userID int64
	err := r.db.Session().Query(queries.GetUserIDByAddressQuery, address).Scan(&userID)
	if err == gocql.ErrNotFound {
		return -1, types.UserData{}, gocql.ErrNotFound
	}
	if err != nil {
		return -1, types.UserData{}, err
	}
	var userData types.UserData
	err = r.db.Session().Query(queries.GetUserDataByIDQuery, userID).Scan(
			&userData.UserID, &userData.UserAddress, &userData.JobIDs, &userData.TotalJobs, &userData.TotalTasks,
			&userData.EtherBalance, &userData.TokenBalance, &userData.UserPoints, 
			&userData.CreatedAt, &userData.LastUpdatedAt)
	if err != nil {
		return -1, types.UserData{}, err
	}
	return userID, userData, nil
}

func (r *userRepository) GetUserPointsByID(id int64) (float64, error) {
	var userPoints float64
	err := r.db.Session().Query(queries.GetUserPointsByIDQuery, id).Scan(&userPoints)
	if err == gocql.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return userPoints, nil
}

func (r *userRepository) GetUserPointsByAddress(address string) (float64, error) {
	var userPoints float64
	err := r.db.Session().Query(queries.GetUserPointsByAddressQuery, address).Scan(&userPoints)
	if err == gocql.ErrNotFound {
		return 0, errors.New("user address not found")
	}
	if err != nil {
		return 0, err
	}
	return userPoints, nil
}

func (r *userRepository) GetUserJobIDsByAddress(address string) (int64, []int64, error) {
	var userID int64
	var jobIDs []int64
	err := r.db.Session().Query(queries.GetUserJobIDsByAddressQuery, address).Scan(&userID, &jobIDs)
	if err == gocql.ErrNotFound {
		return -1, nil, errors.New("user address not found")
	}
	if err != nil {
		return -1, nil, err
	}
	return userID, jobIDs, nil
}

func (r *userRepository) GetUserLeaderboard() ([]types.UserLeaderboardEntry, error) {
	iter := r.db.Session().Query(queries.GetUserLeaderboardQuery).Iter()

	var leaderboard []types.UserLeaderboardEntry
	var userEntry types.UserLeaderboardEntry

	for iter.Scan(
		&userEntry.UserID,
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

	// Sort leaderboard by UserPoints (desc), TotalJobs (desc), TotalTasks (desc), UserID (asc)
	sort.Slice(leaderboard, func(i, j int) bool {
		// First compare UserPoints
		if leaderboard[i].UserPoints != leaderboard[j].UserPoints {
			return leaderboard[i].UserPoints > leaderboard[j].UserPoints
		}
		// If UserPoints equal, compare TotalJobs
		if leaderboard[i].TotalJobs != leaderboard[j].TotalJobs {
			return leaderboard[i].TotalJobs > leaderboard[j].TotalJobs
		}
		// If TotalJobs equal, compare TotalTasks
		if leaderboard[i].TotalTasks != leaderboard[j].TotalTasks {
			return leaderboard[i].TotalTasks > leaderboard[j].TotalTasks
		}
		// If all else equal, sort by UserID ascending
		return leaderboard[i].UserID < leaderboard[j].UserID
	})
	return leaderboard, nil
}

func (r *userRepository) GetUserLeaderboardByAddress(address string) (types.UserLeaderboardEntry, error) {
	var userEntry types.UserLeaderboardEntry
	err := r.db.Session().Query(queries.GetUserLeaderboardByAddressQuery, address).Scan(&userEntry.UserID, &userEntry.UserAddress, &userEntry.TotalJobs, &userEntry.TotalTasks, &userEntry.UserPoints)
	if err != nil {
		return types.UserLeaderboardEntry{}, err
	}

	return userEntry, nil
}
