package redis

import "github.com/trigg3rX/triggerx-backend/pkg/types"	

// TODO:
// - Add a variable and function to maintain the current performer ID, for round robin selection of performers
// - Add a function which maintains the performer lock in redis
// - Add a function to fetch the available performers from health service
// - Add a function to to update the performer ID in DB when it is locked for a task

func GetPerformerData() types.PerformerData {
	performerData := types.PerformerData{}
	performerData.KeeperID = 1
	performerData.KeeperAddress = "0x1234567890"
	return performerData
}