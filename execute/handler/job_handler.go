package handler

import (
    "log"
    "fmt"
    
    "github.com/trigg3rX/triggerx-keeper/execute/executor"
    "github.com/trigg3rX/go-backend/execute/manager" 
)

type JobHandler struct {
    executor *executor.JobExecutor
}

func NewJobHandler() *JobHandler {
    return &JobHandler{
        executor: executor.NewJobExecutor(),
    }
}

func (h *JobHandler) HandleJob(job *manager.Job) error {
    log.Printf("🔧 Received job %s for execution", job.JobID)
    
    // Validate job
    if err := h.validateJob(job); err != nil {
        log.Printf("❌ Job validation failed: %v", err)
        return err
    }

    // Execute job
    result, err := h.executor.Execute(job)
    if err != nil {
        log.Printf("❌ Job execution failed: %v", err)
        return err
    }

    log.Printf("✅ Job %s executed successfully. Result: %v", job.JobID, result)
    return nil
}

func (h *JobHandler) validateJob(job *manager.Job) error {
    if job == nil {
        return fmt.Errorf("received nil job")
    }
    if job.JobID == "" {
        return fmt.Errorf("invalid job: empty job ID")
    }
    return nil
}