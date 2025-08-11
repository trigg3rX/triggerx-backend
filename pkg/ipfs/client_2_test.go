package ipfs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Unit Tests for Delete method
func TestDelete_ValidCID_ReturnsNoError(t *testing.T) {
	client, mockHTTP, mockLogger := createTestClient()

	// Mock search response to find file ID
	searchResponse := `{
		"data": {
			"files": [
				{
					"id": "test-file-id-123",
					"name": "test.txt",
					"cid": "QmTestCID123",
					"size": 1024,
					"mime_type": "text/plain",
					"created_at": "2023-01-01T00:00:00Z"
				}
			]
		}
	}`
	mockSearchResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	// Mock delete response
	deleteResponse := `{"data": "success"}`
	mockDeleteResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(deleteResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockSearchResp, nil)
	mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/test-file-id-123").Return(mockDeleteResp, nil)
	mockLogger.On("Infof", "Successfully deleted file %s from Pinata", []interface{}{"test-file-id-123"}).Return()

	err := client.Delete("QmTestCID123")

	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDelete_EmptyCID_ReturnsError(t *testing.T) {
	client, _, _ := createTestClient()

	err := client.Delete("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CID cannot be empty")
}

func TestDelete_SearchHTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(nil, fmt.Errorf("network error"))

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search for CID QmTestCID123")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_SearchNonOKStatus_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(strings.NewReader("bad request")),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockResponse, nil)

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search for CID QmTestCID123: status 400")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_InvalidSearchJSON_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{"invalid": "json`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockResponse, nil)

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode search response for CID QmTestCID123")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_NoFileFound_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	searchResponse := `{
		"data": {
			"files": []
		}
	}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockResponse, nil)

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no file found with CID QmTestCID123")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_DeleteHTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	// Mock successful search response
	searchResponse := `{
		"data": {
			"files": [
				{
					"id": "test-file-id-123",
					"name": "test.txt",
					"cid": "QmTestCID123"
				}
			]
		}
	}`
	mockSearchResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockSearchResp, nil)
	mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/test-file-id-123").Return(nil, fmt.Errorf("delete network error"))

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete file test-file-id-123")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_DeleteNonOKStatus_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	// Mock successful search response
	searchResponse := `{
		"data": {
			"files": [
				{
					"id": "test-file-id-123",
					"name": "test.txt",
					"cid": "QmTestCID123"
				}
			]
		}
	}`
	mockSearchResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	// Mock failed delete response
	mockDeleteResp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("file not found")),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockSearchResp, nil)
	mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/test-file-id-123").Return(mockDeleteResp, nil)

	err := client.Delete("QmTestCID123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete file test-file-id-123: status 404")
	mockHTTP.AssertExpectations(t)
}

