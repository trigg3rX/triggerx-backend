package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// CacheData represents the current state to be persisted
type CacheData struct {
	ActiveJobs      map[string]interface{} `json:"active_jobs"`
	EventWatchers   []int64                `json:"event_watchers"`
	ConditionJobs   []int64                `json:"condition_jobs"`
	JobQueue        JobQueue               `json:"job_queue"`
	SystemResources SystemResources        `json:"system_resources"`
	LastUpdated     time.Time              `json:"last_updated"`
}

type CacheManager struct {
	cacheDir   string
	cacheFile  string
	cacheMutex sync.RWMutex
	scheduler  *JobScheduler
}

func NewCacheManager(scheduler *JobScheduler) (*CacheManager, error) {
	cacheDir := filepath.Join("data", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	return &CacheManager{
		cacheDir:  cacheDir,
		cacheFile: filepath.Join(cacheDir, "scheduler_cache.json"),
		scheduler: scheduler,
	}, nil
}

func (cm *CacheManager) SaveState() error {
	cm.cacheMutex.Lock()
	defer cm.cacheMutex.Unlock()

	cm.scheduler.mu.RLock()
	defer cm.scheduler.mu.RUnlock()

	// First load existing cache data
	var cacheData CacheData
	if file, err := os.Open(cm.cacheFile); err == nil {
		if err := json.NewDecoder(file).Decode(&cacheData); err == nil {
			file.Close()
		}
	}

	// If active_jobs is nil, initialize it
	if cacheData.ActiveJobs == nil {
		cacheData.ActiveJobs = make(map[string]interface{})
	}

	// Merge new jobs with existing ones
	for jobID, jobData := range cm.scheduler.stateCache {
		cacheData.ActiveJobs[fmt.Sprintf("%d", jobID)] = jobData
	}

	// Update the rest of the cache data
	cacheData.EventWatchers = make([]int64, 0, len(cm.scheduler.eventWatchers))
	cacheData.ConditionJobs = make([]int64, 0, len(cm.scheduler.conditions))
	cacheData.JobQueue = cm.scheduler.balancer.jobQueue
	cacheData.SystemResources = cm.scheduler.balancer.resources
	cacheData.LastUpdated = time.Now()

	// Save to file
	file, err := os.Create(cm.cacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(cacheData)
}

func (cm *CacheManager) LoadState() error {
	cm.cacheMutex.RLock()
	defer cm.cacheMutex.RUnlock()

	file, err := os.Open(cm.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			cm.scheduler.logger.Info("No cache file found, starting with fresh state")
			return nil
		}
		return fmt.Errorf("failed to open cache file: %v", err)
	}
	defer file.Close()

	var cacheData CacheData
	if err := json.NewDecoder(file).Decode(&cacheData); err != nil {
		return fmt.Errorf("failed to decode cache data: %v", err)
	}

	// Verify cache isn't too old
	if time.Since(cacheData.LastUpdated) > 1*time.Hour {
		cm.scheduler.logger.Info("Cache data is too old, starting with fresh state")
		return nil
	}

	cm.scheduler.mu.Lock()
	defer cm.scheduler.mu.Unlock()

	// Restore state
	convertedCache := make(map[int64]interface{})
	for strID, data := range cacheData.ActiveJobs {
		id, _ := strconv.ParseInt(strID, 10, 64)
		convertedCache[id] = data
	}
	cm.scheduler.stateCache = convertedCache
	cm.scheduler.balancer.jobQueue = cacheData.JobQueue
	cm.scheduler.balancer.resources = cacheData.SystemResources

	// Restore event watchers
	for _, jobID := range cacheData.EventWatchers {
		if err := cm.scheduler.StartEventBasedJob(jobID); err != nil {
			cm.scheduler.logger.Errorf("Failed to restore event job %d: %v", jobID, err)
		}
	}

	// Restore condition jobs
	for _, jobID := range cacheData.ConditionJobs {
		if err := cm.scheduler.StartConditionBasedJob(jobID); err != nil {
			cm.scheduler.logger.Errorf("Failed to restore condition job %d: %v", jobID, err)
		}
	}

	cm.scheduler.logger.Info("Successfully restored scheduler state from cache")
	return nil
}

func (s *JobScheduler) RemoveJob(jobID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.workers, jobID)

	s.cacheMutex.Lock()
	delete(s.stateCache, jobID)
	s.cacheMutex.Unlock()

	if err := s.cacheManager.SaveState(); err != nil {
		s.logger.Errorf("Failed to save state after removing job %d: %v", jobID, err)
	}
}