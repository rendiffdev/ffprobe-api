package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
	RequestID string      `json:"request_id"`
}

func setupIntegrationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Header("X-Request-ID", "test-request-123")
		c.Next()
	})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, TestResponse{
				Status:    "healthy",
				Message:   "Service is running",
				RequestID: c.GetHeader("X-Request-ID"),
			})
		})

		storage := v1.Group("/storage")
		{
			storage.POST("/upload", func(c *gin.Context) {
				file, header, err := c.Request.FormFile("file")
				if err != nil {
					c.JSON(http.StatusBadRequest, TestResponse{
						Status:    "error",
						Error:     "No file provided",
						RequestID: c.GetHeader("X-Request-ID"),
					})
					return
				}
				defer file.Close()

				uploadID := "upload-" + header.Filename
				key := "uploads/" + header.Filename

				c.JSON(http.StatusOK, TestResponse{
					Status:  "success",
					Message: "File uploaded successfully",
					Data: map[string]interface{}{
						"upload_id": uploadID,
						"key":       key,
						"size":      header.Size,
						"filename":  header.Filename,
					},
					RequestID: c.GetHeader("X-Request-ID"),
				})
			})

			storage.GET("/info/:key", func(c *gin.Context) {
				key := c.Param("key")
				c.JSON(http.StatusOK, TestResponse{
					Status:  "success",
					Message: "File info retrieved",
					Data: map[string]interface{}{
						"key":  key,
						"url":  "http://example.com/" + key,
						"size": 1024,
					},
					RequestID: c.GetHeader("X-Request-ID"),
				})
			})
		}

		probe := v1.Group("/probe")
		{
			probe.POST("/file", func(c *gin.Context) {
				var req map[string]interface{}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, TestResponse{
						Status:    "error",
						Error:     "Invalid request body",
						RequestID: c.GetHeader("X-Request-ID"),
					})
					return
				}

				c.JSON(http.StatusOK, TestResponse{
					Status:  "success",
					Message: "Analysis started",
					Data: map[string]interface{}{
						"analysis_id": "analysis-123",
						"status":      "processing",
						"file_path":   req["file_path"],
					},
					RequestID: c.GetHeader("X-Request-ID"),
				})
			})
		}
	}

	return router
}

func TestHealthCheckIntegration(t *testing.T) {
	router := setupIntegrationRouter()

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TestResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "Service is running", response.Message)
	assert.Equal(t, "test-request-123", response.RequestID)
}

func TestFileUploadIntegration(t *testing.T) {
	router := setupIntegrationRouter()

	testContent := "This is a test file content"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "test-file.txt")
	require.NoError(t, err)

	_, err = io.WriteString(part, testContent)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/storage/upload", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TestResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "File uploaded successfully", response.Message)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "upload_id")
	assert.Contains(t, data, "key")
	assert.Equal(t, "test-file.txt", data["filename"])
}

func TestProbeIntegration(t *testing.T) {
	router := setupIntegrationRouter()

	reqBody := map[string]interface{}{
		"file_path": "/path/to/test-video.mp4",
		"options": map[string]interface{}{
			"include_streams": true,
			"include_format":  true,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/v1/probe/file", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TestResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "Analysis started", response.Message)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, data, "analysis_id")
	assert.Equal(t, "processing", data["status"])
}

func TestErrorHandlingIntegration(t *testing.T) {
	router := setupIntegrationRouter()

	t.Run("Upload without file", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/storage/upload", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response TestResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Error, "No file provided")
	})

	t.Run("Invalid probe request", func(t *testing.T) {
		invalidJSON := "invalid json"

		req, err := http.NewRequest("POST", "/api/v1/probe/file", strings.NewReader(invalidJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var response TestResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Error, "Invalid request body")
	})
}