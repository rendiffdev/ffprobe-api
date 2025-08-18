package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// executeFFprobeCommand executes an ffprobe command and returns the output
func executeFFprobeCommand(ctx context.Context, cmd []string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("empty command")
	}
	
	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w, output: %s", err, string(output))
	}
	
	return string(output), nil
}

// FFprobe represents the ffprobe service
type FFprobe struct {
	binaryPath       string
	logger           zerolog.Logger
	defaultTimeout   time.Duration
	maxOutputSize    int64
	enhancedAnalyzer *EnhancedAnalyzer
	enableContentAnalysis bool
}

// NewFFprobe creates a new FFprobe instance with binary validation
func NewFFprobe(binaryPath string, logger zerolog.Logger) *FFprobe {
	if binaryPath == "" {
		binaryPath = "ffprobe"
	}

	ffprobe := &FFprobe{
		binaryPath:            binaryPath,
		logger:                logger,
		defaultTimeout:        5 * time.Minute, // Default 5 minute timeout
		maxOutputSize:         100 * 1024 * 1024, // Default 100MB output limit
		enhancedAnalyzer:      NewEnhancedAnalyzer(binaryPath, logger),
		enableContentAnalysis: false, // Disabled by default for performance
	}

	return ffprobe
}

// ValidateBinaryAtStartup validates FFprobe binary is available and executable
// This should be called during application initialization
func (f *FFprobe) ValidateBinaryAtStartup(ctx context.Context) error {
	// First, try to resolve the binary path
	if f.binaryPath != "ffprobe" {
		// Check if absolute path exists
		if _, err := os.Stat(f.binaryPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("ffprobe binary not found at path: %s", f.binaryPath)
			}
			return fmt.Errorf("cannot access ffprobe binary at path %s: %w", f.binaryPath, err)
		}
	} else {
		// Check if ffprobe is in PATH
		if _, err := exec.LookPath("ffprobe"); err != nil {
			return fmt.Errorf("ffprobe binary not found in PATH. Please install FFmpeg or set FFPROBE_PATH environment variable: %w", err)
		}
	}

	// Test binary execution and get version
	version, err := f.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("ffprobe binary found but not executable or returned error: %w", err)
	}

	f.logger.Info().
		Str("binary_path", f.binaryPath).
		Str("version", version).
		Msg("FFprobe binary validated successfully")

	return nil
}

// SetDefaultTimeout sets the default timeout for ffprobe operations
func (f *FFprobe) SetDefaultTimeout(timeout time.Duration) {
	f.defaultTimeout = timeout
}

// SetMaxOutputSize sets the maximum output size limit
func (f *FFprobe) SetMaxOutputSize(size int64) {
	f.maxOutputSize = size
}

// EnableContentAnalysis enables content-based analysis using FFmpeg filters
func (f *FFprobe) EnableContentAnalysis() {
	f.enableContentAnalysis = true
	// Replace with content-enabled analyzer
	ffmpegPath := strings.Replace(f.binaryPath, "ffprobe", "ffmpeg", 1)
	f.enhancedAnalyzer = NewEnhancedAnalyzerWithContentAnalysis(ffmpegPath, f.binaryPath, f.logger)
}

// DisableContentAnalysis disables content-based analysis for performance
func (f *FFprobe) DisableContentAnalysis() {
	f.enableContentAnalysis = false
	f.enhancedAnalyzer = NewEnhancedAnalyzer(f.binaryPath, f.logger)
}

// SetLLMAnalyzer sets the LLM analyzer for AI-powered quality analysis
func (f *FFprobe) SetLLMAnalyzer(llmAnalyzer *LLMEnhancedAnalyzer) {
	if f.enhancedAnalyzer != nil {
		f.enhancedAnalyzer.SetLLMAnalyzer(llmAnalyzer)
	}
}

