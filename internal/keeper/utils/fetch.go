package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func FetchDataFromUrl(url string, logger logging.Logger) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch data from url: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			logger.Error("Error closing response body", "error", err)
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
