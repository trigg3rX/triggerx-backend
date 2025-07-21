package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/docker/config"
	"github.com/trigg3rX/triggerx-backend/pkg/docker/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type CachedFile struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	LastAccessed time.Time `json:"last_accessed"`
	Size         int64     `json:"size"`
	IsValid      bool      `json:"is_valid"`
	CreatedAt    time.Time `json:"created_at"`
}

type FileCache struct {
	cacheDir      string
	fileCache     map[string]*CachedFile
	mutex         sync.RWMutex
	config        config.CacheConfig
	logger        logging.Logger
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	stats         *types.CacheStats
	statsMutex    sync.RWMutex
}

func NewFileCache(cfg config.CacheConfig, logger logging.Logger) (*FileCache, error) {
	cacheDir := filepath.Join("/tmp", "docker-file-cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &FileCache{
		cacheDir:    cacheDir,
		fileCache:   make(map[string]*CachedFile),
		config:      cfg,
		logger:      logger,
		stopCleanup: make(chan struct{}),
		stats: &types.CacheStats{
			HitCount:      0,
			MissCount:     0,
			HitRate:       0.0,
			Size:          0,
			MaxSize:       cfg.MaxCacheSize,
			ItemCount:     0,
			EvictionCount: 0,
			LastCleanup:   time.Now(),
		},
	}

	cache.startCleanupRoutine()

	if err := cache.loadExistingFiles(); err != nil {
		logger.Warnf("Failed to load existing cached files: %v", err)
	}

	return cache, nil
}

// GetOrDownload checks if the file is in the cache by key (CID or URL). If not, it calls downloadFunc to get the content and stores it.
func (c *FileCache) GetOrDownload(key string, downloadFunc func() ([]byte, error)) (string, error) {
	c.mutex.RLock()
	if cachedFile, exists := c.fileCache[key]; exists {
		c.mutex.RUnlock()
		return c.accessCachedFile(cachedFile)
	}
	c.mutex.RUnlock()

	// Cache miss - need to download and store the file
	c.statsMutex.Lock()
	c.stats.MissCount++
	c.updateHitRate()
	c.statsMutex.Unlock()

	content, err := downloadFunc()
	if err != nil {
		return "", err
	}
	return c.storeFile(key, content)
}

// GetByKey retrieves a cached file by its key (CID or URL)
func (c *FileCache) GetByKey(key string) (string, error) {
	c.mutex.RLock()
	cachedFile, exists := c.fileCache[key]
	c.mutex.RUnlock()

	if !exists {
		c.statsMutex.Lock()
		c.stats.MissCount++
		c.updateHitRate()
		c.statsMutex.Unlock()
		return "", fmt.Errorf("file not found in cache: %s", key)
	}

	c.statsMutex.Lock()
	c.stats.HitCount++
	c.updateHitRate()
	c.statsMutex.Unlock()

	return c.accessCachedFile(cachedFile)
}

func (c *FileCache) accessCachedFile(cachedFile *CachedFile) (string, error) {
	// Check if file still exists on disk
	if _, err := os.Stat(cachedFile.Path); os.IsNotExist(err) {
		// File was deleted, remove from cache
		c.mutex.Lock()
		delete(c.fileCache, cachedFile.Hash)
		c.mutex.Unlock()
		return "", fmt.Errorf("cached file not found on disk: %s", cachedFile.Path)
	}

	cachedFile.LastAccessed = time.Now()

	// Check if file is still valid (within TTL)
	if time.Since(cachedFile.LastAccessed) > c.config.CacheTTL {
		cachedFile.IsValid = false
	}

	if !cachedFile.IsValid {
		return "", fmt.Errorf("cached file is expired: %s", cachedFile.Path)
	}

	return cachedFile.Path, nil
}

