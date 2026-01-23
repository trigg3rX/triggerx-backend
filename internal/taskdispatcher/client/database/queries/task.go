package queries

const (
	UpdateTaskStatusQuery = `
		UPDATE triggerx.task_data
		SET task_status = ?,
			task_error = ?
		WHERE task_id = ?`
)