func TestDelete_AcceptedStatus_ReturnsNoError(t *testing.T) {
	client, mockHTTP, mockLogger := createTestClient()

	// Mock search response
	searchResponse := `{
		"data": {
			"files": [
				{
					"id": "test-file-id-123",
					"name": "test.txt",
					"cid": "QmTestCID123"
				}
			]
		}
	}`
	mockSearchResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	// Mock accepted delete response
	deleteResponse := `{"data": "accepted"}`
	mockDeleteResp := &http.Response{
		StatusCode: http.StatusAccepted,
		Body:       io.NopCloser(strings.NewReader(deleteResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmTestCID123&limit=1").Return(mockSearchResp, nil)
	mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/test-file-id-123").Return(mockDeleteResp, nil)
	mockLogger.On("Infof", "Successfully deleted file %s from Pinata", []interface{}{"test-file-id-123"}).Return()

	err := client.Delete("QmTestCID123")

	assert.NoError(t, err)
	mockHTTP.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Unit Tests for ListFiles method
func TestListFiles_SinglePage_ReturnsFiles(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{
		"data": {
			"files": [
				{
					"id": "file-1",
					"name": "test1.txt",
					"cid": "QmCID1",
					"size": 1024,
					"mime_type": "text/plain",
					"created_at": "2023-01-01T00:00:00Z"
				},
				{
					"id": "file-2",
					"name": "test2.txt",
					"cid": "QmCID2",
					"size": 2048,
					"mime_type": "text/plain",
					"created_at": "2023-01-02T00:00:00Z"
				}
			]
		}
	}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

	files, err := client.ListFiles()

	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "file-1", files[0].ID)
	assert.Equal(t, "test1.txt", files[0].Name)
	assert.Equal(t, "QmCID1", files[0].CID)
	assert.Equal(t, int64(1024), files[0].Size)
	assert.Equal(t, "text/plain", files[0].MimeType)
	assert.Equal(t, "file-2", files[1].ID)
	assert.Equal(t, "test2.txt", files[1].Name)
	assert.Equal(t, "QmCID2", files[1].CID)
	assert.Equal(t, int64(2048), files[1].Size)
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_MultiplePages_ReturnsAllFiles(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	// First page response
	firstPageResponse := `{
		"data": {
			"files": [
				{
					"id": "file-1",
					"name": "test1.txt",
					"cid": "QmCID1",
					"size": 1024,
					"mime_type": "text/plain",
					"created_at": "2023-01-01T00:00:00Z"
				}
			],
			"next_page_token": "page2"
		}
	}`
	mockFirstPageResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(firstPageResponse)),
	}

	// Second page response
	secondPageResponse := `{
		"data": {
			"files": [
				{
					"id": "file-2",
					"name": "test2.txt",
					"cid": "QmCID2",
					"size": 2048,
					"mime_type": "text/plain",
					"created_at": "2023-01-02T00:00:00Z"
				}
			]
		}
	}`
	mockSecondPageResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(secondPageResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockFirstPageResp, nil)
	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000&pageToken=page2").Return(mockSecondPageResp, nil)

	files, err := client.ListFiles()

	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "file-1", files[0].ID)
	assert.Equal(t, "file-2", files[1].ID)
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_EmptyResponse_ReturnsEmptySlice(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{
		"data": {
			"files": []
		}
	}`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

	files, err := client.ListFiles()

	assert.NoError(t, err)
	assert.Len(t, files, 0)
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(nil, fmt.Errorf("network error"))

	files, err := client.ListFiles()

	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to list files")
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_NonOKStatus_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	mockResponse := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       io.NopCloser(strings.NewReader("unauthorized")),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

	files, err := client.ListFiles()

	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to list files: status 401")
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_InvalidJSONResponse_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	responseBody := `{"invalid": "json`
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

	files, err := client.ListFiles()

	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to decode list response")
	mockHTTP.AssertExpectations(t)
}

func TestListFiles_SecondPageHTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP, _ := createTestClient()

	// First page response with next page token
	firstPageResponse := `{
		"data": {
			"files": [
				{
					"id": "file-1",
					"name": "test1.txt",
					"cid": "QmCID1"
				}
			],
			"next_page_token": "page2"
		}
	}`
	mockFirstPageResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(firstPageResponse)),
	}

	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockFirstPageResp, nil)
	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000&pageToken=page2").Return(nil, fmt.Errorf("second page error"))

	files, err := client.ListFiles()

	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to list files")
	mockHTTP.AssertExpectations(t)
}

// Edge case tests for Delete and ListFiles
func TestDelete_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		cid         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "very long CID",
			cid:         "Qm" + strings.Repeat("a", 100),
			expectError: false,
		},
		{
			name:        "CID with special characters",
			cid:         "Qm@#$%^&*()",
			expectError: false,
		},
		{
			name:        "CID with spaces",
			cid:         "Qm Test CID",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP, mockLogger := createTestClient()

			if !tt.expectError {
				// Mock successful search and delete responses
				searchResponse := fmt.Sprintf(`{
					"data": {
						"files": [
							{
								"id": "test-file-id",
								"name": "test.txt",
								"cid": "%s"
							}
						]
					}
				}`, tt.cid)
				mockSearchResp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(searchResponse)),
				}

				deleteResponse := `{"data": "success"}`
				mockDeleteResp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(deleteResponse)),
				}

				expectedSearchURL := fmt.Sprintf("https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=%s&limit=1", tt.cid)
				mockHTTP.On("Get", expectedSearchURL).Return(mockSearchResp, nil)
				mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/test-file-id").Return(mockDeleteResp, nil)
				mockLogger.On("Infof", "Successfully deleted file %s from Pinata", []interface{}{"test-file-id"}).Return()
			}

			err := client.Delete(tt.cid)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			mockHTTP.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestListFiles_EdgeCases(t *testing.T) {
	t.Run("files with missing optional fields", func(t *testing.T) {
		client, mockHTTP, _ := createTestClient()

		responseBody := `{
			"data": {
				"files": [
					{
						"id": "file-1",
						"name": "test1.txt",
						"cid": "QmCID1",
						"size": 1024,
						"mime_type": "text/plain"
					},
					{
						"id": "file-2",
						"name": "test2.txt",
						"cid": "QmCID2"
					}
				]
			}
		}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}

		mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

		files, err := client.ListFiles()

		assert.NoError(t, err)
		assert.Len(t, files, 2)
		assert.Equal(t, "file-1", files[0].ID)
		assert.Equal(t, "file-2", files[1].ID)
		// Second file should have zero values for missing fields
		assert.Equal(t, int64(0), files[1].Size)
		assert.Equal(t, "", files[1].MimeType)
		mockHTTP.AssertExpectations(t)
	})

	t.Run("files with keyvalues and group_id", func(t *testing.T) {
		client, mockHTTP, _ := createTestClient()

		responseBody := `{
			"data": {
				"files": [
					{
						"id": "file-1",
						"name": "test1.txt",
						"cid": "QmCID1",
						"size": 1024,
						"mime_type": "text/plain",
						"group_id": "group-123",
						"keyvalues": {
							"key1": "value1",
							"key2": "value2"
						},
						"created_at": "2023-01-01T00:00:00Z"
					}
				]
			}
		}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}

		mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

		files, err := client.ListFiles()

		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "group-123", files[0].GroupID)
		assert.Equal(t, "value1", files[0].Keyvalues["key1"])
		assert.Equal(t, "value2", files[0].Keyvalues["key2"])
		mockHTTP.AssertExpectations(t)
	})
}

