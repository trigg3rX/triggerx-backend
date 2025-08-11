package file

import (
	"encoding/json"
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

type cachedFile struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	LastAccessed time.Time `json:"last_accessed"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"created_at"`
}

type fileCache struct {
	cacheDir      string
	fileCache     map[string]*cachedFile
	mutex         sync.RWMutex
	config        config.CacheConfig
	logger        logging.Logger
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
	stats         *types.CacheStats
	statsMutex    sync.RWMutex
}

func newFileCache(cfg config.CacheConfig, logger logging.Logger) (*fileCache, error) {
	// Use configured cache directory or fallback to persistent location
	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		cacheDir = "/var/lib/triggerx/cache"
	}

	// Ensure the cache directory exists with proper permissions
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	cache := &fileCache{
		cacheDir:    cacheDir,
		fileCache:   make(map[string]*cachedFile),
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

	// cache.startCleanupRoutine()

	if err := cache.loadExistingFiles(); err != nil {
		logger.Warnf("Failed to load existing cached files: %v", err)
	}

	return cache, nil
}

// GetOrDownload checks if the file is in the cache by key (CID or URL). If not, it calls downloadFunc to get the content and stores it.
func (c *fileCache) getOrDownloadFile(key string, downloadFunc func() ([]byte, error)) (string, error) {
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

func (c *fileCache) accessCachedFile(cachedFile *cachedFile) (string, error) {
	// Check if file still exists on disk
	if _, err := os.Stat(cachedFile.Path); os.IsNotExist(err) {
		// File was deleted, remove from cache
		c.mutex.Lock()
		delete(c.fileCache, cachedFile.Hash)
		c.stats.ItemCount--
		c.stats.Size -= cachedFile.Size
		c.mutex.Unlock()

		// Save updated metadata
		if saveErr := c.saveMetadata(); saveErr != nil {
			c.logger.Warnf("Failed to save cache metadata after removal: %v", saveErr)
		}

		return "", fmt.Errorf("cached file not found on disk: %s", cachedFile.Path)
	}

	cachedFile.LastAccessed = time.Now()

	// Save metadata to persist access time
	if err := c.saveMetadata(); err != nil {
		c.logger.Warnf("Failed to save cache metadata after access: %v", err)
	}

	return cachedFile.Path, nil
}

// storeFile now uses the key (CID or URL) as the filename (with .go extension)
func (c *fileCache) storeFile(key string, content []byte) (string, error) {
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

	cachedFile := &cachedFile{
		Path:         filePath,
		Hash:         key,
		LastAccessed: time.Now(),
		Size:         fileInfo.Size(),
		CreatedAt:    time.Now(),
	}

	c.mutex.Lock()
	c.fileCache[key] = cachedFile
	c.stats.ItemCount++
	c.stats.Size += fileInfo.Size()
	c.mutex.Unlock()

	// Save metadata to persist cache information
	if err := c.saveMetadata(); err != nil {
		c.logger.Warnf("Failed to save cache metadata: %v", err)
	}

	c.logger.Infof("Stored file in cache (size: %d bytes)", fileInfo.Size())
	return filePath, nil
}

func (c *fileCache) ensureSpace(requiredSize int64) error {
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
		file *cachedFile
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

	c.logger.Infof("Evicted %d bytes (%d files) from cache", evictedSize, c.stats.EvictionCount)

	// Save updated metadata after eviction
	if err := c.saveMetadata(); err != nil {
		c.logger.Warnf("Failed to save cache metadata after eviction: %v", err)
	}

	return nil
}

func (c *fileCache) loadExistingFiles() error {
	// Load metadata file if it exists
	metadataPath := filepath.Join(c.cacheDir, "cache_metadata.json")
	if _, err := os.Stat(metadataPath); err == nil {
		if err := c.loadMetadata(metadataPath); err != nil {
			c.logger.Warnf("Failed to load cache metadata: %v", err)
		}
	}

	// Also scan directory for any files not in metadata
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || entry.Name() == "cache_metadata.json" {
			continue
		}

		// Extract hash from filename
		key := strings.TrimSuffix(entry.Name(), ".go")
		filePath := filepath.Join(c.cacheDir, entry.Name())

		// Skip if already loaded from metadata
		if _, exists := c.fileCache[key]; exists {
			continue
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			c.logger.Warnf("Failed to stat cached file %s: %v", filePath, err)
			continue
		}

		cachedFile := &cachedFile{
			Path:         filePath,
			Hash:         key,
			LastAccessed: fileInfo.ModTime(),
			Size:         fileInfo.Size(),
			CreatedAt:    fileInfo.ModTime(),
		}

		c.fileCache[key] = cachedFile
		c.stats.ItemCount++
		c.stats.Size += fileInfo.Size()
	}

	c.logger.Infof("Loaded %d existing cached files", len(c.fileCache))
	return nil
}

func (c *fileCache) loadMetadata(metadataPath string) error {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var metadata map[string]*cachedFile
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	for key, cachedFile := range metadata {
		// Verify file still exists
		if _, err := os.Stat(cachedFile.Path); err == nil {
			c.fileCache[key] = cachedFile
			c.stats.ItemCount++
			c.stats.Size += cachedFile.Size
		}
	}

	return nil
}

func (c *fileCache) saveMetadata() error {
	metadataPath := filepath.Join(c.cacheDir, "cache_metadata.json")
	data, err := json.MarshalIndent(c.fileCache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func (c *fileCache) updateHitRate() {
	total := c.stats.HitCount + c.stats.MissCount
	if total > 0 {
		c.stats.HitRate = float64(c.stats.HitCount) / float64(total)
	}
}

func (c *fileCache) getCacheStats() *types.CacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	// Create a copy to avoid race conditions
	stats := *c.stats
	return &stats
}

func (c *fileCache) close() error {
	if c.cleanupTicker != nil {
		close(c.stopCleanup)
	}
	return nil
}
