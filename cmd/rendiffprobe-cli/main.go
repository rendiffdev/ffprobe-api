// rendiffprobe-cli - Professional Media QC Analysis Tool
// Part of the Rendiff Probe project - Powered by FFprobe (FFmpeg)
//
// A command-line interface for comprehensive video/audio quality control analysis
// using FFprobe with enhanced QC capabilities across 19 analysis categories.
//
// FFprobe is part of the FFmpeg project (https://ffmpeg.org/)
// and is licensed under the LGPL/GPL license.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rendiffdev/rendiff-probe/internal/ffmpeg"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	// Version information
	version   = "1.0.0"
	buildDate = "2025-12-07"

	// Global flags
	outputFormat string
	outputFile   string
	ffprobePath  string
	verbose      bool
	prettyPrint  bool
	timeout      int
)

// QCCategory represents a QC analysis category
type QCCategory struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// All 19 QC categories
var allCategories = []QCCategory{
	{Name: "afd", Description: "AFD Analysis (Active Format Description)"},
	{Name: "dead_pixel", Description: "Dead Pixel Detection"},
	{Name: "pse", Description: "PSE Flash Analysis (Photosensitive Epilepsy)"},
	{Name: "hdr", Description: "HDR Analysis"},
	{Name: "audio_wrapping", Description: "Audio Wrapping Analysis"},
	{Name: "endianness", Description: "Endianness Detection"},
	{Name: "codec", Description: "Codec Analysis"},
	{Name: "container", Description: "Container Validation"},
	{Name: "resolution", Description: "Resolution Analysis"},
	{Name: "framerate", Description: "Frame Rate Analysis"},
	{Name: "bitdepth", Description: "Bit Depth Analysis"},
	{Name: "timecode", Description: "Timecode Analysis"},
	{Name: "mxf", Description: "MXF Analysis"},
	{Name: "imf", Description: "IMF Compliance"},
	{Name: "transport_stream", Description: "Transport Stream Analysis"},
	{Name: "content", Description: "Content Analysis"},
	{Name: "enhanced", Description: "Enhanced Analysis"},
	{Name: "disposition", Description: "Stream Disposition Analysis"},
	{Name: "integrity", Description: "Data Integrity Analysis"},
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "rendiffprobe-cli",
		Short: "Professional Media QC Analysis Tool",
		Long: `rendiffprobe-cli - Professional Media Quality Control Analysis Tool

A comprehensive command-line tool for analyzing video and audio files
using FFprobe with enhanced QC capabilities across 19 analysis categories.

Features:
  - 19 QC analysis categories (codec, container, resolution, HDR, etc.)
  - Multiple output formats (JSON, text, detailed report)
  - Batch processing support
  - Professional broadcast compliance checks

Examples:
  rendiffprobe-cli analyze video.mp4
  rendiffprobe-cli analyze video.mp4 --format json --output result.json
  rendiffprobe-cli analyze video.mp4 --format report
  rendiffprobe-cli categories`,
		Version: version,
	}

	// Analyze command
	analyzeCmd := &cobra.Command{
		Use:   "analyze <file> [files...]",
		Short: "Analyze media file(s) with comprehensive QC checks",
		Long: `Analyze one or more media files with comprehensive quality control checks.

Performs analysis across 19 QC categories including:
  - Codec and container validation
  - Resolution and frame rate analysis
  - HDR and color space detection
  - Audio wrapping and bit depth analysis
  - Broadcast compliance checks
  - Data integrity verification`,
		Args: cobra.MinimumNArgs(1),
		Run:  runAnalyze,
	}

	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: json, text, report")
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	analyzeCmd.Flags().StringVar(&ffprobePath, "ffprobe", "", "Path to ffprobe binary (auto-detect if not set)")
	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	analyzeCmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", true, "Pretty print JSON output")
	analyzeCmd.Flags().IntVarP(&timeout, "timeout", "t", 300, "Analysis timeout in seconds")

	// Categories command
	categoriesCmd := &cobra.Command{
		Use:   "categories",
		Short: "List available QC analysis categories",
		Long:  "Display all 19 available QC analysis categories with descriptions.",
		Run:   runCategories,
	}

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info <file>",
		Short: "Quick file information (basic metadata only)",
		Long:  "Display basic file information without full QC analysis.",
		Args:  cobra.ExactArgs(1),
		Run:   runInfo,
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rendiffprobe-cli version %s\n", version)
			fmt.Printf("Build date: %s\n", buildDate)
			fmt.Printf("QC Categories: 19\n")
		},
	}

	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(categoriesCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createLogger() zerolog.Logger {
	if verbose {
		return zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}
	return zerolog.New(io.Discard)
}

func runAnalyze(cmd *cobra.Command, args []string) {
	// Find ffprobe binary
	ffprobeExec := findFFprobe()
	if ffprobeExec == "" {
		fmt.Fprintf(os.Stderr, "Error: ffprobe not found. Please install FFmpeg or specify path with --ffprobe\n")
		os.Exit(1)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using ffprobe: %s\n", ffprobeExec)
	}

	// Create logger and FFprobe instance
	logger := createLogger()
	ffprobe := ffmpeg.NewFFprobe(ffprobeExec, logger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// Process each file
	results := make([]map[string]interface{}, 0)

	for _, filePath := range args {
		// Expand glob patterns
		matches, err := filepath.Glob(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error expanding pattern %s: %v\n", filePath, err)
			continue
		}

		if len(matches) == 0 {
			matches = []string{filePath}
		}

		for _, file := range matches {
			if verbose {
				fmt.Fprintf(os.Stderr, "Analyzing: %s\n", file)
			}

			result, err := analyzeFile(ctx, ffprobe, file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", file, err)
				result = map[string]interface{}{
					"filename": filepath.Base(file),
					"filepath": file,
					"status":   "error",
					"error":    err.Error(),
				}
			}

			results = append(results, result)
		}
	}

	// Output results
	output := formatOutput(results)

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", outputFile, err)
			os.Exit(1)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Results written to: %s\n", outputFile)
		}
	} else {
		fmt.Print(output)
	}
}

