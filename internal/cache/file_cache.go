package cache

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
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
		return "", err
	}
	entry, ok := m[key]
	if !ok {
		return "", errors.New("key not found")
	}
	if entry.Expiration > 0 && time.Now().Unix() > entry.Expiration {
		// expired, delete
		delete(m, key)
		_ = f.save(m)
		return "", errors.New("key expired")
	}
	return entry.Value, nil
}

func (f *FileCache) Set(key string, value string, ttl time.Duration) error {
	m, err := f.load()
	if err != nil {
		return err
	}
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).Unix()
	}
	m[key] = fileCacheEntry{Value: value, Expiration: exp}
	return f.save(m)
}

func (f *FileCache) Delete(key string) error {
	m, err := f.load()
	if err != nil {
		return err
	}
	delete(m, key)
	return f.save(m)
}

// AcquirePerformerLock attempts to acquire a lock for a performer using file-based storage
func (f *FileCache) AcquirePerformerLock(performerID string, ttl time.Duration) (bool, error) {
	key := "performer:busy:" + performerID

	// Check if lock already exists and is not expired
	existingValue, err := f.Get(key)
	if err == nil && existingValue != "" {
		return false, nil // Lock already exists
	}

	// Set the lock with TTL
	err = f.Set(key, "1", ttl)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReleasePerformerLock releases a performer lock by deleting the key
func (f *FileCache) ReleasePerformerLock(performerID string) error {
	key := "performer:busy:" + performerID
	return f.Delete(key)
}
