package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// LLMEnhancedAnalyzer provides AI-powered quality control analysis and reporting
type LLMEnhancedAnalyzer struct {
	ollamaURL    string
	modelName    string
	fallbackModel string
	httpClient   *http.Client
	logger       zerolog.Logger
	enabled      bool
}

// NewLLMEnhancedAnalyzer creates a new LLM-enhanced analyzer
func NewLLMEnhancedAnalyzer(ollamaURL, modelName, fallbackModel string, logger zerolog.Logger) *LLMEnhancedAnalyzer {
	return &LLMEnhancedAnalyzer{
		ollamaURL:     ollamaURL,
		modelName:     modelName,
		fallbackModel: fallbackModel,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		logger:  logger,
		enabled: ollamaURL != "",
	}
}

// LLMEnhancedReport contains AI-generated analysis and insights
type LLMEnhancedReport struct {
	TechnicalSummary            string                      `json:"technical_summary"`
	QualityAssessment           string                      `json:"quality_assessment"`
	ComplianceAnalysis          string                      `json:"compliance_analysis"`
	OptimizationRecommendations string                      `json:"optimization_recommendations"`
	IssueIdentification         []QualityIssue              `json:"issue_identification"`
	WorkflowRecommendations     []WorkflowRecommendation    `json:"workflow_recommendations"`
	ExecutiveSummary            string                      `json:"executive_summary"`
	AdvancedQCInsights          *AdvancedQCInsights         `json:"advanced_qc_insights,omitempty"`
	RiskAssessment              *RiskAssessment             `json:"risk_assessment,omitempty"`
	IntegrationRecommendations  []IntegrationRecommendation `json:"integration_recommendations,omitempty"`
	ProcessingTime              time.Duration               `json:"processing_time"`
	ModelUsed                   string                      `json:"model_used"`
	Confidence                  float64                     `json:"confidence"`
}

// QualityIssue represents an identified quality or technical issue
// AdvancedQCInsights contains insights from advanced QC analysis
type AdvancedQCInsights struct {
	TimecodeInsights       string   `json:"timecode_insights,omitempty"`
	AFDInsights            string   `json:"afd_insights,omitempty"`
	TransportStreamInsights string  `json:"transport_stream_insights,omitempty"`
	EndiannessInsights     string   `json:"endianness_insights,omitempty"`
	AudioWrappingInsights  string   `json:"audio_wrapping_insights,omitempty"`
	IMFInsights            string   `json:"imf_insights,omitempty"`
	MXFInsights            string   `json:"mxf_insights,omitempty"`
	DeadPixelInsights      string   `json:"dead_pixel_insights,omitempty"`
	PSEInsights            string   `json:"pse_insights,omitempty"`
	OverallQCScore         float64  `json:"overall_qc_score"`        // 0-100
	CriticalFindings       []string `json:"critical_findings,omitempty"`
	RecommendedActions     []string `json:"recommended_actions,omitempty"`
}

// RiskAssessment contains AI-generated risk analysis
type RiskAssessment struct {
	TechnicalRisk          string   `json:"technical_risk"`          // "low", "medium", "high"
	ComplianceRisk         string   `json:"compliance_risk"`         // "low", "medium", "high"
	OperationalRisk        string   `json:"operational_risk"`        // "low", "medium", "high"
	SafetyRisk             string   `json:"safety_risk"`             // "low", "medium", "high"
	OverallRiskLevel       string   `json:"overall_risk_level"`      // "low", "medium", "high", "critical"
	RiskFactors            []string `json:"risk_factors,omitempty"`
	MitigationStrategies   []string `json:"mitigation_strategies,omitempty"`
	MonitoringRecommendations []string `json:"monitoring_recommendations,omitempty"`
}