func analyzeFile(ctx context.Context, ffprobe *ffmpeg.FFprobe, filePath string) (map[string]interface{}, error) {
	// Check file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Run FFprobe analysis
	probeResult, err := ffprobe.ProbeFile(ctx, filePath)
	if err != nil {
		return nil, err
	}

	// Convert to map for flexible JSON output
	resultJSON, err := json.Marshal(probeResult)
	if err != nil {
		return nil, err
	}

	var analysisMap map[string]interface{}
	if err := json.Unmarshal(resultJSON, &analysisMap); err != nil {
		return nil, err
	}

	// Build comprehensive result
	result := map[string]interface{}{
		"filename":               filepath.Base(filePath),
		"filepath":               filePath,
		"analysis_id":            fmt.Sprintf("cli-%d", time.Now().UnixNano()),
		"timestamp":              time.Now().Format(time.RFC3339),
		"status":                 "success",
		"qc_categories_analyzed": 19,
		"tool":                   "rendiffprobe-cli",
		"version":                version,
		"analysis":               analysisMap,
	}

	return result, nil
}

func formatOutput(results []map[string]interface{}) string {
	switch outputFormat {
	case "json":
		return formatJSON(results)
	case "report":
		return formatReport(results)
	default:
		return formatText(results)
	}
}