// storeFile now uses the key (CID or URL) as the filename (with .go extension)
func (c *FileCache) storeFile(key string, content []byte) (string, error) {
	if err := c.ensureSpace(int64(len(content))); err != nil {
		return "", fmt.Errorf("failed to ensure cache space: %w", err)
	}

	// Sanitize key for filename (replace / and :)
	filename := strings.ReplaceAll(strings.ReplaceAll(key, "/", "_"), ":", "_") + ".go"
	filePath := filepath.Join(c.cacheDir, filename)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write cached file: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	cachedFile := &CachedFile{
		Path:         filePath,
		Hash:         key,
		LastAccessed: time.Now(),
		Size:         fileInfo.Size(),
		IsValid:      true,
		CreatedAt:    time.Now(),
	}

	c.mutex.Lock()
	c.fileCache[key] = cachedFile
	c.stats.ItemCount++
	c.stats.Size += fileInfo.Size()
	c.mutex.Unlock()

	c.logger.Infof("Stored file in cache (size: %d bytes)", fileInfo.Size())
	return filePath, nil
}

func (c *FileCache) ensureSpace(requiredSize int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If we have enough space, no need to evict
	if c.stats.Size+requiredSize <= c.config.MaxCacheSize {
		return nil
	}

	// Need to evict files
	evictionSize := c.stats.Size + requiredSize - c.config.MaxCacheSize

	// Sort files by last accessed time (oldest first)
	type fileEntry struct {
		hash string
		file *CachedFile
	}

	var files []fileEntry
	for hash, file := range c.fileCache {
		files = append(files, fileEntry{hash: hash, file: file})
	}

	// Sort by last accessed time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].file.LastAccessed.After(files[j].file.LastAccessed) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Evict files until we have enough space
	evictedSize := int64(0)
	for _, entry := range files {
		if evictedSize >= evictionSize {
			break
		}

		// Remove file from disk
		if err := os.Remove(entry.file.Path); err != nil {
			c.logger.Warnf("Failed to remove cached file: %v", err)
			continue
		}

		// Remove from cache
		delete(c.fileCache, entry.hash)
		evictedSize += entry.file.Size
		c.stats.EvictionCount++
		c.stats.ItemCount--
		c.stats.Size -= entry.file.Size

		c.logger.Debugf("Evicted cached file: %s (size: %d bytes)", entry.hash, entry.file.Size)
	}

	return nil
}

func (c *FileCache) startCleanupRoutine() {
	c.cleanupTicker = time.NewTicker(c.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-c.cleanupTicker.C:
				c.cleanup()
			case <-c.stopCleanup:
				c.cleanupTicker.Stop()
				return
			}
		}
	}()
}

func (c *FileCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredFiles := make([]string, 0)

	// Find expired files
	for hash, file := range c.fileCache {
		if now.Sub(file.LastAccessed) > c.config.CacheTTL {
			expiredFiles = append(expiredFiles, hash)
		}
	}

	// Remove expired files
	for _, hash := range expiredFiles {
		file := c.fileCache[hash]
		if err := os.Remove(file.Path); err != nil {
			c.logger.Warnf("Failed to remove expired file: %v", err)
			continue
		}

		delete(c.fileCache, hash)
		c.stats.ItemCount--
		c.stats.Size -= file.Size
		c.stats.EvictionCount++
	}

	if len(expiredFiles) > 0 {
		c.logger.Infof("Cleaned up %d expired files from cache", len(expiredFiles))
	}

	c.stats.LastCleanup = now
}

func (c *FileCache) loadExistingFiles() error {
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}

		// Extract hash from filename
		key := strings.TrimSuffix(entry.Name(), ".go")
		filePath := filepath.Join(c.cacheDir, entry.Name())

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			c.logger.Warnf("Failed to stat cached file %s: %v", filePath, err)
			continue
		}

		cachedFile := &CachedFile{
			Path:         filePath,
			Hash:         key,
			LastAccessed: fileInfo.ModTime(),
			Size:         fileInfo.Size(),
			IsValid:      true,
			CreatedAt:    fileInfo.ModTime(),
		}

		c.fileCache[key] = cachedFile
		c.stats.ItemCount++
		c.stats.Size += fileInfo.Size()
	}

	c.logger.Infof("Loaded %d existing cached files", len(c.fileCache))
	return nil
}

func (c *FileCache) updateHitRate() {
	total := c.stats.HitCount + c.stats.MissCount
	if total > 0 {
		c.stats.HitRate = float64(c.stats.HitCount) / float64(total)
	}
}

func (c *FileCache) GetStats() *types.CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *c.stats
	return &stats
}

func (c *FileCache) Close() error {
	if c.cleanupTicker != nil {
		close(c.stopCleanup)
	}
	return nil
}
