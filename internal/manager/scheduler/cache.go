package scheduler

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// // CacheData represents the complete state of the job scheduler that needs to be persisted
// // including active jobs, watchers, queues and system resources
// type CacheData struct {
// 	ActiveJobs      map[string]interface{} `json:"active_jobs"`
// 	EventWatchers   []int64                `json:"event_watchers"`
// 	ConditionJobs   []int64                `json:"condition_jobs"`
// 	JobQueue        JobQueue               `json:"job_queue"`
// 	SystemResources SystemResources        `json:"system_resources"`
// 	LastUpdated     time.Time              `json:"last_updated"`
// 	JobChains       map[int64]int64        `json:"job_chains"`
// }

// // CacheManager handles persisting and restoring scheduler state to disk
// // Uses file-based JSON storage with mutex-protected access
// type CacheManager struct {
// 	cacheDir   string
// 	cacheFile  string
// 	cacheMutex sync.RWMutex
// 	scheduler  *JobScheduler
// }

// func NewCacheManager(scheduler *JobScheduler) (*CacheManager, error) {
// 	cacheDir := filepath.Join("data", "cache")
// 	if err := os.MkdirAll(cacheDir, 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create cache directory: %v", err)
// 	}

// 	return &CacheManager{
// 		cacheDir:  cacheDir,
// 		cacheFile: filepath.Join(cacheDir, "scheduler_cache.json"),
// 		scheduler: scheduler,
// 	}, nil
// }

// // SaveState persists the current scheduler state to disk
// // Merges existing cache with current state to avoid data loss
// func (cm *CacheManager) SaveState() error {
// 	cm.cacheMutex.Lock()
// 	defer cm.cacheMutex.Unlock()

// 	cm.scheduler.mu.RLock()
// 	defer cm.scheduler.mu.RUnlock()

// 	var cacheData CacheData
// 	if file, err := os.Open(cm.cacheFile); err == nil {
// 		if err := json.NewDecoder(file).Decode(&cacheData); err == nil {
// 			file.Close()
// 		}
// 	}

// 	if cacheData.ActiveJobs == nil {
// 		cacheData.ActiveJobs = make(map[string]interface{})
// 	}

// 	for jobID, jobData := range cm.scheduler.stateCache {
// 		cacheData.ActiveJobs[fmt.Sprintf("%d", jobID)] = jobData
// 	}

// 	cacheData.EventWatchers = make([]int64, 0, len(cm.scheduler.eventWatchers))
// 	cacheData.ConditionJobs = make([]int64, 0, len(cm.scheduler.conditions))
// 	cacheData.JobQueue = cm.scheduler.balancer.jobQueue
// 	cacheData.SystemResources = cm.scheduler.balancer.resources
// 	cacheData.LastUpdated = time.Now()

// 	// Add job chains to cache
// 	cacheData.JobChains = make(map[int64]int64)
// 	for jobID := range cm.scheduler.workers {
// 		// Get job data using database call
// 		success, respData := cm.scheduler.SendDataToDatabase(fmt.Sprintf("job_data/%d", jobID), 2, nil)
// 		if !success || respData == nil {
// 			continue
// 		}

// 		var jobData types.HandleCreateJobData
// 		if err := json.Unmarshal([]byte(respData.(string)), &jobData); err != nil {
// 			cm.scheduler.logger.Errorf("Failed to unmarshal job data: %v", err)
// 			continue
// 		}

// 		if jobData.LinkJobID > 0 {
// 			cacheData.JobChains[jobID] = jobData.LinkJobID
// 		}
// 	}

// 	file, err := os.Create(cm.cacheFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to create cache file: %v", err)
// 	}
// 	defer file.Close()

// 	encoder := json.NewEncoder(file)
// 	encoder.SetIndent("", "    ")
// 	return encoder.Encode(cacheData)
// }

// // LoadState restores scheduler state from disk cache
// // Skips restoration if cache is too old (>1h) and initializes jobs from cached state
// func (cm *CacheManager) LoadState() error {
// 	cm.cacheMutex.RLock()
// 	defer cm.cacheMutex.RUnlock()