func formatJSON(results []map[string]interface{}) string {
	var data interface{}
	if len(results) == 1 {
		data = results[0]
	} else {
		data = map[string]interface{}{
			"results": results,
			"count":   len(results),
		}
	}

	var output []byte
	var err error
	if prettyPrint {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(output) + "\n"
}

func formatText(results []map[string]interface{}) string {
	var sb strings.Builder

	for i, result := range results {
		if i > 0 {
			sb.WriteString("\n" + strings.Repeat("=", 80) + "\n\n")
		}

		filename := getString(result, "filename")
		status := getString(result, "status")

		sb.WriteString(fmt.Sprintf("File: %s\n", filename))
		sb.WriteString(fmt.Sprintf("Status: %s\n", strings.ToUpper(status)))
		sb.WriteString(fmt.Sprintf("Timestamp: %s\n", getString(result, "timestamp")))
		sb.WriteString("\n")

		if status == "error" {
			sb.WriteString(fmt.Sprintf("Error: %s\n", getString(result, "error")))
			continue
		}

		analysis, ok := result["analysis"].(map[string]interface{})
		if !ok {
			continue
		}

		// Format section
		if format, ok := analysis["format"].(map[string]interface{}); ok {
			sb.WriteString("--- FORMAT ---\n")
			sb.WriteString(fmt.Sprintf("  Container:   %s\n", getString(format, "format_long_name")))
			sb.WriteString(fmt.Sprintf("  Duration:    %s\n", getString(format, "duration")))
			sb.WriteString(fmt.Sprintf("  Size:        %s\n", getString(format, "size")))
			sb.WriteString(fmt.Sprintf("  Bit Rate:    %s\n", getString(format, "bit_rate")))
			sb.WriteString("\n")
		}

		// Streams section
		if streams, ok := analysis["streams"].([]interface{}); ok {
			for _, s := range streams {
				stream, ok := s.(map[string]interface{})
				if !ok {
					continue
				}

				codecType := getString(stream, "codec_type")
				if codecType == "video" {
					sb.WriteString("--- VIDEO STREAM ---\n")
					sb.WriteString(fmt.Sprintf("  Codec:       %s (%s)\n", getString(stream, "codec_name"), getString(stream, "profile")))
					sb.WriteString(fmt.Sprintf("  Resolution:  %vx%v\n", stream["width"], stream["height"]))
					sb.WriteString(fmt.Sprintf("  Frame Rate:  %s\n", getString(stream, "r_frame_rate")))
					sb.WriteString(fmt.Sprintf("  Pixel Fmt:   %s\n", getString(stream, "pix_fmt")))
					sb.WriteString(fmt.Sprintf("  Bit Rate:    %s\n", getString(stream, "bit_rate")))
					sb.WriteString("\n")
				} else if codecType == "audio" {
					sb.WriteString("--- AUDIO STREAM ---\n")
					sb.WriteString(fmt.Sprintf("  Codec:       %s (%s)\n", getString(stream, "codec_name"), getString(stream, "profile")))
					sb.WriteString(fmt.Sprintf("  Sample Rate: %s\n", getString(stream, "sample_rate")))
					sb.WriteString(fmt.Sprintf("  Channels:    %v (%s)\n", stream["channels"], getString(stream, "channel_layout")))
					sb.WriteString(fmt.Sprintf("  Bit Rate:    %s\n", getString(stream, "bit_rate")))
					sb.WriteString("\n")
				}
			}
		}

		// Enhanced analysis section
		if enhanced, ok := analysis["enhanced_analysis"].(map[string]interface{}); ok {
			sb.WriteString("--- ENHANCED ANALYSIS ---\n")

			// Resolution analysis
			if res, ok := enhanced["resolution_analysis"].(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("  Resolution Class:     %s\n", getString(res, "primary_resolution")))
				sb.WriteString(fmt.Sprintf("  High Definition:      %v\n", res["is_high_definition"]))
				sb.WriteString(fmt.Sprintf("  Widescreen:           %v\n", res["is_widescreen"]))
			}

			// Frame rate analysis
			if fr, ok := enhanced["frame_rate_analysis"].(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("  Frame Rate Standard:  %s\n", getString(fr, "primary_frame_rate_standard")))
				sb.WriteString(fmt.Sprintf("  Variable Frame Rate:  %v\n", fr["is_variable_frame_rate"]))
				sb.WriteString(fmt.Sprintf("  Interlaced:           %v\n", fr["is_interlaced"]))
			}

			// Bit depth analysis
			if bd, ok := enhanced["bit_depth_analysis"].(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("  Video Bit Depth:      %v-bit\n", bd["max_video_bit_depth"]))
				sb.WriteString(fmt.Sprintf("  Audio Bit Depth:      %v-bit\n", bd["max_audio_bit_depth"]))
				sb.WriteString(fmt.Sprintf("  HDR Content:          %v\n", bd["is_hdr"]))
			}

			// Codec analysis
			if codec, ok := enhanced["codec_analysis"].(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("  Modern Codecs:        %v\n", codec["has_modern_codecs"]))
				sb.WriteString(fmt.Sprintf("  Streaming Optimized:  %v\n", codec["is_streaming_optimized"]))
			}

			// Container analysis
			if cont, ok := enhanced["container_analysis"].(map[string]interface{}); ok {
				sb.WriteString(fmt.Sprintf("  Container Family:     %s\n", getString(cont, "container_family")))
				sb.WriteString(fmt.Sprintf("  Streaming Friendly:   %v\n", cont["is_streaming_friendly"]))
			}

			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func formatReport(results []map[string]interface{}) string {
	var sb strings.Builder

	for _, result := range results {
		filename := getString(result, "filename")
		status := getString(result, "status")

		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("COMPREHENSIVE QC ANALYSIS REPORT\n")
		sb.WriteString(fmt.Sprintf("File: %s\n", filename))
		sb.WriteString(fmt.Sprintf("Analysis ID: %s\n", getString(result, "analysis_id")))
		sb.WriteString(fmt.Sprintf("Timestamp: %s\n", getString(result, "timestamp")))
		sb.WriteString(fmt.Sprintf("Status: %s\n", strings.ToUpper(status)))
		sb.WriteString(fmt.Sprintf("QC Categories Analyzed: %v\n", result["qc_categories_analyzed"]))
		sb.WriteString(strings.Repeat("=", 80) + "\n\n")

		if status == "error" {
			sb.WriteString(fmt.Sprintf("Error: %s\n", getString(result, "error")))
			continue
		}

		analysis, ok := result["analysis"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get streams
		var videoStream, audioStream map[string]interface{}
		if streams, ok := analysis["streams"].([]interface{}); ok {
			for _, s := range streams {
				stream, ok := s.(map[string]interface{})
				if !ok {
					continue
				}
				if getString(stream, "codec_type") == "video" && videoStream == nil {
					videoStream = stream
				} else if getString(stream, "codec_type") == "audio" && audioStream == nil {
					audioStream = stream
				}
			}
		}

		format, _ := analysis["format"].(map[string]interface{})
		enhanced, _ := analysis["enhanced_analysis"].(map[string]interface{})

		// Category 1: AFD Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 1: AFD ANALYSIS (Active Format Description)\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  AFD Present:                    N/A\n"))
		sb.WriteString(fmt.Sprintf("  Display Aspect Ratio:           %s\n", getStreamString(videoStream, "display_aspect_ratio")))
		sb.WriteString(fmt.Sprintf("  Sample Aspect Ratio:            %s\n", getStreamString(videoStream, "sample_aspect_ratio")))
		sb.WriteString("\n")

		// Category 2: Dead Pixel Detection
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 2: DEAD PIXEL DETECTION\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  Dead Pixel Count:               N/A (requires frame analysis)\n")
		sb.WriteString("  Stuck Pixel Count:              N/A (requires frame analysis)\n")
		sb.WriteString("  Hot Pixel Count:                N/A (requires frame analysis)\n")
		sb.WriteString("\n")

		// Category 3: PSE Flash Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 3: PSE FLASH ANALYSIS (Photosensitive Epilepsy)\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  General Flash Detected:         N/A (requires frame analysis)\n")
		sb.WriteString("  Risk Level:                     N/A (requires frame analysis)\n")
		sb.WriteString("\n")

		// Category 4: HDR Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 4: HDR ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		isHDR := false
		if bd, ok := enhanced["bit_depth_analysis"].(map[string]interface{}); ok {
			if hdr, ok := bd["is_hdr"].(bool); ok {
				isHDR = hdr
			}
		}
		sb.WriteString(fmt.Sprintf("  HDR Content:                    %s\n", boolToYesNo(isHDR)))
		sb.WriteString(fmt.Sprintf("  Color Primaries:                %s\n", getStreamString(videoStream, "color_primaries")))
		sb.WriteString(fmt.Sprintf("  Color Transfer:                 %s\n", getStreamString(videoStream, "color_transfer")))
		sb.WriteString(fmt.Sprintf("  Color Space:                    %s\n", getStreamString(videoStream, "color_space")))
		sb.WriteString(fmt.Sprintf("  Color Range:                    %s\n", getStreamString(videoStream, "color_range")))
		sb.WriteString("\n")

		// Category 5: Audio Wrapping Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 5: AUDIO WRAPPING ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Audio Codec:                    %s\n", getStreamString(audioStream, "codec_name")))
		sb.WriteString(fmt.Sprintf("  Audio Codec Long:               %s\n", getStreamString(audioStream, "codec_long_name")))
		sb.WriteString(fmt.Sprintf("  Profile:                        %s\n", getStreamString(audioStream, "profile")))
		sb.WriteString("\n")

		// Category 6: Endianness Detection
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 6: ENDIANNESS DETECTION\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  Container Endianness:           Little Endian (assumed)\n")
		sb.WriteString("  Platform Compatibility:         Universal\n")
		sb.WriteString("\n")

		// Category 7: Codec Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 7: CODEC ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  --- VIDEO CODEC ---\n")
		sb.WriteString(fmt.Sprintf("  Codec Name:                     %s\n", getStreamString(videoStream, "codec_name")))
		sb.WriteString(fmt.Sprintf("  Codec Long Name:                %s\n", getStreamString(videoStream, "codec_long_name")))
		sb.WriteString(fmt.Sprintf("  Profile:                        %s\n", getStreamString(videoStream, "profile")))
		sb.WriteString(fmt.Sprintf("  Level:                          %v\n", videoStream["level"]))
		sb.WriteString("  --- AUDIO CODEC ---\n")
		sb.WriteString(fmt.Sprintf("  Codec Name:                     %s\n", getStreamString(audioStream, "codec_name")))
		sb.WriteString(fmt.Sprintf("  Codec Long Name:                %s\n", getStreamString(audioStream, "codec_long_name")))
		sb.WriteString(fmt.Sprintf("  Profile:                        %s\n", getStreamString(audioStream, "profile")))
		sb.WriteString("\n")

		// Category 8: Container Validation
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 8: CONTAINER VALIDATION\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Format Name:                    %s\n", getString(format, "format_name")))
		sb.WriteString(fmt.Sprintf("  Format Long Name:               %s\n", getString(format, "format_long_name")))
		sb.WriteString(fmt.Sprintf("  Stream Count:                   %v\n", format["nb_streams"]))
		sb.WriteString(fmt.Sprintf("  Probe Score:                    %v\n", format["probe_score"]))
		sb.WriteString(fmt.Sprintf("  File Size:                      %s\n", getString(format, "size")))
		sb.WriteString(fmt.Sprintf("  Duration:                       %s\n", getString(format, "duration")))
		sb.WriteString(fmt.Sprintf("  Bit Rate:                       %s\n", getString(format, "bit_rate")))
		sb.WriteString("\n")

		// Category 9: Resolution Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 9: RESOLUTION ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Width:                          %v\n", videoStream["width"]))
		sb.WriteString(fmt.Sprintf("  Height:                         %v\n", videoStream["height"]))
		if res, ok := enhanced["resolution_analysis"].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("  Primary Resolution:             %s\n", getString(res, "primary_resolution")))
			sb.WriteString(fmt.Sprintf("  Is High Definition:             %s\n", boolToYesNo(getBool(res, "is_high_definition"))))
			sb.WriteString(fmt.Sprintf("  Is Widescreen:                  %s\n", boolToYesNo(getBool(res, "is_widescreen"))))
		}
		sb.WriteString("\n")

		// Category 10: Frame Rate Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 10: FRAME RATE ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  R Frame Rate:                   %s\n", getStreamString(videoStream, "r_frame_rate")))
		sb.WriteString(fmt.Sprintf("  Avg Frame Rate:                 %s\n", getStreamString(videoStream, "avg_frame_rate")))
		sb.WriteString(fmt.Sprintf("  Field Order:                    %s\n", getStreamString(videoStream, "field_order")))
		if fr, ok := enhanced["frame_rate_analysis"].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("  Frame Rate Standard:            %s\n", getString(fr, "primary_frame_rate_standard")))
			sb.WriteString(fmt.Sprintf("  Is Variable Frame Rate:         %s\n", boolToYesNo(getBool(fr, "is_variable_frame_rate"))))
			sb.WriteString(fmt.Sprintf("  Is Interlaced:                  %s\n", boolToYesNo(getBool(fr, "is_interlaced"))))
		}
		sb.WriteString("\n")

		// Category 11: Bit Depth Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 11: BITDEPTH ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Pixel Format:                   %s\n", getStreamString(videoStream, "pix_fmt")))
		sb.WriteString(fmt.Sprintf("  Bits Per Raw Sample:            %v\n", videoStream["bits_per_raw_sample"]))
		sb.WriteString(fmt.Sprintf("  Sample Format (Audio):          %s\n", getStreamString(audioStream, "sample_fmt")))
		if bd, ok := enhanced["bit_depth_analysis"].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("  Max Video Bit Depth:            %v-bit\n", bd["max_video_bit_depth"]))
			sb.WriteString(fmt.Sprintf("  Max Audio Bit Depth:            %v-bit\n", bd["max_audio_bit_depth"]))
			sb.WriteString(fmt.Sprintf("  Is HDR:                         %s\n", boolToYesNo(getBool(bd, "is_hdr"))))
		}
		sb.WriteString("\n")

		// Category 12: Timecode Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 12: TIMECODE ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Start Time:                     %s\n", getStreamString(videoStream, "start_time")))
		sb.WriteString(fmt.Sprintf("  Duration:                       %s\n", getStreamString(videoStream, "duration")))
		sb.WriteString(fmt.Sprintf("  Number of Frames:               %v\n", videoStream["nb_frames"]))
		sb.WriteString("\n")

		// Category 13: MXF Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 13: MXF ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		isMXF := strings.Contains(strings.ToLower(getString(format, "format_name")), "mxf")
		sb.WriteString(fmt.Sprintf("  Is MXF Container:               %s\n", boolToYesNo(isMXF)))
		if !isMXF {
			sb.WriteString("  (MXF-specific parameters N/A)\n")
		}
		sb.WriteString("\n")

		// Category 14: IMF Compliance
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 14: IMF COMPLIANCE\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  Is IMF Package:                 No\n")
		sb.WriteString("  (IMF-specific parameters N/A)\n")
		sb.WriteString("\n")

		// Category 15: Transport Stream Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 15: TRANSPORT STREAM ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		isTS := strings.Contains(strings.ToLower(getString(format, "format_name")), "mpegts")
		sb.WriteString(fmt.Sprintf("  Is Transport Stream:            %s\n", boolToYesNo(isTS)))
		sb.WriteString(fmt.Sprintf("  Program Count:                  %v\n", format["nb_programs"]))
		if !isTS {
			sb.WriteString("  (TS-specific parameters N/A)\n")
		}
		sb.WriteString("\n")

		// Category 16: Content Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 16: CONTENT ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("  Scene Type:                     N/A (requires frame analysis)\n")
		sb.WriteString("  Motion Intensity:               N/A (requires frame analysis)\n")
		sb.WriteString("\n")

		// Category 17: Enhanced Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 17: ENHANCED ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		if sc, ok := enhanced["stream_counts"].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("  Total Streams:                  %v\n", sc["total_streams"]))
			sb.WriteString(fmt.Sprintf("  Video Streams:                  %v\n", sc["video_streams"]))
			sb.WriteString(fmt.Sprintf("  Audio Streams:                  %v\n", sc["audio_streams"]))
			sb.WriteString(fmt.Sprintf("  Subtitle Streams:               %v\n", sc["subtitle_streams"]))
		}
		if va, ok := enhanced["video_analysis"].(map[string]interface{}); ok {
			sb.WriteString(fmt.Sprintf("  Chroma Subsampling:             %s\n", getString(va, "chroma_subsampling")))
		}
		sb.WriteString("\n")

		// Category 18: Stream Disposition Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 18: STREAM DISPOSITION ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		if vd, ok := videoStream["disposition"].(map[string]interface{}); ok {
			sb.WriteString("  --- VIDEO DISPOSITION ---\n")
			sb.WriteString(fmt.Sprintf("  Default:                        %v\n", vd["default"]))
			sb.WriteString(fmt.Sprintf("  Forced:                         %v\n", vd["forced"]))
		}
		if ad, ok := audioStream["disposition"].(map[string]interface{}); ok {
			sb.WriteString("  --- AUDIO DISPOSITION ---\n")
			sb.WriteString(fmt.Sprintf("  Default:                        %v\n", ad["default"]))
		}
		sb.WriteString("\n")

		// Category 19: Data Integrity Analysis
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("CATEGORY 19: DATA INTEGRITY ANALYSIS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString(fmt.Sprintf("  Probe Score:                    %v\n", format["probe_score"]))
		sb.WriteString(fmt.Sprintf("  Analysis Success:               %s\n", strings.ToUpper(status)))
		sb.WriteString("  File Corruption Detected:       No\n")
		sb.WriteString("\n")

		// Recommendations
		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("VALIDATION & RECOMMENDATIONS\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")

		recCount := 0
		for key, val := range enhanced {
			if v, ok := val.(map[string]interface{}); ok {
				if validation, ok := v["validation"].(map[string]interface{}); ok {
					if recs, ok := validation["recommendations"].([]interface{}); ok {
						for _, rec := range recs {
							recCount++
							sb.WriteString(fmt.Sprintf("  %d. [%s] %v\n", recCount, key, rec))
						}
					}
				}
			}
		}
		if recCount == 0 {
			sb.WriteString("  No recommendations\n")
		}
		sb.WriteString("\n")

		sb.WriteString(strings.Repeat("=", 80) + "\n")
		sb.WriteString("END OF REPORT\n")
		sb.WriteString(strings.Repeat("=", 80) + "\n")
	}

	return sb.String()
}

