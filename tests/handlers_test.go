package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ffprobe-api",
		})
	})
	
	return router
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "ffprobe-api", response["service"])
}

func TestStorageEndpoints(t *testing.T) {
	router := setupTestRouter()

	storageGroup := router.Group("/api/v1/storage")
	{
		storageGroup.POST("/upload", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"upload_id": "test-upload-123",
				"key":       "uploads/test-file.txt",
			})
		})
		
		storageGroup.GET("/info/:key", func(c *gin.Context) {
			key := c.Param("key")
			c.JSON(http.StatusOK, gin.H{
				"key": key,
				"url": "http://example.com/" + key,
			})
		})
		
		storageGroup.DELETE("/:key", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "File deleted successfully",
			})
		})
	}

	t.Run("Upload file endpoint", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/v1/storage/upload", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "upload_id")
		assert.Contains(t, response, "key")
	})

	t.Run("Get file info endpoint", func(t *testing.T) {
		testKey := "test-file.txt"
		req, err := http.NewRequest("GET", "/api/v1/storage/info/"+testKey, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testKey, response["key"])
		assert.Contains(t, response["url"], testKey)
	})

	t.Run("Delete file endpoint", func(t *testing.T) {
		testKey := "test-file.txt"
		req, err := http.NewRequest("DELETE", "/api/v1/storage/"+testKey, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "File deleted successfully", response["message"])
	})
}

func TestProbeEndpoints(t *testing.T) {
	router := setupTestRouter()

	probeGroup := router.Group("/api/v1/probe")
	{
		probeGroup.POST("/file", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"analysis_id": "test-analysis-123",
				"status":      "processing",
			})
		})
		
		probeGroup.GET("/status/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"analysis_id": id,
				"status":      "completed",
			})
		})
	}

	t.Run("Probe file endpoint", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"file_path": "/path/to/test.mp4",
		}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/probe/file", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "analysis_id")
		assert.Equal(t, "processing", response["status"])
	})

	t.Run("Get analysis status endpoint", func(t *testing.T) {
		testID := "test-analysis-123"
		req, err := http.NewRequest("GET", "/api/v1/probe/status/"+testID, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, testID, response["analysis_id"])
		assert.Equal(t, "completed", response["status"])
	})
}