// Probe executes ffprobe with the given options
func (f *FFprobe) Probe(ctx context.Context, options *FFprobeOptions) (*FFprobeResult, error) {
	startTime := time.Now()

	// Validate options first
	if err := ValidateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Build command arguments
	args, err := f.buildArgs(options)
	if err != nil {
		return nil, fmt.Errorf("failed to build ffprobe arguments: %w", err)
	}

	// Apply timeout
	timeout := f.defaultTimeout
	if options.Timeout > 0 {
		timeout = options.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, f.binaryPath, args...)
	
	// Prepare stdout and stderr capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	f.logger.Debug().
		Str("command", f.binaryPath).
		Strs("args", args).
		Msg("Executing ffprobe command")

	// Execute command
	err = cmd.Run()
	executionTime := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	result := &FFprobeResult{
		Command:       append([]string{f.binaryPath}, args...),
		ExecutionTime: executionTime,
		Success:       err == nil,
		ExitCode:      exitCode,
		Output:        stdout.String(),
		StdErr:        stderr.String(),
	}

	// Check output size limit
	if options.MaxOutputSize > 0 && int64(len(result.Output)) > options.MaxOutputSize {
		return result, fmt.Errorf("output size %d exceeds limit %d", len(result.Output), options.MaxOutputSize)
	}
	if int64(len(result.Output)) > f.maxOutputSize {
		return result, fmt.Errorf("output size %d exceeds default limit %d", len(result.Output), f.maxOutputSize)
	}

	// Log execution details
	f.logger.Info().
		Dur("execution_time", executionTime).
		Int("exit_code", exitCode).
		Bool("success", result.Success).
		Int("output_size", len(result.Output)).
		Msg("FFprobe execution completed")

	if err != nil {
		f.logger.Error().
			Err(err).
			Str("stderr", result.StdErr).
			Msg("FFprobe execution failed")
		return result, fmt.Errorf("ffprobe execution failed: %w", err)
	}

	// Parse output based on format
	if err := f.parseOutput(result, options); err != nil {
		f.logger.Error().
			Err(err).
			Msg("Failed to parse ffprobe output")
		return result, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Validate parsed result
	if err := ValidateResult(result); err != nil {
		f.logger.Warn().
			Err(err).
			Msg("Result validation warning")
		// Don't fail on validation warnings, just log them
	}

	// Perform enhanced analysis
	if f.enableContentAnalysis {
		// Perform comprehensive content analysis with all advanced QC features
		if err := f.enhancedAnalyzer.AnalyzeResultWithContent(ctx, result, options.Input); err != nil {
			f.logger.Warn().
				Err(err).
				Msg("Enhanced content analysis failed")
			// Don't fail on enhanced analysis errors, just log them
		}
		
		// Also perform advanced QC analysis for professional features
		if err := f.enhancedAnalyzer.AnalyzeResultWithAdvancedQC(ctx, result, options.Input); err != nil {
			f.logger.Warn().
				Err(err).
				Msg("Advanced QC analysis failed")
			// Don't fail on advanced QC errors, just log them
		}
	} else {
		// Perform standard enhanced analysis with HDR detection
		if err := f.enhancedAnalyzer.AnalyzeResultWithHDR(ctx, result, options.Input); err != nil {
			f.logger.Warn().
				Err(err).
				Msg("Enhanced analysis with HDR failed")
			// Don't fail on enhanced analysis errors, just log them
		}
		
		// Also perform advanced QC analysis for professional features (always enabled)
		if err := f.enhancedAnalyzer.AnalyzeResultWithAdvancedQC(ctx, result, options.Input); err != nil {
			f.logger.Warn().
				Err(err).
				Msg("Advanced QC analysis failed")
			// Don't fail on advanced QC errors, just log them
		}
		
		// Perform LLM-enhanced analysis if available
		if err := f.enhancedAnalyzer.AnalyzeResultWithLLM(ctx, result, options.Input); err != nil {
			f.logger.Warn().
				Err(err).
				Msg("LLM enhanced analysis failed")
			// Don't fail on LLM analysis errors, just log them
		}
	}

	return result, nil
}

// ProbeFile is a convenience method for probing a single file with comprehensive analysis
func (f *FFprobe) ProbeFile(ctx context.Context, filePath string) (*FFprobeResult, error) {
	options := &FFprobeOptions{
		Input:           filePath,
		OutputFormat:    OutputJSON,
		ShowFormat:      true,
		ShowStreams:     true,
		ShowChapters:    true,
		ShowPrograms:    true,
		ShowPrivateData: true,
		ShowFrames:      true,           // Enable frame analysis for GOP structure
		CountFrames:     true,
		CountPackets:    true,
		ProbeSize:       50 * 1024 * 1024, // 50MB probe size for detailed analysis
		AnalyzeDuration: 10 * 1000000,     // 10 seconds analysis duration
		PrettyPrint:     true,
		HideBanner:      true,
		ReadIntervals:   "0%+#100",      // Analyze first 100 frames for GOP analysis
	}

	return f.Probe(ctx, options)
}

// ProbeFileWithOptions probes a file with custom options
func (f *FFprobe) ProbeFileWithOptions(ctx context.Context, filePath string, options *FFprobeOptions) (*FFprobeResult, error) {
	if options == nil {
		options = &FFprobeOptions{}
	}
	
	options.Input = filePath
	if options.OutputFormat == "" {
		options.OutputFormat = OutputJSON
	}
	if !options.HideBanner {
		options.HideBanner = true // Hide banner by default for cleaner output
	}

	return f.Probe(ctx, options)
}

// ProbeFileWithContentAnalysis probes a file with enhanced content analysis enabled
func (f *FFprobe) ProbeFileWithContentAnalysis(ctx context.Context, filePath string) (*FFprobeResult, error) {
	// Temporarily enable content analysis
	wasEnabled := f.enableContentAnalysis
	f.EnableContentAnalysis()
	defer func() {
		if !wasEnabled {
			f.DisableContentAnalysis()
		}
	}()

	return f.ProbeFile(ctx, filePath)
}

// buildArgs constructs the command line arguments for ffprobe
func (f *FFprobe) buildArgs(options *FFprobeOptions) ([]string, error) {
	var args []string

	// Hide banner (should be first)
	if options.HideBanner {
		args = append(args, "-hide_banner")
	}

	// Log level
	if options.LogLevel != "" {
		args = append(args, "-loglevel", string(options.LogLevel))
	}

	// Report
	if options.Report {
		args = append(args, "-report")
	}

	// Input format
	if options.Format != "" {
		args = append(args, "-f", options.Format)
	}

	// Input options
	for key, value := range options.InputOptions {
		args = append(args, fmt.Sprintf("-%s", key), value)
	}

	// Probe size
	if options.ProbeSize > 0 {
		args = append(args, "-probesize", strconv.FormatInt(options.ProbeSize, 10))
	}

	// Analyze duration
	if options.AnalyzeDuration > 0 {
		args = append(args, "-analyzeduration", strconv.FormatInt(options.AnalyzeDuration, 10))
	}

	// Output format
	if options.OutputFormat != "" && options.OutputFormat != OutputDefault {
		args = append(args, "-of", string(options.OutputFormat))
		
		// Pretty print for JSON
		if options.PrettyPrint && options.OutputFormat == OutputJSON {
			args = append(args, "-pretty")
		}
	}

	// Show options
	if options.ShowFormat {
		args = append(args, "-show_format")
	}
	if options.ShowStreams {
		args = append(args, "-show_streams")
	}
	if options.ShowPackets {
		args = append(args, "-show_packets")
	}
	if options.ShowFrames {
		args = append(args, "-show_frames")
	}
	if options.ShowChapters {
		args = append(args, "-show_chapters")
	}
	if options.ShowPrograms {
		args = append(args, "-show_programs")
	}
	if options.ShowError {
		args = append(args, "-show_error")
	}
	if options.ShowData {
		args = append(args, "-show_data")
	}
	if options.ShowPrivateData {
		args = append(args, "-show_private_data")
	}

	// Selection options
	if options.SelectStreams != "" {
		args = append(args, "-select_streams", options.SelectStreams)
	}
	if options.ReadIntervals != "" {
		args = append(args, "-read_intervals", options.ReadIntervals)
	}
	if options.ShowEntries != "" {
		args = append(args, "-show_entries", options.ShowEntries)
	}

	// Count options
	if options.CountFrames {
		args = append(args, "-count_frames")
	}
	if options.CountPackets {
		args = append(args, "-count_packets")
	}

	// Input file (must be last)
	if options.Input == "" {
		return nil, fmt.Errorf("input file is required")
	}
	args = append(args, "-i", options.Input)

	return args, nil
}

// parseOutput parses the ffprobe output based on the output format
func (f *FFprobe) parseOutput(result *FFprobeResult, options *FFprobeOptions) error {
	if result.Output == "" {
		return nil
	}

	// Only parse JSON output for structured data
	if options.OutputFormat == OutputJSON {
		return f.parseJSONOutput(result)
	}

	// For other formats, output is kept as raw string
	return nil
}

// parseJSONOutput parses JSON output from ffprobe
func (f *FFprobe) parseJSONOutput(result *FFprobeResult) error {
	var data map[string]interface{}
	
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON output: %w", err)
	}

	// Parse format information
	if formatData, ok := data["format"].(map[string]interface{}); ok {
		formatJSON, _ := json.Marshal(formatData)
		var format FormatInfo
		if err := json.Unmarshal(formatJSON, &format); err == nil {
			result.Format = &format
		}
	}

	// Parse streams information
	if streamsData, ok := data["streams"].([]interface{}); ok {
		for _, streamData := range streamsData {
			streamJSON, _ := json.Marshal(streamData)
			var stream StreamInfo
			if err := json.Unmarshal(streamJSON, &stream); err == nil {
				result.Streams = append(result.Streams, stream)
			}
		}
	}

	// Parse packets information
	if packetsData, ok := data["packets"].([]interface{}); ok {
		for _, packetData := range packetsData {
			packetJSON, _ := json.Marshal(packetData)
			var packet PacketInfo
			if err := json.Unmarshal(packetJSON, &packet); err == nil {
				result.Packets = append(result.Packets, packet)
			}
		}
	}

	// Parse frames information
	if framesData, ok := data["frames"].([]interface{}); ok {
		for _, frameData := range framesData {
			frameJSON, _ := json.Marshal(frameData)
			var frame FrameInfo
			if err := json.Unmarshal(frameJSON, &frame); err == nil {
				result.Frames = append(result.Frames, frame)
			}
		}
	}

	// Parse chapters information
	if chaptersData, ok := data["chapters"].([]interface{}); ok {
		for _, chapterData := range chaptersData {
			chapterJSON, _ := json.Marshal(chapterData)
			var chapter ChapterInfo
			if err := json.Unmarshal(chapterJSON, &chapter); err == nil {
				result.Chapters = append(result.Chapters, chapter)
			}
		}
	}

	// Parse programs information
	if programsData, ok := data["programs"].([]interface{}); ok {
		for _, programData := range programsData {
			programJSON, _ := json.Marshal(programData)
			var program ProgramInfo
			if err := json.Unmarshal(programJSON, &program); err == nil {
				result.Programs = append(result.Programs, program)
			}
		}
	}

	// Parse error information
	if errorData, ok := data["error"].(map[string]interface{}); ok {
		errorJSON, _ := json.Marshal(errorData)
		var errorInfo ErrorInfo
		if err := json.Unmarshal(errorJSON, &errorInfo); err == nil {
			result.Error = &errorInfo
		}
	}

	return nil
}