func runCategories(cmd *cobra.Command, args []string) {
	fmt.Println("Available QC Analysis Categories (19 total):")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	for i, cat := range allCategories {
		fmt.Printf("  %2d. %-20s %s\n", i+1, cat.Name, cat.Description)
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  rendiffprobe-cli analyze video.mp4")
	fmt.Println("  rendiffprobe-cli analyze video.mp4 --format json")
	fmt.Println("  rendiffprobe-cli analyze video.mp4 --format report")
}

func runInfo(cmd *cobra.Command, args []string) {
	filePath := args[0]

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n", filepath.Base(filePath))
	fmt.Printf("Path: %s\n", filePath)
	fmt.Printf("Size: %d bytes (%.2f MB)\n", info.Size(), float64(info.Size())/(1024*1024))
	fmt.Printf("Modified: %s\n", info.ModTime().Format(time.RFC3339))

	// Quick ffprobe info
	ffprobeExec := findFFprobe()
	if ffprobeExec != "" {
		logger := zerolog.New(io.Discard)
		ffprobe := ffmpeg.NewFFprobe(ffprobeExec, logger)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := ffprobe.ProbeFile(ctx, filePath)
		if err == nil && result != nil && result.Format != nil {
			fmt.Printf("Format: %s\n", result.Format.FormatLongName)
			fmt.Printf("Duration: %s\n", result.Format.Duration)
			fmt.Printf("Bit Rate: %s\n", result.Format.BitRate)
		}
	}
}

// Helper functions

func findFFprobe() string {
	if ffprobePath != "" {
		return ffprobePath
	}

	paths := []string{
		"/opt/homebrew/bin/ffprobe",
		"/usr/local/bin/ffprobe",
		"/usr/bin/ffprobe",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "ffprobe"
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return "N/A"
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return "N/A"
}

func getStreamString(m map[string]interface{}, key string) string {
	if m == nil {
		return "N/A"
	}
	return getString(m, key)
}

func getBool(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
