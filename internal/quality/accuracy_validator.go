package quality

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AccuracyValidator validates the accuracy of quality metrics
type AccuracyValidator struct {
	analyzer   *QualityAnalyzer
	logger     zerolog.Logger
	testSuites []QualityTestSuite
}

// NewAccuracyValidator creates a new accuracy validator
func NewAccuracyValidator(analyzer *QualityAnalyzer, logger zerolog.Logger) *AccuracyValidator {
	return &AccuracyValidator{
		analyzer:   analyzer,
		logger:     logger,
		testSuites: loadDefaultTestSuites(),
	}
}

// ValidateMetricAccuracy validates the accuracy of a specific quality metric
func (av *AccuracyValidator) ValidateMetricAccuracy(ctx context.Context, metric QualityMetricType) (*AccuracyValidationResult, error) {
	av.logger.Info().
		Str("metric", string(metric)).
		Msg("Starting metric accuracy validation")

	result := &AccuracyValidationResult{
		ID:        uuid.New(),
		Metric:    metric,
		Status:    ValidationStatusProcessing,
		StartedAt: time.Now(),
		Tests:     make([]*AccuracyTestResult, 0),
	}

	// Get relevant test suites for this metric
	relevantSuites := av.getRelevantTestSuites(metric)
	if len(relevantSuites) == 0 {
		return nil, fmt.Errorf("no test suites available for metric %s", metric)
	}

	var totalTests, passedTests int
	var totalError, maxError float64

	// Run tests for each suite
	for _, suite := range relevantSuites {
		suiteResult, err := av.runTestSuite(ctx, suite, metric)
		if err != nil {
			av.logger.Error().
				Err(err).
				Str("suite", suite.Name).
				Msg("Test suite failed")
			continue
		}

		result.Tests = append(result.Tests, suiteResult.Tests...)
		totalTests += len(suiteResult.Tests)
		passedTests += suiteResult.PassedTests
		totalError += suiteResult.TotalError
		if suiteResult.MaxError > maxError {
			maxError = suiteResult.MaxError
		}
	}

	// Calculate overall metrics
	if totalTests > 0 {
		result.AccuracyScore = float64(passedTests) / float64(totalTests) * 100
		result.AverageError = totalError / float64(totalTests)
		result.MaxError = maxError
		result.PassedTests = passedTests
		result.TotalTests = totalTests
	}

	// Determine validation status
	result.Status = av.determineValidationStatus(result)
	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	// Generate recommendations
	result.Recommendations = av.generateRecommendations(result)

	av.logger.Info().
		Str("metric", string(metric)).
		Float64("accuracy_score", result.AccuracyScore).
		Int("passed_tests", passedTests).
		Int("total_tests", totalTests).
		Msg("Metric accuracy validation completed")

	return result, nil
}

// runTestSuite runs a specific test suite for a metric
func (av *AccuracyValidator) runTestSuite(ctx context.Context, suite QualityTestSuite, metric QualityMetricType) (*TestSuiteResult, error) {
	av.logger.Debug().
		Str("suite", suite.Name).
		Str("metric", string(metric)).
		Int("test_count", len(suite.TestCases)).
		Msg("Running test suite")

	result := &TestSuiteResult{
		SuiteName:   suite.Name,
		Tests:       make([]*AccuracyTestResult, 0, len(suite.TestCases)),
		PassedTests: 0,
		TotalError:  0,
		MaxError:    0,
	}

	for _, testCase := range suite.TestCases {
		testResult, err := av.runTestCase(ctx, testCase, metric)
		if err != nil {
			av.logger.Warn().
				Err(err).
				Str("test", testCase.Name).
				Msg("Test case failed to run")

			testResult = &AccuracyTestResult{
				TestName:       testCase.Name,
				Expected:       testCase.ExpectedScore,
				Actual:         0,
				Error:          math.Inf(1),
				Passed:         false,
				ErrorMessage:   err.Error(),
				ProcessingTime: 0,
			}
		}

		result.Tests = append(result.Tests, testResult)

		if testResult.Passed {
			result.PassedTests++
		}

		result.TotalError += math.Abs(testResult.Error)
		if math.Abs(testResult.Error) > result.MaxError {
			result.MaxError = math.Abs(testResult.Error)
		}
	}

	return result, nil
}

