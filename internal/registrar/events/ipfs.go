package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/registrar/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// Pinata v3 API structures
type PinataFile struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CID       string            `json:"cid"`
	Size      int64             `json:"size"`
	MimeType  string            `json:"mime_type"`
	GroupID   string            `json:"group_id,omitempty"`
	Keyvalues map[string]string `json:"keyvalues,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type PinataListResponse struct {
	Data struct {
		Files         []PinataFile `json:"files"`
		NextPageToken string       `json:"next_page_token,omitempty"`
	} `json:"data"`
}

type PinataDeleteResponse struct {
	Data interface{} `json:"data"`
}

// Helper function to make authenticated requests to Pinata
func makePinataRequest(method, url string, body io.Reader, logger logging.Logger) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.GetPinataJWT())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// Find file ID by CID using Pinata v3 API
func findPinataFileIDByCID(cid string, logger logging.Logger) (string, error) {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?cid=%s&limit=1", network, cid)

	resp, err := makePinataRequest("GET", url, nil, logger)
	if err != nil {
		return "", fmt.Errorf("failed to search for CID %s: %w", cid, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to search for CID %s: status %d, body: %s",
			cid, resp.StatusCode, string(body))
	}

	var listResp PinataListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return "", fmt.Errorf("failed to decode search response for CID %s: %w", cid, err)
	}

	if len(listResp.Data.Files) == 0 {
		return "", fmt.Errorf("no file found with CID %s", cid)
	}

	return listResp.Data.Files[0].ID, nil
}

// Delete file by ID using Pinata v3 API
func deletePinataFileByID(fileID string, logger logging.Logger) error {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s/%s", network, fileID)

	resp, err := makePinataRequest("DELETE", url, nil, logger)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileID, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to delete file %s: status %d, body: %s",
			fileID, resp.StatusCode, string(body))
	}

	var deleteResp PinataDeleteResponse
	if err := json.Unmarshal(body, &deleteResp); err != nil {
		// If we can't parse the response but got a success status, that's still OK
		logger.Debugf("Could not parse delete response for file %s, but status was %d", fileID, resp.StatusCode)
	}

	logger.Infof("Successfully deleted file %s from Pinata", fileID)
	return nil
}

// Delete file by CID (finds ID first, then deletes)
func DeletePinataCID(cid string, logger logging.Logger) error {
	fileID, err := findPinataFileIDByCID(cid, logger)
	if err != nil {
		return fmt.Errorf("failed to find file ID for CID %s: %w", cid, err)
	}

	return deletePinataFileByID(fileID, logger)
}

// List all files from Pinata v3 API
func listPinataFiles(logger logging.Logger) ([]PinataFile, error) {
	network := config.GetPinataHost() // Use configurable network from config
	url := fmt.Sprintf("https://api.pinata.cloud/v3/files/%s?limit=1000", network)

	var allFiles []PinataFile
	nextPageToken := ""

	for {
		requestURL := url
		if nextPageToken != "" {
			requestURL = fmt.Sprintf("%s&pageToken=%s", url, nextPageToken)
		}

		resp, err := makePinataRequest("GET", requestURL, nil, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if err := resp.Body.Close(); err != nil {
				logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to list files: status %d, body: %s",
				resp.StatusCode, string(body))
		}

		var listResp PinataListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			if err := resp.Body.Close(); err != nil {
				logger.Debugf("failed to close response body: %v", err)
			}
			return nil, fmt.Errorf("failed to decode list response: %w", err)
		}
		if err := resp.Body.Close(); err != nil {
			logger.Debugf("failed to close response body: %v", err)
		}

		allFiles = append(allFiles, listResp.Data.Files...)

		// Check if there are more pages
		if listResp.Data.NextPageToken == "" {
			break
		}
		nextPageToken = listResp.Data.NextPageToken
	}

	return allFiles, nil
}

// Schedule a 24h delayed deletion for a CID
func scheduleCIDDeletion(cid string, logger logging.Logger) {
	go func() {
		logger.Infof("Scheduled deletion for CID %s in 24h", cid)
		time.Sleep(24 * time.Hour)

		if err := DeletePinataCID(cid, logger); err != nil {
			logger.Errorf("Failed to delete CID %s after 24h: %v", cid, err)
		} else {
			logger.Infof("Successfully deleted CID %s after 24h delay", cid)
		}
	}()
}

// Weekly cleanup: delete files older than 1 day every Sunday
func StartWeeklyPinataCleanup(logger logging.Logger) {
	go func() {
		for {
			now := time.Now().UTC()
			// Calculate next Sunday 00:05 UTC
			daysUntilSunday := (7 - int(now.Weekday())) % 7
			if daysUntilSunday == 0 && now.Hour() >= 0 && now.Minute() >= 5 {
				daysUntilSunday = 7
			}
			nextSunday := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, time.UTC).AddDate(0, 0, daysUntilSunday)
			wait := nextSunday.Sub(now)

			logger.Infof("Weekly Pinata cleanup scheduled for: %v (in %v)", nextSunday, wait)
			time.Sleep(wait)

			logger.Info("Starting weekly Pinata cleanup: deleting files older than 1 day...")
			files, err := listPinataFiles(logger)
			if err != nil {
				logger.Errorf("Failed to list Pinata files: %v", err)
				continue
			}

			cutoff := time.Now().Add(-24 * time.Hour)
			deletedCount := 0

			for _, file := range files {
				if file.CreatedAt.Before(cutoff) {
					logger.Infof("Deleting old file %s (CID: %s) created at %s",
						file.ID, file.CID, file.CreatedAt)

					if err := deletePinataFileByID(file.ID, logger); err != nil {
						logger.Errorf("Failed to delete old file %s: %v", file.ID, err)
					} else {
						deletedCount++
					}

					// Add small delay between deletions to avoid rate limiting
					time.Sleep(100 * time.Millisecond)
				}
			}

			logger.Infof("Weekly cleanup completed. Deleted %d old files.", deletedCount)
		}
	}()
}