// Integration Tests for Delete and ListFiles
func TestIPFSClientIntegration_UploadDeleteList_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client, mockHTTP, mockLogger := createTestClient()

	// Mock upload response
	uploadResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"data": {"cid": "QmIntegrationTestCID"}}`)),
	}
	mockHTTP.On("DoWithRetry", mock.AnythingOfType("*http.Request")).Return(uploadResponse, nil)
	mockLogger.On("Info", "Successfully uploaded to IPFS", mock.Anything, mock.Anything).Return()

	// Mock search response for delete
	searchResponse := `{
		"data": {
			"files": [
				{
					"id": "integration-file-id",
					"name": "integration-test.txt",
					"cid": "QmIntegrationTestCID"
				}
			]
		}
	}`
	mockSearchResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(searchResponse)),
	}

	// Mock delete response
	deleteResponse := `{"data": "success"}`
	mockDeleteResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(deleteResponse)),
	}

	// Mock list files response
	listResponse := `{
		"data": {
			"files": [
				{
					"id": "other-file-id",
					"name": "other-file.txt",
					"cid": "QmOtherCID"
				}
			]
		}
	}`
	mockListResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(listResponse)),
	}

	// Mock Close method
	mockHTTP.On("Close").Return()

	// Set up expectations
	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmIntegrationTestCID&limit=1").Return(mockSearchResp, nil)
	mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/integration-file-id").Return(mockDeleteResp, nil)
	mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockListResp, nil)
	mockLogger.On("Infof", "Successfully deleted file %s from Pinata", []interface{}{"integration-file-id"}).Return()

	// Test upload
	filename := "integration-test.txt"
	data := []byte("integration test data")
	cid, err := client.Upload(filename, data)
	assert.NoError(t, err)
	assert.Equal(t, "QmIntegrationTestCID", cid)

	// Test delete
	err = client.Delete(cid)
	assert.NoError(t, err)

	// Test list files
	files, err := client.ListFiles()
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "other-file-id", files[0].ID)

	// Close client
	err = client.Close()
	assert.NoError(t, err)

	mockHTTP.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Benchmark Tests for Delete and ListFiles
func BenchmarkDelete_SingleFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, mockHTTP, mockLogger := createTestClient()

		// Mock responses
		searchResponse := `{
			"data": {
				"files": [
					{
						"id": "benchmark-file-id",
						"name": "benchmark.txt",
						"cid": "QmBenchmarkCID"
					}
				]
			}
		}`
		mockSearchResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(searchResponse)),
		}

		deleteResponse := `{"data": "success"}`
		mockDeleteResp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(deleteResponse)),
		}

		mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?cid=QmBenchmarkCID&limit=1").Return(mockSearchResp, nil)
		mockHTTP.On("Delete", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud/benchmark-file-id").Return(mockDeleteResp, nil)
		mockLogger.On("Infof", "Successfully deleted file %s from Pinata", []interface{}{"benchmark-file-id"}).Return()

		err := client.Delete("QmBenchmarkCID")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListFiles_SinglePage(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, mockHTTP, _ := createTestClient()

		responseBody := `{
			"data": {
				"files": [
					{
						"id": "file-1",
						"name": "test1.txt",
						"cid": "QmCID1",
						"size": 1024,
						"mime_type": "text/plain"
					}
				]
			}
		}`
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}

		mockHTTP.On("Get", "https://api.pinata.cloud/v3/files/gateway.pinata.cloud?limit=1000").Return(mockResponse, nil)

		_, err := client.ListFiles()
		if err != nil {
			b.Fatal(err)
		}
	}
}
