package cache

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type fileCacheEntry struct {
	Value      string `json:"value"`
	Expiration int64  `json:"expiration"` // unix timestamp
}

type FileCache struct{}

var (
	fileCachePath = "cache.json"
	fileCacheMu   sync.Mutex
)

func (f *FileCache) load() (map[string]fileCacheEntry, error) {
	fileCacheMu.Lock()
	defer fileCacheMu.Unlock()
	m := make(map[string]fileCacheEntry)
	file, err := os.Open(fileCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func (f *FileCache) save(m map[string]fileCacheEntry) error {
	fileCacheMu.Lock()
	defer fileCacheMu.Unlock()
	file, err := os.Create(fileCachePath)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	return enc.Encode(m)
}

func (f *FileCache) Get(key string) (string, error) {
	m, err := f.load()
	if err != nil {
		logging.GetServiceLogger().Errorf("[FileCache] Error loading cache: %v", err)
		return "", err
	}
	entry, ok := m[key]
	if !ok {
		logging.GetServiceLogger().Warnf("[FileCache] Cache miss for key: %s", key)
		return "", errors.New("key not found")
	}
	if entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
		// expired, delete
		delete(m, key)
		_ = f.save(m)
		logging.GetServiceLogger().Warnf("[FileCache] Cache expired for key: %s", key)
		return "", errors.New("key expired")
	}
	logging.GetServiceLogger().Infof("[FileCache] Cache hit for key: %s", key)
	return entry.Value, nil
}

func (f *FileCache) Set(key string, value string, ttl time.Duration) error {
	m, err := f.load()
	if err != nil {
		logging.GetServiceLogger().Errorf("[FileCache] Error loading cache for set: %v", err)
		return err
	}
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).Unix()
	}
	m[key] = fileCacheEntry{Value: value, Expiration: exp}
	err = f.save(m)
	if err != nil {
		logging.GetServiceLogger().Errorf("[FileCache] Error saving cache for set: %v", err)
	} else {
		logging.GetServiceLogger().Infof("[FileCache] Set key: %s (ttl: %v)", key, ttl)
	}
	return err
}

func (f *FileCache) Delete(key string) error {
	m, err := f.load()
	if err != nil {
		logging.GetServiceLogger().Errorf("[FileCache] Error loading cache for delete: %v", err)
		return err
	}
	delete(m, key)
	err = f.save(m)
	if err != nil {
		logging.GetServiceLogger().Errorf("[FileCache] Error saving cache for delete: %v", err)
	} else {
		logging.GetServiceLogger().Infof("[FileCache] Deleted key: %s", key)
	}
	return err
}

func (f *FileCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	logging.GetServiceLogger().Infof("[FileCache] AcquirePerformerLock called for performerID: %s (no-op)", performerID)
	return true, nil
}

func (f *FileCache) ReleasePerformerLock(performerID string) error {
	logging.GetServiceLogger().Infof("[FileCache] ReleasePerformerLock called for performerID: %s (no-op)", performerID)
	return nil
}
