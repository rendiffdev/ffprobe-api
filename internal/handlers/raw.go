package handlers

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// RawHandler handles raw ffprobe data endpoints
type RawHandler struct {
	analysisService *services.AnalysisService
	logger          zerolog.Logger
}

// NewRawHandler creates a new raw data handler
func NewRawHandler(analysisService *services.AnalysisService, logger zerolog.Logger) *RawHandler {
	return &RawHandler{
		analysisService: analysisService,
		logger:          logger,
	}
}

// GetRawData returns raw ffprobe data in specified format
// @Summary Get raw ffprobe data
// @Description Get raw ffprobe data in JSON, CSV, or XML format
// @Tags probe
// @Accept json
// @Produce json,text/csv,application/xml
// @Param id path string true "Analysis ID"
// @Param format query string false "Output format (json, csv, xml)" default(json)
// @Param section query string false "Data section (format, streams, frames, all)" default(all)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/raw/{id} [get]
func (h *RawHandler) GetRawData(c *gin.Context) {
	analysisID := c.Param("id")
	format := c.DefaultQuery("format", "json")
	section := c.DefaultQuery("section", "all")

	// Validate UUID
	if _, err := uuid.Parse(analysisID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Validate format
	validFormats := map[string]bool{"json": true, "csv": true, "xml": true}
	if !validFormats[format] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid format. Supported formats: json, csv, xml",
		})
		return
	}

	// Validate section
	validSections := map[string]bool{"format": true, "streams": true, "frames": true, "all": true}
	if !validSections[section] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid section. Supported sections: format, streams, frames, all",
		})
		return
	}

	// Get analysis
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Analysis not found",
			})
			return
		}

		h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to get analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve analysis",
		})
		return
	}

	// Check if analysis has ffprobe data
	if result.Analysis.FFprobeData == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No ffprobe data available for this analysis",
		})
		return
	}

	// Extract requested section
	data := h.extractSection(result.Analysis.FFprobeData, section)

	// Return data in requested format
	switch format {
	case "json":
		h.returnJSON(c, data)
	case "csv":
		h.returnCSV(c, data, section)
	case "xml":
		h.returnXML(c, data, section)
	}
}

// GetRawStreams returns raw streams data
// @Summary Get raw streams data
// @Description Get detailed streams information in various formats
// @Tags probe
// @Accept json
// @Produce json,text/csv,application/xml
// @Param id path string true "Analysis ID"
// @Param format query string false "Output format (json, csv, xml)" default(json)
// @Param stream_index query int false "Specific stream index"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/raw/{id}/streams [get]
func (h *RawHandler) GetRawStreams(c *gin.Context) {
	analysisID := c.Param("id")
	format := c.DefaultQuery("format", "json")
	streamIndexStr := c.Query("stream_index")

	// Validate UUID
	if _, err := uuid.Parse(analysisID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get analysis
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Analysis not found",
			})
			return
		}

		h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to get analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve analysis",
		})
		return
	}

	// Extract streams data
	streams, ok := result.Analysis.FFprobeData["streams"]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No streams data available",
		})
		return
	}

	// Filter by stream index if specified
	var data interface{} = streams
	if streamIndexStr != "" {
		streamIndex, err := strconv.Atoi(streamIndexStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid stream index",
			})
			return
		}

		if streamsList, ok := streams.([]interface{}); ok {
			if streamIndex >= 0 && streamIndex < len(streamsList) {
				data = streamsList[streamIndex]
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Stream index not found",
				})
				return
			}
		}
	}

	// Return data in requested format
	switch format {
	case "json":
		h.returnJSON(c, data)
	case "csv":
		h.returnCSV(c, data, "streams")
	case "xml":
		h.returnXML(c, data, "streams")
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid format. Supported formats: json, csv, xml",
		})
	}
}