// IntegrationRecommendation contains workflow integration suggestions
type IntegrationRecommendation struct {
	Category           string  `json:"category"`              // "workflow", "technology", "process"
	Priority           string  `json:"priority"`              // "low", "medium", "high", "critical"
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	Implementation     string  `json:"implementation"`
	ExpectedBenefit    string  `json:"expected_benefit"`
	EstimatedCost      string  `json:"estimated_cost"`        // "low", "medium", "high"
	Complexity         string  `json:"complexity"`            // "simple", "moderate", "complex"
	Timeline           string  `json:"timeline"`              // "immediate", "short-term", "long-term"
}

type QualityIssue struct {
	Category    string   `json:"category"`    // "video", "audio", "container", "compliance"
	Severity    string   `json:"severity"`    // "critical", "major", "minor", "informational"
	Issue       string   `json:"issue"`       // Description of the issue
	Impact      string   `json:"impact"`      // Impact on quality/workflow
	Recommendation string `json:"recommendation"` // How to fix it
	FrameRange  string   `json:"frame_range,omitempty"` // If applicable
	StreamIndex int      `json:"stream_index,omitempty"` // If applicable
}

// WorkflowRecommendation represents workflow optimization suggestions
type WorkflowRecommendation struct {
	Category     string `json:"category"`     // "encoding", "delivery", "archival", "editing"
	Recommendation string `json:"recommendation"` // The recommendation
	Benefit      string `json:"benefit"`      // Expected benefit
	Complexity   string `json:"complexity"`   // "low", "medium", "high"
	Priority     string `json:"priority"`     // "high", "medium", "low"
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// GenerateEnhancedReport generates an AI-powered comprehensive analysis report
func (lla *LLMEnhancedAnalyzer) GenerateEnhancedReport(ctx context.Context, result *FFprobeResult) (*LLMEnhancedReport, error) {
	if !lla.enabled {
		return &LLMEnhancedReport{
			TechnicalSummary: "LLM analysis disabled - Ollama URL not configured",
			ProcessingTime:   0,
			Confidence:       0,
		}, nil
	}

	start := time.Now()
	
	// Build comprehensive analysis prompt
	prompt := lla.buildComprehensivePrompt(result)
	
	// Try primary model first, fallback if needed
	response, modelUsed, err := lla.generateWithFallback(ctx, prompt)
	if err != nil {
		lla.logger.Error().Err(err).Msg("Failed to generate LLM analysis")
		return &LLMEnhancedReport{
			TechnicalSummary: fmt.Sprintf("LLM analysis failed: %v", err),
			ProcessingTime:   time.Since(start),
			ModelUsed:        modelUsed,
			Confidence:       0,
		}, nil
	}

	// Parse the structured response
	report := lla.parseEnhancedResponse(response)
	
	// Generate advanced QC insights
	report.AdvancedQCInsights = lla.generateAdvancedQCInsights(result)
	
	// Generate risk assessment
	report.RiskAssessment = lla.generateRiskAssessment(result)
	
	// Generate integration recommendations
	report.IntegrationRecommendations = lla.generateIntegrationRecommendations(result)
	
	report.ProcessingTime = time.Since(start)
	report.ModelUsed = modelUsed
	report.Confidence = lla.calculateConfidence(result, response)

	lla.logger.Info().
		Dur("processing_time", report.ProcessingTime).
		Str("model_used", modelUsed).
		Float64("confidence", report.Confidence).
		Msg("LLM enhanced analysis completed")

	return report, nil
}

// buildComprehensivePrompt creates a detailed prompt for comprehensive analysis
func (lla *LLMEnhancedAnalyzer) buildComprehensivePrompt(result *FFprobeResult) string {
	// Convert analysis data to structured JSON
	analysisJSON, _ := json.MarshalIndent(result, "", "  ")

	prompt := fmt.Sprintf(`You are a senior video engineer with 15+ years of experience in broadcast, streaming, and post-production. Analyze this media file comprehensively and provide a detailed professional assessment.

CRITICAL: Respond ONLY with valid JSON in the exact structure specified below. Do not include any markdown, explanations, or text outside the JSON.

Required JSON Structure:
{
  "technical_summary": "Brief technical overview of the media file",
  "quality_assessment": "Detailed quality analysis covering video/audio quality, compression efficiency, artifacts",
  "compliance_analysis": "Broadcasting standards, streaming compatibility, format compliance",
  "optimization_recommendations": "Specific encoding/workflow improvements",
  "issue_identification": [
    {
      "category": "video|audio|container|compliance",
      "severity": "critical|major|minor|informational", 
      "issue": "Description of the issue",
      "impact": "Impact on quality/workflow",
      "recommendation": "How to fix it",
      "frame_range": "Optional: frame range if applicable",
      "stream_index": 0
    }
  ],
  "workflow_recommendations": [
    {
      "category": "encoding|delivery|archival|editing",
      "recommendation": "The recommendation",
      "benefit": "Expected benefit", 
      "complexity": "low|medium|high",
      "priority": "high|medium|low"
    }
  ],
  "executive_summary": "High-level summary for management/clients"
}

Media Analysis Data:
%s

Provide actionable insights focusing on practical implications for production workflows. Identify specific technical issues, quality concerns, and optimization opportunities.`, string(analysisJSON))

	return prompt
}

// generateWithFallback tries primary model, falls back to secondary if needed
func (lla *LLMEnhancedAnalyzer) generateWithFallback(ctx context.Context, prompt string) (string, string, error) {
	// Try primary model first
	response, err := lla.callOllama(ctx, prompt, lla.modelName)
	if err == nil {
		return response, lla.modelName, nil
	}

	lla.logger.Warn().
		Err(err).
		Str("primary_model", lla.modelName).
		Msg("Primary model failed, trying fallback")

	// Try fallback model
	if lla.fallbackModel != "" && lla.fallbackModel != lla.modelName {
		response, err = lla.callOllama(ctx, prompt, lla.fallbackModel)
		if err == nil {
			return response, lla.fallbackModel, nil
		}
	}

	return "", lla.modelName, fmt.Errorf("both primary and fallback models failed: %w", err)
}

// callOllama makes a request to the Ollama API
func (lla *LLMEnhancedAnalyzer) callOllama(ctx context.Context, prompt, model string) (string, error) {
	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"num_predict": 2000,
			"num_ctx":     4096,
			"temperature": 0.2,
			"num_thread":  4,
			"top_k":       40,
			"top_p":       0.9,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", lla.ollamaURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := lla.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return ollamaResp.Response, nil
}

// parseEnhancedResponse parses the LLM response into structured data
func (lla *LLMEnhancedAnalyzer) parseEnhancedResponse(response string) *LLMEnhancedReport {
	report := &LLMEnhancedReport{
		IssueIdentification:     []QualityIssue{},
		WorkflowRecommendations: []WorkflowRecommendation{},
	}

	// Try to parse as JSON first
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &jsonResponse); err == nil {
		// Successful JSON parsing
		if summary, ok := jsonResponse["technical_summary"].(string); ok {
			report.TechnicalSummary = summary
		}
		if quality, ok := jsonResponse["quality_assessment"].(string); ok {
			report.QualityAssessment = quality
		}
		if compliance, ok := jsonResponse["compliance_analysis"].(string); ok {
			report.ComplianceAnalysis = compliance
		}
		if optimization, ok := jsonResponse["optimization_recommendations"].(string); ok {
			report.OptimizationRecommendations = optimization
		}
		if executive, ok := jsonResponse["executive_summary"].(string); ok {
			report.ExecutiveSummary = executive
		}

		// Parse issues array
		if issues, ok := jsonResponse["issue_identification"].([]interface{}); ok {
			for _, issueData := range issues {
				if issueMap, ok := issueData.(map[string]interface{}); ok {
					issue := QualityIssue{}
					if category, ok := issueMap["category"].(string); ok {
						issue.Category = category
					}
					if severity, ok := issueMap["severity"].(string); ok {
						issue.Severity = severity
					}
					if issueDesc, ok := issueMap["issue"].(string); ok {
						issue.Issue = issueDesc
					}
					if impact, ok := issueMap["impact"].(string); ok {
						issue.Impact = impact
					}
					if recommendation, ok := issueMap["recommendation"].(string); ok {
						issue.Recommendation = recommendation
					}
					report.IssueIdentification = append(report.IssueIdentification, issue)
				}
			}
		}

		// Parse workflow recommendations array
		if workflows, ok := jsonResponse["workflow_recommendations"].([]interface{}); ok {
			for _, workflowData := range workflows {
				if workflowMap, ok := workflowData.(map[string]interface{}); ok {
					workflow := WorkflowRecommendation{}
					if category, ok := workflowMap["category"].(string); ok {
						workflow.Category = category
					}
					if recommendation, ok := workflowMap["recommendation"].(string); ok {
						workflow.Recommendation = recommendation
					}
					if benefit, ok := workflowMap["benefit"].(string); ok {
						workflow.Benefit = benefit
					}
					if complexity, ok := workflowMap["complexity"].(string); ok {
						workflow.Complexity = complexity
					}
					if priority, ok := workflowMap["priority"].(string); ok {
						workflow.Priority = priority
					}
					report.WorkflowRecommendations = append(report.WorkflowRecommendations, workflow)
				}
			}
		}
	} else {
		// Fallback: parse as free-form text
		lla.logger.Warn().Err(err).Msg("Failed to parse JSON response, using text parsing")
		report = lla.parseTextResponse(response)
	}

	return report
}