// ValidateInput checks if the input file exists and is accessible
func (f *FFprobe) ValidateInput(input string) error {
	// Check if it's a URL (starts with protocol)
	if strings.Contains(input, "://") {
		return nil // Assume URLs are valid for now
	}

	// Check if local file exists
	if _, err := os.Stat(input); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", input)
		}
		return fmt.Errorf("cannot access input file: %w", err)
	}

	return nil
}

// GetVersion returns the ffprobe version
func (f *FFprobe) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, f.binaryPath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ffprobe version: %w", err)
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ffprobe version") {
			return strings.TrimSpace(line), nil
		}
	}

	return string(output), nil
}

// CheckBinary verifies that ffprobe binary is available and executable
func (f *FFprobe) CheckBinary(ctx context.Context) error {
	version, err := f.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("ffprobe binary not found or not executable: %w", err)
	}

	f.logger.Info().Str("version", version).Msg("FFprobe binary verified")
	return nil
}

// ProbeWithProgress probes a file with progress reporting for large files
func (f *FFprobe) ProbeWithProgress(ctx context.Context, options *FFprobeOptions, progressCallback func(float64)) (*FFprobeResult, error) {
	// This is a simplified implementation
	// For real progress reporting, you'd need to parse ffprobe's stderr output
	// and extract progress information
	
	if progressCallback != nil {
		progressCallback(0.0)
	}

	result, err := f.Probe(ctx, options)

	if progressCallback != nil {
		if err != nil {
			progressCallback(0.0) // Reset on error
		} else {
			progressCallback(1.0) // Complete
		}
	}

	return result, err
}

// StreamFile provides streaming analysis for very large files
func (f *FFprobe) StreamFile(ctx context.Context, filePath string, chunkCallback func(chunk string) error) error {
	options := &FFprobeOptions{
		Input:        filePath,
		OutputFormat: OutputJSON,
		ShowStreams:  true,
		ShowFormat:   true,
		HideBanner:   true,
	}

	args, err := f.buildArgs(options)
	if err != nil {
		return fmt.Errorf("failed to build ffprobe arguments: %w", err)
	}

	cmd := exec.CommandContext(ctx, f.binaryPath, args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffprobe: %w", err)
	}

	// Read output in chunks
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := chunkCallback(scanner.Text()); err != nil {
			cmd.Process.Kill()
			return fmt.Errorf("chunk callback error: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ffprobe output: %w", err)
	}

	return cmd.Wait()
}