// runTestCase runs a single test case
func (av *AccuracyValidator) runTestCase(ctx context.Context, testCase QualityTestCase, metric QualityMetricType) (*AccuracyTestResult, error) {
	startTime := time.Now()

	// Create quality comparison request
	request := &QualityComparisonRequest{
		ReferenceFile: testCase.ReferenceFile,
		DistortedFile: testCase.DistortedFile,
		Metrics:       []QualityMetricType{metric},
		Configuration: testCase.Configuration,
	}

	// Run quality analysis
	qualityResult, err := av.analyzer.AnalyzeQuality(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("quality analysis failed: %w", err)
	}

	// Find the relevant analysis result
	var actualScore float64
	var found bool
	for _, analysis := range qualityResult.Analysis {
		if analysis.MetricType == metric {
			actualScore = analysis.OverallScore
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("metric %s not found in analysis results", metric)
	}

	// Calculate error and determine if test passed
	error := actualScore - testCase.ExpectedScore
	tolerance := testCase.Tolerance
	if tolerance == 0 {
		tolerance = getDefaultTolerance(metric)
	}

	passed := math.Abs(error) <= tolerance

	result := &AccuracyTestResult{
		TestName:       testCase.Name,
		Expected:       testCase.ExpectedScore,
		Actual:         actualScore,
		Error:          error,
		Tolerance:      tolerance,
		Passed:         passed,
		ProcessingTime: time.Since(startTime),
	}

	av.logger.Debug().
		Str("test", testCase.Name).
		Float64("expected", testCase.ExpectedScore).
		Float64("actual", actualScore).
		Float64("error", error).
		Float64("tolerance", tolerance).
		Bool("passed", passed).
		Msg("Test case completed")

	return result, nil
}

// getRelevantTestSuites returns test suites relevant to the given metric
func (av *AccuracyValidator) getRelevantTestSuites(metric QualityMetricType) []QualityTestSuite {
	var relevant []QualityTestSuite

	for _, suite := range av.testSuites {
		for _, supportedMetric := range suite.SupportedMetrics {
			if supportedMetric == metric {
				relevant = append(relevant, suite)
				break
			}
		}
	}

	return relevant
}

// determineValidationStatus determines the overall validation status
func (av *AccuracyValidator) determineValidationStatus(result *AccuracyValidationResult) ValidationStatus {
	if result.TotalTests == 0 {
		return ValidationStatusFailed
	}

	// Thresholds for validation status
	excellentThreshold := 95.0
	goodThreshold := 85.0
	acceptableThreshold := 70.0

	if result.AccuracyScore >= excellentThreshold {
		return ValidationStatusExcellent
	} else if result.AccuracyScore >= goodThreshold {
		return ValidationStatusGood
	} else if result.AccuracyScore >= acceptableThreshold {
		return ValidationStatusAcceptable
	} else {
		return ValidationStatusPoor
	}
}

// generateRecommendations generates recommendations based on validation results
func (av *AccuracyValidator) generateRecommendations(result *AccuracyValidationResult) []string {
	var recommendations []string

	switch result.Status {
	case ValidationStatusExcellent:
		recommendations = append(recommendations, "Metric accuracy is excellent. No immediate action required.")

	case ValidationStatusGood:
		recommendations = append(recommendations, "Metric accuracy is good. Consider monitoring for consistency.")
		if result.MaxError > getDefaultTolerance(result.Metric)*2 {
			recommendations = append(recommendations, "Some tests show higher than expected error. Review outlier cases.")
		}

	case ValidationStatusAcceptable:
		recommendations = append(recommendations, "Metric accuracy is acceptable but could be improved.")
		recommendations = append(recommendations, "Consider calibrating metric parameters or updating reference data.")
		if result.AverageError > getDefaultTolerance(result.Metric) {
			recommendations = append(recommendations, "Average error exceeds tolerance. Review calculation methodology.")
		}

	case ValidationStatusPoor:
		recommendations = append(recommendations, "Metric accuracy is poor. Immediate investigation required.")
		recommendations = append(recommendations, "Review metric implementation and test data quality.")
		recommendations = append(recommendations, "Consider updating metric algorithms or reference standards.")

	case ValidationStatusFailed:
		recommendations = append(recommendations, "Validation failed to complete. Check system configuration and test data availability.")
	}

	// Add metric-specific recommendations
	switch result.Metric {
	case MetricVMAF:
		if result.AccuracyScore < 90 {
			recommendations = append(recommendations, "For VMAF: Ensure model version matches reference implementation.")
			recommendations = append(recommendations, "For VMAF: Verify frame rate and resolution alignment.")
		}

	case MetricPSNR:
		if result.AccuracyScore < 95 {
			recommendations = append(recommendations, "For PSNR: Check bit depth and color space handling.")
			recommendations = append(recommendations, "For PSNR: Verify numerical precision in calculations.")
		}

	case MetricSSIM:
		if result.AccuracyScore < 90 {
			recommendations = append(recommendations, "For SSIM: Review window size and gaussian weights.")
			recommendations = append(recommendations, "For SSIM: Check luminance and contrast masking parameters.")
		}
	}

	return recommendations
}

// getDefaultTolerance returns default tolerance for a metric
func getDefaultTolerance(metric QualityMetricType) float64 {
	switch metric {
	case MetricVMAF:
		return 0.5 // VMAF tolerance: 0.5 points
	case MetricPSNR:
		return 0.1 // PSNR tolerance: 0.1 dB
	case MetricSSIM:
		return 0.005 // SSIM tolerance: 0.005
	case MetricMSE:
		return 1.0 // MSE tolerance: 1.0
	default:
		return 0.1
	}
}

// loadDefaultTestSuites loads built-in test suites
func loadDefaultTestSuites() []QualityTestSuite {
	return []QualityTestSuite{
		{
			Name:             "Basic VMAF Tests",
			Description:      "Basic VMAF accuracy validation tests",
			SupportedMetrics: []QualityMetricType{MetricVMAF},
			TestCases: []QualityTestCase{
				// These would be replaced with actual test files and expected scores
				{
					Name:          "VMAF_Test_1",
					Description:   "High quality reference vs slight degradation",
					ReferenceFile: "/test/reference/high_quality.mp4",
					DistortedFile: "/test/distorted/slight_degradation.mp4",
					ExpectedScore: 85.5,
					Tolerance:     0.5,
				},
				{
					Name:          "VMAF_Test_2",
					Description:   "High quality reference vs medium degradation",
					ReferenceFile: "/test/reference/high_quality.mp4",
					DistortedFile: "/test/distorted/medium_degradation.mp4",
					ExpectedScore: 65.2,
					Tolerance:     0.5,
				},
			},
		},
		{
			Name:             "Basic PSNR Tests",
			Description:      "Basic PSNR accuracy validation tests",
			SupportedMetrics: []QualityMetricType{MetricPSNR},
			TestCases: []QualityTestCase{
				{
					Name:          "PSNR_Test_1",
					Description:   "High quality reference vs slight degradation",
					ReferenceFile: "/test/reference/high_quality.mp4",
					DistortedFile: "/test/distorted/slight_degradation.mp4",
					ExpectedScore: 42.5,
					Tolerance:     0.1,
				},
			},
		},
		{
			Name:             "Basic SSIM Tests",
			Description:      "Basic SSIM accuracy validation tests",
			SupportedMetrics: []QualityMetricType{MetricSSIM},
			TestCases: []QualityTestCase{
				{
					Name:          "SSIM_Test_1",
					Description:   "High quality reference vs slight degradation",
					ReferenceFile: "/test/reference/high_quality.mp4",
					DistortedFile: "/test/distorted/slight_degradation.mp4",
					ExpectedScore: 0.95,
					Tolerance:     0.005,
				},
			},
		},
	}
}

// Validation types
type AccuracyValidationResult struct {
	ID              uuid.UUID             `json:"id"`
	Metric          QualityMetricType     `json:"metric"`
	Status          ValidationStatus      `json:"status"`
	AccuracyScore   float64               `json:"accuracy_score"`
	PassedTests     int                   `json:"passed_tests"`
	TotalTests      int                   `json:"total_tests"`
	AverageError    float64               `json:"average_error"`
	MaxError        float64               `json:"max_error"`
	Tests           []*AccuracyTestResult `json:"tests"`
	Recommendations []string              `json:"recommendations"`
	StartedAt       time.Time             `json:"started_at"`
	CompletedAt     time.Time             `json:"completed_at"`
	Duration        time.Duration         `json:"duration"`
}

type ValidationStatus string

const (
	ValidationStatusProcessing ValidationStatus = "processing"
	ValidationStatusExcellent  ValidationStatus = "excellent"
	ValidationStatusGood       ValidationStatus = "good"
	ValidationStatusAcceptable ValidationStatus = "acceptable"
	ValidationStatusPoor       ValidationStatus = "poor"
	ValidationStatusFailed     ValidationStatus = "failed"
)

type AccuracyTestResult struct {
	TestName       string        `json:"test_name"`
	Expected       float64       `json:"expected"`
	Actual         float64       `json:"actual"`
	Error          float64       `json:"error"`
	Tolerance      float64       `json:"tolerance"`
	Passed         bool          `json:"passed"`
	ErrorMessage   string        `json:"error_message,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
}

type TestSuiteResult struct {
	SuiteName   string                `json:"suite_name"`
	Tests       []*AccuracyTestResult `json:"tests"`
	PassedTests int                   `json:"passed_tests"`
	TotalError  float64               `json:"total_error"`
	MaxError    float64               `json:"max_error"`
}

type QualityTestSuite struct {
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	SupportedMetrics []QualityMetricType `json:"supported_metrics"`
	TestCases        []QualityTestCase   `json:"test_cases"`
}

type QualityTestCase struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	ReferenceFile string        `json:"reference_file"`
	DistortedFile string        `json:"distorted_file"`
	ExpectedScore float64       `json:"expected_score"`
	Tolerance     float64       `json:"tolerance"`
	Configuration QualityConfig `json:"configuration,omitempty"`
}