// parseTextResponse parses non-JSON response into structured format
func (lla *LLMEnhancedAnalyzer) parseTextResponse(response string) *LLMEnhancedReport {
	lines := strings.Split(response, "\n")
	
	report := &LLMEnhancedReport{
		TechnicalSummary:        "AI analysis completed (text format)",
		QualityAssessment:       "",
		ComplianceAnalysis:      "",
		OptimizationRecommendations: "",
		ExecutiveSummary:        "",
		IssueIdentification:     []QualityIssue{},
		WorkflowRecommendations: []WorkflowRecommendation{},
	}

	// Simple text parsing - extract key insights
	currentSection := ""
	content := []string{}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect section headers
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "technical") || strings.Contains(lowerLine, "summary") {
			currentSection = "technical"
		} else if strings.Contains(lowerLine, "quality") {
			currentSection = "quality"
		} else if strings.Contains(lowerLine, "compliance") {
			currentSection = "compliance"
		} else if strings.Contains(lowerLine, "optimization") || strings.Contains(lowerLine, "recommendation") {
			currentSection = "optimization"
		} else if strings.Contains(lowerLine, "executive") {
			currentSection = "executive"
		} else if strings.Contains(lowerLine, "issue") {
			currentSection = "issues"
		} else {
			content = append(content, line)
		}

		// Assign content to sections
		if len(content) > 0 && currentSection != "" {
			switch currentSection {
			case "technical":
				report.TechnicalSummary = strings.Join(content, " ")
			case "quality":
				report.QualityAssessment = strings.Join(content, " ")
			case "compliance":
				report.ComplianceAnalysis = strings.Join(content, " ")
			case "optimization":
				report.OptimizationRecommendations = strings.Join(content, " ")
			case "executive":
				report.ExecutiveSummary = strings.Join(content, " ")
			}
			content = []string{}
		}
	}

	// If no structured sections found, use entire response as technical summary
	if report.TechnicalSummary == "AI analysis completed (text format)" && len(response) > 100 {
		report.TechnicalSummary = response[:500] + "..."
	}

	return report
}

