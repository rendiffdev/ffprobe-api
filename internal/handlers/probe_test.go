package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
	"github.com/rendiffdev/ffprobe-api/internal/models"
)

// MockAnalysisService is a mock implementation of the analysis service
type MockAnalysisService struct {
	mock.Mock
}

func (m *MockAnalysisService) CreateAnalysis(ctx context.Context, request *models.CreateAnalysisRequest) (*models.Analysis, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*models.Analysis), args.Error(1)
}

func (m *MockAnalysisService) ProcessAnalysis(ctx context.Context, analysisID uuid.UUID, options *ffmpeg.FFprobeOptions) error {
	args := m.Called(ctx, analysisID, options)
	return args.Error(0)
}

func (m *MockAnalysisService) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Analysis), args.Error(1)
}

func (m *MockAnalysisService) GetAnalysesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Analysis, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Analysis), args.Error(1)
}

func (m *MockAnalysisService) DeleteAnalysis(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAnalysisService) CheckFFprobeAvailability(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAnalysisService) GetFFprobeVersion(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestProbeHandler_Health(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.GET("/health", handler.Health)

	// Test case: healthy
	mockService.On("CheckFFprobeAvailability", mock.Anything).Return(nil)
	mockService.On("GetFFprobeVersion", mock.Anything).Return("ffprobe version 4.4.0", nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "ffprobe version 4.4.0", response.FFprobeVersion)

	mockService.AssertExpectations(t)
}

func TestProbeHandler_Health_Unhealthy(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.GET("/health", handler.Health)

	// Test case: unhealthy
	mockService.On("CheckFFprobeAvailability", mock.Anything).Return(assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", response.Status)
	assert.Contains(t, response.Error, "assert.AnError")

	mockService.AssertExpectations(t)
}

func TestProbeHandler_ProbeFile(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.POST("/probe/file", handler.ProbeFile)

	analysisID := uuid.New()
	
	// Mock analysis creation and processing
	analysis := &models.Analysis{
		ID:       analysisID,
		FileName: "test.mp4",
		FilePath: "/path/to/test.mp4",
		Status:   models.StatusCompleted,
	}

	mockService.On("CreateAnalysis", mock.Anything, mock.MatchedBy(func(req *models.CreateAnalysisRequest) bool {
		return req.FileName == "test.mp4" && req.FilePath == "/path/to/test.mp4"
	})).Return(analysis, nil)
	
	mockService.On("ProcessAnalysis", mock.Anything, analysisID, mock.Anything).Return(nil)
	mockService.On("GetAnalysis", mock.Anything, analysisID).Return(analysis, nil)

	// Test synchronous probe
	requestBody := ProbeFileRequest{
		FilePath:   "/path/to/test.mp4",
		SourceType: "local",
		Async:      false,
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/probe/file", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, analysisID, response.AnalysisID)
	assert.Equal(t, "completed", response.Status)
	assert.NotNil(t, response.Analysis)

	mockService.AssertExpectations(t)
}

func TestProbeHandler_ProbeFile_Async(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.POST("/probe/file", handler.ProbeFile)

	analysisID := uuid.New()
	
	analysis := &models.Analysis{
		ID:       analysisID,
		FileName: "test.mp4",
		FilePath: "/path/to/test.mp4",
		Status:   models.StatusPending,
	}

	mockService.On("CreateAnalysis", mock.Anything, mock.MatchedBy(func(req *models.CreateAnalysisRequest) bool {
		return req.FileName == "test.mp4" && req.FilePath == "/path/to/test.mp4"
	})).Return(analysis, nil)

	// Test async probe
	requestBody := ProbeFileRequest{
		FilePath:   "/path/to/test.mp4",
		SourceType: "local",
		Async:      true,
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/probe/file", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	
	var response ProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, analysisID, response.AnalysisID)
	assert.Equal(t, "processing", response.Status)
	assert.Contains(t, response.Message, "check status endpoint")

	mockService.AssertExpectations(t)
}

func TestProbeHandler_GetAnalysisStatus(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.GET("/status/:id", handler.GetAnalysisStatus)

	analysisID := uuid.New()
	analysis := &models.Analysis{
		ID:       analysisID,
		FileName: "test.mp4",
		Status:   models.StatusCompleted,
	}

	mockService.On("GetAnalysis", mock.Anything, analysisID).Return(analysis, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/"+analysisID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, analysisID, response.AnalysisID)
	assert.Equal(t, models.StatusCompleted, response.Status)

	mockService.AssertExpectations(t)
}

func TestProbeHandler_GetAnalysisStatus_InvalidID(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.GET("/status/:id", handler.GetAnalysisStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status/invalid-id", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid analysis ID")
}

func TestProbeHandler_ListAnalyses(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.GET("/analyses", handler.ListAnalyses)

	analyses := []models.Analysis{
		{ID: uuid.New(), FileName: "test1.mp4", Status: models.StatusCompleted},
		{ID: uuid.New(), FileName: "test2.mp4", Status: models.StatusProcessing},
	}

	mockService.On("GetAnalysesByUser", mock.Anything, mock.Anything, 20, 0).Return(analyses, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/analyses", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ListAnalysesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Analyses, 2)
	assert.Equal(t, 20, response.Limit)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 2, response.Count)

	mockService.AssertExpectations(t)
}

func TestProbeHandler_DeleteAnalysis(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.DELETE("/analyses/:id", handler.DeleteAnalysis)

	analysisID := uuid.New()
	mockService.On("DeleteAnalysis", mock.Anything, analysisID).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/analyses/"+analysisID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	mockService.AssertExpectations(t)
}

func TestProbeHandler_ProbeFile_InvalidRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.POST("/probe/file", handler.ProbeFile)

	// Test with invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/probe/file", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Invalid request body")
}

func TestProbeHandler_QuickProbe(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := new(MockAnalysisService)
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	handler := NewProbeHandler(mockService, logger)

	router := gin.New()
	router.POST("/quick", handler.QuickProbe)

	analysisID := uuid.New()
	
	analysis := &models.Analysis{
		ID:       analysisID,
		FileName: "test.mp4",
		FilePath: "/path/to/test.mp4",
		Status:   models.StatusCompleted,
	}

	mockService.On("CreateAnalysis", mock.Anything, mock.MatchedBy(func(req *models.CreateAnalysisRequest) bool {
		return req.FileName == "test.mp4" && req.FilePath == "/path/to/test.mp4"
	})).Return(analysis, nil)
	
	mockService.On("ProcessAnalysis", mock.Anything, analysisID, mock.MatchedBy(func(options *ffmpeg.FFprobeOptions) bool {
		// Verify quick options are set
		return options.ProbeSize == 1024*1024 && options.AnalyzeDuration == 1000000
	})).Return(nil)
	
	mockService.On("GetAnalysis", mock.Anything, analysisID).Return(analysis, nil)

	requestBody := ProbeFileRequest{
		FilePath:   "/path/to/test.mp4",
		SourceType: "local",
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/quick", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ProbeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, analysisID, response.AnalysisID)
	assert.Equal(t, "completed", response.Status)

	mockService.AssertExpectations(t)
}