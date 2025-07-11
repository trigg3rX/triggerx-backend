package database

import (
	"github.com/trigg3rX/triggerx-backend-imua/internal/registrar/types"
)

// UpdateTaskSubmissionData updates task number, success status and execution details in database
func (dm *DatabaseClient) UpdateTaskSubmissionData(data types.TaskSubmissionData) error {
	dm.logger.Infof("Updating task %d with number %d and acceptance status %t", data.TaskID, data.TaskNumber, data.IsAccepted)

	performerId := data.KeeperIds[0]
	attesterIds := data.KeeperIds[1:]

	if err := dm.db.Session().Query(`
		UPDATE triggerx.task_data 
		SET task_number = ?, is_accepted = ?, task_submission_tx_hash = ?, 
		    task_performer_id = ?, task_attester_ids = ?, ta_signature = ?, tp_signature = ?
		WHERE task_id = ?`,
		data.TaskNumber, data.IsAccepted, data.TaskSubmissionTxHash, performerId, attesterIds, data.AttesterSignatures, data.PerformerSignature, data.TaskID).Exec(); err != nil {
		dm.logger.Errorf("Error updating task execution details for task ID %d: %v", data.TaskID, err)
		return err
	}

	dm.logger.Infof("Successfully updated task %d with submission details", data.TaskID)
	return nil
}

// UpdateJobStatus updates job status in database
func (dm *DatabaseClient) AddTaskIdToJob(taskID int64, jobID int64) error {
	dm.logger.Infof("Adding task %d to job %d", taskID, jobID)

	// First get the job ID from task ID
	var taskIds []int64
	if err := dm.db.Session().Query(`
		SELECT task_ids FROM triggerx.job_data WHERE job_id = ?`,
		jobID).Scan(&taskIds); err != nil {
		dm.logger.Errorf("Failed to get job ID for task ID %d: %v", jobID, err)
		return err
	}
	taskIds = append(taskIds, taskID)

	// Update job status
	if err := dm.db.Session().Query(`
		UPDATE triggerx.job_data 
		SET task_ids = task_ids ?
		WHERE job_id = ?`,
		taskIds, jobID).Exec(); err != nil {
		dm.logger.Errorf("Error updating job status for job ID %d: %v", jobID, err)
		return err
	}

	dm.logger.Infof("Successfully added task %d to job %d", taskID, jobID)
	return nil
}