// calculateConfidence estimates the confidence in the analysis
func (lla *LLMEnhancedAnalyzer) calculateConfidence(result *FFprobeResult, response string) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence based on data availability
	if result.EnhancedAnalysis != nil {
		confidence += 0.2
	}
	if len(result.Streams) > 0 {
		confidence += 0.1
	}
	if result.Format != nil {
		confidence += 0.1
	}

	// Increase confidence based on response quality
	if len(response) > 500 {
		confidence += 0.1
	}
	
	// Check if response contains structured analysis
	if strings.Contains(response, "technical") && strings.Contains(response, "quality") {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// IsEnabled returns whether LLM enhancement is enabled
func (lla *LLMEnhancedAnalyzer) IsEnabled() bool {
	return lla.enabled
}

// HealthCheck verifies LLM service connectivity
func (lla *LLMEnhancedAnalyzer) HealthCheck(ctx context.Context) error {
	if !lla.enabled {
		return fmt.Errorf("LLM service disabled")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", lla.ollamaURL+"/api/version", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := lla.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("LLM service unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LLM service returned status %d", resp.StatusCode)
	}

	return nil
}

// generateAdvancedQCInsights generates insights from all advanced QC analysis results
func (lla *LLMEnhancedAnalyzer) generateAdvancedQCInsights(result *FFprobeResult) *AdvancedQCInsights {
	insights := &AdvancedQCInsights{
		OverallQCScore:     100.0,
		CriticalFindings:   []string{},
		RecommendedActions: []string{},
	}

	if result.EnhancedAnalysis == nil {
		insights.OverallQCScore = 50.0
		insights.CriticalFindings = append(insights.CriticalFindings, "No enhanced analysis available")
		return insights
	}

	// Analyze timecode results
	if result.EnhancedAnalysis.TimecodeAnalysis != nil {
		if result.EnhancedAnalysis.TimecodeAnalysis.HasTimecode {
			insights.TimecodeInsights = "Timecode present and properly formatted"
			if result.EnhancedAnalysis.TimecodeAnalysis.IsDropFrame {
				insights.TimecodeInsights += " (drop frame)"
			}
		} else {
			insights.TimecodeInsights = "No timecode detected - consider adding for professional workflows"
			insights.OverallQCScore -= 5.0
			insights.RecommendedActions = append(insights.RecommendedActions, "Add timecode track for professional editing workflows")
		}
	}

	// Analyze AFD results
	if result.EnhancedAnalysis.AFDAnalysis != nil {
		if result.EnhancedAnalysis.AFDAnalysis.HasAFD {
			insights.AFDInsights = "Active Format Description signaling present and compliant"
		} else {
			insights.AFDInsights = "No AFD signaling - may impact broadcast compliance"
			insights.OverallQCScore -= 3.0
			insights.RecommendedActions = append(insights.RecommendedActions, "Consider adding AFD signaling for broadcast compatibility")
		}
	}

	// Analyze transport stream results
	if result.EnhancedAnalysis.TransportStreamAnalysis != nil {
		if result.EnhancedAnalysis.TransportStreamAnalysis.IsMPEGTransportStream {
			insights.TransportStreamInsights = fmt.Sprintf("MPEG Transport Stream with %d PIDs", 
				result.EnhancedAnalysis.TransportStreamAnalysis.TotalPIDs)
			if len(result.EnhancedAnalysis.TransportStreamAnalysis.Errors) > 0 {
				insights.CriticalFindings = append(insights.CriticalFindings, "Transport stream errors detected")
				insights.OverallQCScore -= 10.0
			}
		} else {
			insights.TransportStreamInsights = "Not a transport stream format"
		}
	}

	// Analyze endianness results
	if result.EnhancedAnalysis.EndiannessAnalysis != nil {
		insights.EndiannessInsights = fmt.Sprintf("Container endianness: %s", 
			result.EnhancedAnalysis.EndiannessAnalysis.ContainerEndianness)
		if result.EnhancedAnalysis.EndiannessAnalysis.PlatformCompatibility != nil {
			if !result.EnhancedAnalysis.EndiannessAnalysis.PlatformCompatibility.CrossPlatformCompatible {
				insights.CriticalFindings = append(insights.CriticalFindings, "Platform compatibility issues detected")
				insights.OverallQCScore -= 8.0
			}
		}
	}

	// Analyze audio wrapping results
	if result.EnhancedAnalysis.AudioWrappingAnalysis != nil {
		professionalFormats := 0
		for _, stream := range result.EnhancedAnalysis.AudioWrappingAnalysis.AudioStreams {
			if stream.IsProfessionalFormat {
				professionalFormats++
			}
		}
		if professionalFormats > 0 {
			insights.AudioWrappingInsights = fmt.Sprintf("Professional audio wrapping detected in %d streams", professionalFormats)
		} else {
			insights.AudioWrappingInsights = "Standard audio wrapping - consider professional formats for broadcast"
			insights.OverallQCScore -= 2.0
		}
	}

	// Analyze IMF results
	if result.EnhancedAnalysis.IMFAnalysis != nil {
		if result.EnhancedAnalysis.IMFAnalysis.IsIMFCompliant {
			insights.IMFInsights = "IMF compliant package suitable for professional distribution"
		} else {
			insights.IMFInsights = "Not IMF compliant - may limit professional distribution options"
			insights.OverallQCScore -= 5.0
		}
	}

	// Analyze MXF results
	if result.EnhancedAnalysis.MXFAnalysis != nil {
		if result.EnhancedAnalysis.MXFAnalysis.IsMXFFile {
			insights.MXFInsights = fmt.Sprintf("MXF file with %s operational pattern", 
				result.EnhancedAnalysis.MXFAnalysis.MXFProfile)
			if result.EnhancedAnalysis.MXFAnalysis.ValidationResults != nil && 
				!result.EnhancedAnalysis.MXFAnalysis.ValidationResults.OverallCompliance {
				insights.CriticalFindings = append(insights.CriticalFindings, "MXF compliance issues detected")
				insights.OverallQCScore -= 12.0
			}
		} else {
			insights.MXFInsights = "Not an MXF file - consider MXF for professional workflows"
		}
	}

	// Analyze dead pixel results
	if result.EnhancedAnalysis.DeadPixelAnalysis != nil {
		totalDefects := result.EnhancedAnalysis.DeadPixelAnalysis.DeadPixelCount + 
			result.EnhancedAnalysis.DeadPixelAnalysis.StuckPixelCount + 
			result.EnhancedAnalysis.DeadPixelAnalysis.HotPixelCount
		if totalDefects > 0 {
			insights.DeadPixelInsights = fmt.Sprintf("%d pixel defects detected", totalDefects)
			if totalDefects > 10 {
				insights.CriticalFindings = append(insights.CriticalFindings, "Significant pixel defects detected")
				insights.OverallQCScore -= 15.0
			} else {
				insights.OverallQCScore -= float64(totalDefects)
			}
		} else {
			insights.DeadPixelInsights = "No pixel defects detected - clean video content"
		}
	}

	// Analyze PSE results
	if result.EnhancedAnalysis.PSEAnalysis != nil {
		insights.PSEInsights = fmt.Sprintf("PSE risk level: %s", result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel)
		if result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel == "high" || 
			result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel == "extreme" {
			insights.CriticalFindings = append(insights.CriticalFindings, "High photosensitive epilepsy risk detected")
			insights.OverallQCScore -= 20.0
			insights.RecommendedActions = append(insights.RecommendedActions, "Add PSE warning and consider content modification")
		} else if result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel == "medium" {
			insights.OverallQCScore -= 5.0
			insights.RecommendedActions = append(insights.RecommendedActions, "Consider PSE warning for sensitive viewers")
		}
	}

	// Ensure score doesn't go below 0
	if insights.OverallQCScore < 0 {
		insights.OverallQCScore = 0
	}

	return insights
}

// generateRiskAssessment generates a comprehensive risk assessment
func (lla *LLMEnhancedAnalyzer) generateRiskAssessment(result *FFprobeResult) *RiskAssessment {
	assessment := &RiskAssessment{
		TechnicalRisk:             "low",
		ComplianceRisk:            "low",
		OperationalRisk:           "low",
		SafetyRisk:                "low",
		OverallRiskLevel:          "low",
		RiskFactors:               []string{},
		MitigationStrategies:      []string{},
		MonitoringRecommendations: []string{},
	}

	if result.EnhancedAnalysis == nil {
		assessment.TechnicalRisk = "medium"
		assessment.OverallRiskLevel = "medium"
		assessment.RiskFactors = append(assessment.RiskFactors, "Limited analysis available")
		return assessment
	}

	// Assess technical risks
	technicalRiskScore := 0
	if result.EnhancedAnalysis.TransportStreamAnalysis != nil && len(result.EnhancedAnalysis.TransportStreamAnalysis.Errors) > 0 {
		technicalRiskScore += 2
		assessment.RiskFactors = append(assessment.RiskFactors, "Transport stream errors")
	}
	if result.EnhancedAnalysis.DeadPixelAnalysis != nil {
		totalDefects := result.EnhancedAnalysis.DeadPixelAnalysis.DeadPixelCount + 
			result.EnhancedAnalysis.DeadPixelAnalysis.StuckPixelCount
		if totalDefects > 10 {
			technicalRiskScore += 2
			assessment.RiskFactors = append(assessment.RiskFactors, "Significant pixel defects")
		}
	}

	// Assess compliance risks
	complianceRiskScore := 0
	if result.EnhancedAnalysis.IMFAnalysis != nil && !result.EnhancedAnalysis.IMFAnalysis.IsIMFCompliant {
		complianceRiskScore += 1
		assessment.RiskFactors = append(assessment.RiskFactors, "IMF non-compliance")
	}
	if result.EnhancedAnalysis.MXFAnalysis != nil && result.EnhancedAnalysis.MXFAnalysis.ValidationResults != nil &&
		!result.EnhancedAnalysis.MXFAnalysis.ValidationResults.OverallCompliance {
		complianceRiskScore += 2
		assessment.RiskFactors = append(assessment.RiskFactors, "MXF compliance issues")
	}

	// Assess safety risks
	safetyRiskScore := 0
	if result.EnhancedAnalysis.PSEAnalysis != nil {
		switch result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel {
		case "extreme":
			safetyRiskScore += 4
			assessment.RiskFactors = append(assessment.RiskFactors, "Extreme PSE risk")
		case "high":
			safetyRiskScore += 3
			assessment.RiskFactors = append(assessment.RiskFactors, "High PSE risk")
		case "medium":
			safetyRiskScore += 1
			assessment.RiskFactors = append(assessment.RiskFactors, "Medium PSE risk")
		}
	}

	// Calculate risk levels
	assessment.TechnicalRisk = lla.scoreToRiskLevel(technicalRiskScore)
	assessment.ComplianceRisk = lla.scoreToRiskLevel(complianceRiskScore)
	assessment.SafetyRisk = lla.scoreToRiskLevel(safetyRiskScore)

	// Calculate overall risk
	maxRiskScore := technicalRiskScore
	if complianceRiskScore > maxRiskScore {
		maxRiskScore = complianceRiskScore
	}
	if safetyRiskScore > maxRiskScore {
		maxRiskScore = safetyRiskScore
	}
	
	assessment.OverallRiskLevel = lla.scoreToRiskLevel(maxRiskScore)

	// Generate mitigation strategies
	if technicalRiskScore > 0 {
		assessment.MitigationStrategies = append(assessment.MitigationStrategies, 
			"Implement comprehensive QC testing before distribution")
	}
	if complianceRiskScore > 0 {
		assessment.MitigationStrategies = append(assessment.MitigationStrategies, 
			"Review content against relevant compliance standards")
	}
	if safetyRiskScore > 0 {
		assessment.MitigationStrategies = append(assessment.MitigationStrategies, 
			"Add appropriate viewer warnings and consider content modification")
	}

	// Generate monitoring recommendations
	assessment.MonitoringRecommendations = append(assessment.MonitoringRecommendations,
		"Regular automated QC analysis", "Compliance verification before distribution")

	return assessment
}

// generateIntegrationRecommendations generates workflow integration recommendations
func (lla *LLMEnhancedAnalyzer) generateIntegrationRecommendations(result *FFprobeResult) []IntegrationRecommendation {
	recommendations := []IntegrationRecommendation{}

	// Always recommend basic QC integration
	recommendations = append(recommendations, IntegrationRecommendation{
		Category:        "workflow",
		Priority:        "high",
		Title:           "Automated QC Integration",
		Description:     "Integrate comprehensive QC analysis into content workflow",
		Implementation:  "Add QC analysis as mandatory step before content approval",
		ExpectedBenefit: "Catch quality issues early, reduce distribution problems",
		EstimatedCost:   "low",
		Complexity:      "simple",
		Timeline:        "immediate",
	})

	if result.EnhancedAnalysis != nil {
		// PSE monitoring recommendation
		if result.EnhancedAnalysis.PSEAnalysis != nil && 
			result.EnhancedAnalysis.PSEAnalysis.PSERiskLevel != "safe" {
			recommendations = append(recommendations, IntegrationRecommendation{
				Category:        "technology",
				Priority:        "critical",
				Title:           "PSE Risk Monitoring",
				Description:     "Implement automated PSE risk detection in content pipeline",
				Implementation:  "Deploy PSE analysis before content approval and distribution",
				ExpectedBenefit: "Prevent potential viewer safety issues and regulatory violations",
				EstimatedCost:   "medium",
				Complexity:      "moderate",
				Timeline:        "short-term",
			})
		}

		// Dead pixel monitoring for acquisition
		if result.EnhancedAnalysis.DeadPixelAnalysis != nil && 
			(result.EnhancedAnalysis.DeadPixelAnalysis.DeadPixelCount > 0 || 
			 result.EnhancedAnalysis.DeadPixelAnalysis.StuckPixelCount > 0) {
			recommendations = append(recommendations, IntegrationRecommendation{
				Category:        "process",
				Priority:        "medium",
				Title:           "Camera/Sensor Quality Monitoring",
				Description:     "Regular monitoring of acquisition equipment for pixel defects",
				Implementation:  "Scheduled QC analysis of equipment outputs",
				ExpectedBenefit: "Early detection of equipment issues, improved content quality",
				EstimatedCost:   "low",
				Complexity:      "simple",
				Timeline:        "immediate",
			})
		}

		// Professional format recommendations
		if result.EnhancedAnalysis.MXFAnalysis != nil && result.EnhancedAnalysis.MXFAnalysis.IsMXFFile {
			recommendations = append(recommendations, IntegrationRecommendation{
				Category:        "workflow",
				Priority:        "medium",
				Title:           "Professional Format Validation",
				Description:     "Implement MXF compliance verification in professional workflows",
				Implementation:  "Add MXF validation checkpoints in production pipeline",
				ExpectedBenefit: "Ensure professional format compliance and interoperability",
				EstimatedCost:   "low",
				Complexity:      "simple",
				Timeline:        "short-term",
			})
		}
	}

	return recommendations
}

// scoreToRiskLevel converts numeric risk score to risk level
func (lla *LLMEnhancedAnalyzer) scoreToRiskLevel(score int) string {
	switch {
	case score == 0:
		return "low"
	case score <= 2:
		return "medium"
	case score <= 4:
		return "high"
	default:
		return "critical"
	}
}