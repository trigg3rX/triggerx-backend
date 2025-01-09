package manager

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// CacheData represents the scheduler state that needs to be persisted
type CacheData struct {
    Jobs            map[string]*Job     `json:"jobs"`
    Quorums         map[string]*Quorum  `json:"quorums"`
    WaitingQueue    []WaitingJob        `json:"waiting_queue"`
    LastUpdated     time.Time           `json:"last_updated"`
    SystemResources SystemResources     `json:"system_resources"`
}

// CacheManager handles persistence of scheduler state
type CacheManager struct {
    cacheDir     string
    cacheFile    string
    cacheMutex   sync.RWMutex
    scheduler    *JobScheduler
    saveInterval time.Duration
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(scheduler *JobScheduler, cacheInterval time.Duration) (*CacheManager, error) {
    cacheDir := filepath.Join("data", "jobcache")
    if err := os.MkdirAll(cacheDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create cache directory: %v", err)
    }

    return &CacheManager{
        cacheDir:     cacheDir,
        cacheFile:    filepath.Join(cacheDir, "cache.json"),
        scheduler:    scheduler,
        saveInterval: cacheInterval,
    }, nil
}

// Start begins periodic caching of scheduler state
func (cm *CacheManager) Start() {
    go func() {
        ticker := time.NewTicker(cm.saveInterval)
        defer ticker.Stop()

        for {
            select {
            case <-cm.scheduler.ctx.Done():
                // Save one final time before shutting down
                if err := cm.SaveState(); err != nil {
                    log.Printf("Failed to save final cache state: %v", err)
                }
                return
            case <-ticker.C:
                if err := cm.SaveState(); err != nil {
                    log.Printf("Failed to save cache state: %v", err)
                }
            }
        }
    }()
}

// SaveState persists the current scheduler state to disk
func (cm *CacheManager) SaveState() error {
    cm.cacheMutex.Lock()
    defer cm.cacheMutex.Unlock()

    cm.scheduler.mu.RLock()
    cm.scheduler.waitingQueueMu.RLock()
    
    cacheData := CacheData{
        Jobs:            make(map[string]*Job),
        Quorums:         make(map[string]*Quorum),
        WaitingQueue:    make([]WaitingJob, len(cm.scheduler.waitingQueue)),
        LastUpdated:     time.Now(),
        SystemResources: cm.scheduler.resources,
    }

    // Deep copy jobs and quorums to avoid race conditions
    for id, job := range cm.scheduler.jobs {
        jobCopy := *job
        cacheData.Jobs[id] = &jobCopy
    }

    for id, quorum := range cm.scheduler.quorums {
        quorumCopy := *quorum
        cacheData.Quorums[id] = &quorumCopy
    }

    copy(cacheData.WaitingQueue, cm.scheduler.waitingQueue)

    cm.scheduler.waitingQueueMu.RUnlock()
    cm.scheduler.mu.RUnlock()

    // Create temporary file for atomic write
    tempFile := cm.cacheFile + ".tmp"
    file, err := os.Create(tempFile)
    if err != nil {
        return fmt.Errorf("failed to create temp cache file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "    ")
    if err := encoder.Encode(cacheData); err != nil {
        return fmt.Errorf("failed to encode cache data: %v", err)
    }

    // Atomic rename
    if err := os.Rename(tempFile, cm.cacheFile); err != nil {
        return fmt.Errorf("failed to save cache file: %v", err)
    }

    log.Printf("Successfully saved scheduler state to cache")
    return nil
}

// LoadState restores scheduler state from cache
func (cm *CacheManager) LoadState() error {
    cm.cacheMutex.RLock()
    defer cm.cacheMutex.RUnlock()

    file, err := os.Open(cm.cacheFile)
    if err != nil {
        if os.IsNotExist(err) {
            log.Printf("No cache file found, starting with fresh state")
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
        log.Printf("Cache data is too old, starting with fresh state")
        return nil
    }

    cm.scheduler.mu.Lock()
    cm.scheduler.waitingQueueMu.Lock()
    defer cm.scheduler.mu.Unlock()
    defer cm.scheduler.waitingQueueMu.Unlock()

    // Restore scheduler state
    cm.scheduler.jobs = cacheData.Jobs
    cm.scheduler.quorums = cacheData.Quorums
    cm.scheduler.waitingQueue = cacheData.WaitingQueue
    cm.scheduler.resources = cacheData.SystemResources

    // Reschedule active jobs
    for _, job := range cacheData.Jobs {
        if job.Status != "completed" && job.Status != "failed" {
            if err := cm.scheduler.scheduleJob(job); err != nil {
                log.Printf("Failed to reschedule job %s: %v", job.JobID, err)
            }
        }
    }

    log.Printf("Successfully restored scheduler state from cache")
    return nil
}