// GetRawFormat returns raw format data
// @Summary Get raw format data
// @Description Get detailed format information in various formats
// @Tags probe
// @Accept json
// @Produce json,text/csv,application/xml
// @Param id path string true "Analysis ID"
// @Param format query string false "Output format (json, csv, xml)" default(json)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/probe/raw/{id}/format [get]
func (h *RawHandler) GetRawFormat(c *gin.Context) {
	analysisID := c.Param("id")
	format := c.DefaultQuery("format", "json")

	// Validate UUID
	if _, err := uuid.Parse(analysisID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid analysis ID format",
		})
		return
	}

	// Get analysis
	result, err := h.analysisService.GetAnalysisResult(c.Request.Context(), analysisID)
	if err != nil {
		if err == services.ErrAnalysisNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Analysis not found",
			})
			return
		}

		h.logger.Error().Err(err).Str("analysis_id", analysisID).Msg("Failed to get analysis")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve analysis",
		})
		return
	}

	// Extract format data
	formatData, ok := result.Analysis.FFprobeData["format"]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No format data available",
		})
		return
	}

	// Return data in requested format
	switch format {
	case "json":
		h.returnJSON(c, formatData)
	case "csv":
		h.returnCSV(c, formatData, "format")
	case "xml":
		h.returnXML(c, formatData, "format")
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid format. Supported formats: json, csv, xml",
		})
	}
}

// Helper methods

func (h *RawHandler) extractSection(data map[string]interface{}, section string) interface{} {
	switch section {
	case "format":
		if format, ok := data["format"]; ok {
			return format
		}
		return nil
	case "streams":
		if streams, ok := data["streams"]; ok {
			return streams
		}
		return nil
	case "frames":
		if frames, ok := data["frames"]; ok {
			return frames
		}
		return nil
	case "all":
		return data
	default:
		return data
	}
}

func (h *RawHandler) returnJSON(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, data)
}

func (h *RawHandler) returnCSV(c *gin.Context, data interface{}, section string) {
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"ffprobe_%s.csv\"", section))

	// Convert data to CSV
	csvData, err := h.convertToCSV(data, section)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to convert data to CSV")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to convert data to CSV",
		})
		return
	}

	c.String(http.StatusOK, csvData)
}

func (h *RawHandler) returnXML(c *gin.Context, data interface{}, section string) {
	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"ffprobe_%s.xml\"", section))

	// Convert data to XML
	xmlData, err := h.convertToXML(data, section)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to convert data to XML")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to convert data to XML",
		})
		return
	}

	c.String(http.StatusOK, xmlData)
}

func (h *RawHandler) convertToCSV(data interface{}, section string) (string, error) {
	var records [][]string

	switch section {
	case "format":
		if formatMap, ok := data.(map[string]interface{}); ok {
			records = append(records, []string{"Property", "Value"})
			for key, value := range formatMap {
				records = append(records, []string{key, fmt.Sprintf("%v", value)})
			}
		}
	case "streams":
		if streamsList, ok := data.([]interface{}); ok {
			// Get headers from first stream
			if len(streamsList) > 0 {
				if firstStream, ok := streamsList[0].(map[string]interface{}); ok {
					var headers []string
					for key := range firstStream {
						headers = append(headers, key)
					}
					records = append(records, headers)

					// Add data rows
					for _, stream := range streamsList {
						if streamMap, ok := stream.(map[string]interface{}); ok {
							var row []string
							for _, header := range headers {
								if value, exists := streamMap[header]; exists {
									row = append(row, fmt.Sprintf("%v", value))
								} else {
									row = append(row, "")
								}
							}
							records = append(records, row)
						}
					}
				}
			}
		}
	default:
		// Convert any data to key-value pairs
		if dataMap, ok := data.(map[string]interface{}); ok {
			records = append(records, []string{"Property", "Value"})
			for key, value := range dataMap {
				records = append(records, []string{key, fmt.Sprintf("%v", value)})
			}
		}
	}

	// Convert records to CSV string
	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}
	writer.Flush()
	return csvBuilder.String(), nil
}

func (h *RawHandler) convertToXML(data interface{}, section string) (string, error) {
	// Create XML wrapper
	type XMLWrapper struct {
		XMLName xml.Name    `xml:"ffprobe"`
		Data    interface{} `xml:",innerxml"`
	}

	// Convert data to XML-friendly format
	xmlData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Create basic XML structure
	xmlString := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<%s>\n%s\n</%s>", section, string(xmlData), section)
	return xmlString, nil
}