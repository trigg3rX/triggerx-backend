package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SequentialRotator implements a log rotator with sequential numbering
type SequentialRotator struct {
	filename   string
	maxSize    int64 // bytes
	maxAge     int   // days
	maxBackups int   // number of old log files to keep
	compress   bool
	mu         sync.Mutex
	file       *os.File
	size       int64
}

// NewSequentialRotator creates a new sequential rotator
func NewSequentialRotator(filename string, maxSizeMB, maxAge, maxBackups int, compress bool) *SequentialRotator {
	return &SequentialRotator{
		filename:   filename,
		maxSize:    int64(maxSizeMB) * 1024 * 1024, // Convert MB to bytes
		maxAge:     maxAge,
		maxBackups: maxBackups,
		compress:   compress,
	}
}

// Write implements io.Writer
func (r *SequentialRotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Open file if not already open
	if r.file == nil {
		if err := r.openFile(); err != nil {
			return 0, err
		}
	}

	// Check if rotation is needed
	if r.size+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	// Write to file
	n, err = r.file.Write(p)
	r.size += int64(n)
	return n, err
}

// Close closes the current log file
func (r *SequentialRotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		err := r.file.Close()
		r.file = nil
		return err
	}
	return nil
}

// openFile opens the log file for writing
func (r *SequentialRotator) openFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(r.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Get current file size if it exists
	info, err := os.Stat(r.filename)
	if err == nil {
		r.size = info.Size()
	} else {
		r.size = 0
	}

	// Open file for appending
	file, err := os.OpenFile(r.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	r.file = file
	return nil
}

// rotate rotates the current log file
func (r *SequentialRotator) rotate() error {
	// Close current file
	if r.file != nil {
		if err := r.file.Close(); err != nil {
			return err
		}
		r.file = nil
	}

	// Get the next sequence number
	nextSeq := r.getNextSequenceNumber()

	// Generate new filename with sequence number
	base := strings.TrimSuffix(r.filename, ".log")
	rotatedName := fmt.Sprintf("%s.%d.log", base, nextSeq)

	// Rename current file to rotated name
	if err := os.Rename(r.filename, rotatedName); err != nil {
		return err
	}

	// Clean up old files
	r.cleanupOldFiles()

	// Open new file
	r.size = 0
	return r.openFile()
}

// getNextSequenceNumber finds the next available sequence number
func (r *SequentialRotator) getNextSequenceNumber() int {
	dir := filepath.Dir(r.filename)
	base := strings.TrimSuffix(filepath.Base(r.filename), ".log")

	// Get all rotated files
	files, err := filepath.Glob(filepath.Join(dir, base+".*.log"))
	if err != nil {
		return 1
	}

	maxSeq := 0
	for _, file := range files {
		baseName := filepath.Base(file)
		// Extract sequence number from filename like "2025-07-01.1.log"
		parts := strings.Split(baseName, ".")
		if len(parts) >= 3 {
			if seq, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
				if seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	return maxSeq + 1
}

// cleanupOldFiles removes old log files based on maxBackups and maxAge
func (r *SequentialRotator) cleanupOldFiles() {
	dir := filepath.Dir(r.filename)
	base := strings.TrimSuffix(filepath.Base(r.filename), ".log")

	// Get all rotated files
	files, err := filepath.Glob(filepath.Join(dir, base+".*.log"))
	if err != nil {
		return
	}

	// Sort files by modification time (newest first)
	type fileInfo struct {
		path    string
		modTime time.Time
		seq     int
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		baseName := filepath.Base(file)
		parts := strings.Split(baseName, ".")
		seq := 0
		if len(parts) >= 3 {
			if s, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
				seq = s
			}
		}

		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
			seq:     seq,
		})
	}

	// Sort by sequence number (descending)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].seq > fileInfos[j].seq
	})

	// Remove files that exceed maxBackups
	if r.maxBackups > 0 && len(fileInfos) > r.maxBackups {
		for i := r.maxBackups; i < len(fileInfos); i++ {
			err := os.Remove(fileInfos[i].path)
			if err != nil {
				log.Println("Failed to remove log file: ", err)
			}
		}
		fileInfos = fileInfos[:r.maxBackups]
	}

	// Remove files older than maxAge
	if r.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -r.maxAge)
		for _, fi := range fileInfos {
			if fi.modTime.Before(cutoff) {
				err := os.Remove(fi.path)
				if err != nil {
					log.Println("Failed to remove log file: ", err)
				}
			}
		}
	}
}