// 	file, err := os.Open(cm.cacheFile)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			cm.scheduler.logger.Info("No cache file found, starting with fresh state")
// 			return nil
// 		}
// 		return fmt.Errorf("failed to open cache file: %v", err)
// 	}
// 	defer file.Close()

// 	var cacheData CacheData
// 	if err := json.NewDecoder(file).Decode(&cacheData); err != nil {
// 		return fmt.Errorf("failed to decode cache data: %v", err)
// 	}

// 	if time.Since(cacheData.LastUpdated) > 120*time.Hour {
// 		cm.scheduler.logger.Info("Cache data is too old, starting with fresh state")
// 		return nil
// 	}

// 	cm.scheduler.mu.Lock()
// 	defer cm.scheduler.mu.Unlock()

// 	convertedCache := make(map[int64]interface{})
// 	for strID, data := range cacheData.ActiveJobs {
// 		id, _ := strconv.ParseInt(strID, 10, 64)
// 		convertedCache[id] = data
// 	}
// 	cm.scheduler.stateCache = convertedCache
// 	cm.scheduler.balancer.jobQueue = cacheData.JobQueue
// 	cm.scheduler.balancer.resources = cacheData.SystemResources

// 	for _, jobID := range cacheData.EventWatchers {
// 		// Fetch job data from database first
// 		success, respData := cm.scheduler.SendDataToDatabase(fmt.Sprintf("job_data/%d", jobID), 2, nil)
// 		if !success || respData == nil {
// 			continue
// 		}

// 		var jobData types.HandleCreateJobData
// 		if err := json.Unmarshal([]byte(respData.(string)), &jobData); err != nil {
// 			cm.scheduler.logger.Errorf("Failed to unmarshal job data: %v", err)
// 			continue
// 		}

// 		if err := cm.scheduler.StartEventBasedJob(jobData); err != nil {
// 			cm.scheduler.logger.Errorf("Failed to restore event job %d: %v", jobID, err)
// 		}
// 	}

// 	for _, jobID := range cacheData.ConditionJobs {
// 		// Fetch job data from database first
// 		success, respData := cm.scheduler.SendDataToDatabase(fmt.Sprintf("job_data/%d", jobID), 2, nil)
// 		if !success || respData == nil {
// 			continue
// 		}

// 		var jobData types.HandleCreateJobData
// 		if err := json.Unmarshal([]byte(respData.(string)), &jobData); err != nil {
// 			cm.scheduler.logger.Errorf("Failed to unmarshal job data: %v", err)
// 			continue
// 		}

// 		if err := cm.scheduler.StartConditionBasedJob(jobData); err != nil {
// 			cm.scheduler.logger.Errorf("Failed to restore condition job %d: %v", jobID, err)
// 		}
// 	}

// 	cm.scheduler.logger.Info("Successfully restored scheduler state from cache")
// 	return nil
// }

// // RemoveJob deletes a job from both memory and persistent cache
// // Updates cache file to reflect the removal while maintaining other jobs' state
// func (s *JobScheduler) RemoveJob(jobID int64) {
// 	s.mu.Lock()
// 	delete(s.workers, jobID)
// 	s.mu.Unlock()

// 	s.cacheMutex.Lock()
// 	delete(s.stateCache, jobID)
// 	s.cacheMutex.Unlock()

// 	s.cacheManager.cacheMutex.Lock()
// 	defer s.cacheManager.cacheMutex.Unlock()

// 	var cacheData CacheData
// 	if file, err := os.Open(s.cacheManager.cacheFile); err == nil {
// 		if err := json.NewDecoder(file).Decode(&cacheData); err == nil {
// 			file.Close()
// 			delete(cacheData.ActiveJobs, fmt.Sprintf("%d", jobID))

// 			file, err = os.Create(s.cacheManager.cacheFile)
// 			if err == nil {
// 				defer file.Close()
// 				encoder := json.NewEncoder(file)
// 				encoder.SetIndent("", "    ")
// 				if err := encoder.Encode(cacheData); err != nil {
// 					s.logger.Errorf("Failed to encode cache data: %v", err)
// 				}
// 			}
// 		}
// 	}
// }
