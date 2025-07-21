package types

import "time"

type CacheStats struct {
	HitCount      int64     `json:"hit_count"`
	MissCount     int64     `json:"miss_count"`
	HitRate       float64   `json:"hit_rate"`
	Size          int64     `json:"size"`
	MaxSize       int64     `json:"max_size"`
	ItemCount     int       `json:"item_count"`
	EvictionCount int64     `json:"eviction_count"`
	LastCleanup   time.Time `json:"last_cleanup"`
}
