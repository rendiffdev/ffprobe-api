package ffmpeg

import "time"

// IMF (Interoperable Master Format) analysis constants and thresholds.
// These constants are based on SMPTE 2067 standards and industry best practices
// for IMF package validation and compliance checking.

// SMPTE 2067 Standard Versions
const (
	// SMPTE2067_2_Version is the version of SMPTE 2067-2 (Core Constraints) implemented
	SMPTE2067_2_Version = "2020"
	
	// SMPTE2067_3_Version is the version of SMPTE 2067-3 (Text Profile) implemented
	SMPTE2067_3_Version = "2020"
	
	// SMPTE2067_5_Version is the version of SMPTE 2067-5 (Audio Essence) implemented
	SMPTE2067_5_Version = "2020"
	
	// SMPTE2067_20_Version is the version of SMPTE 2067-20 (Application #2) implemented
	SMPTE2067_20_Version = "2016"
	
	// SMPTE2067_21_Version is the version of SMPTE 2067-21 (Application #2E) implemented
	SMPTE2067_21_Version = "2020"
)

// IMF File and Directory Naming Constraints
const (
	// MaxAssetFilenameLength is the maximum allowed filename length for IMF assets
	MaxAssetFilenameLength = 255
	
	// MaxDirectoryNameLength is the maximum allowed directory name length
	MaxDirectoryNameLength = 255
	
	// MaxPathLength is the maximum allowed full path length
	MaxPathLength = 260
	
	// UUIDLength is the standard length for UUID strings
	UUIDLength = 36
	
	// UUIDPattern is the regex pattern for valid UUIDs
	UUIDPattern = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
)

// IMF Package Structure Requirements
const (
	// RequiredCPLCount is the minimum number of CPL files required
	RequiredCPLCount = 1
	
	// RequiredPKLCount is the minimum number of PKL files required  
	RequiredPKLCount = 1
	
	// RequiredAssetMapCount is the minimum number of Asset Map files required
	RequiredAssetMapCount = 1
	
	// MaxCPLCount is the maximum recommended number of CPL files
	MaxCPLCount = 10
	
	// MaxConcurrentValidations is the maximum number of concurrent IMF validations
	MaxConcurrentValidations = 3
)

// IMF Validation Timeouts
const (
	// DefaultIMFAnalysisTimeout is the default timeout for IMF analysis operations
	DefaultIMFAnalysisTimeout = 15 * time.Minute
	
	// CPLValidationTimeout is the timeout for CPL validation
	CPLValidationTimeout = 2 * time.Minute
	
	// PKLValidationTimeout is the timeout for PKL validation  
	PKLValidationTimeout = 5 * time.Minute
	
	// AssetMapValidationTimeout is the timeout for Asset Map validation
	AssetMapValidationTimeout = 1 * time.Minute
	
	// TrackFileValidationTimeout is the timeout for individual track file validation
	TrackFileValidationTimeout = 10 * time.Minute
)

// Video Constraints (SMPTE 2067-20/21)
const (
	// IMF_VideoCodec_JPEG2000 is the required video codec for IMF
	IMF_VideoCodec_JPEG2000 = "jpeg2000"
	
	// IMF_VideoProfile_Cinema2K is the 2K cinema profile
	IMF_VideoProfile_Cinema2K = "Cinema2K"
	
	// IMF_VideoProfile_Cinema4K is the 4K cinema profile  
	IMF_VideoProfile_Cinema4K = "Cinema4K"
	
	// IMF_VideoProfile_HDTV is the HDTV profile
	IMF_VideoProfile_HDTV = "HDTV"
	
	// IMF_VideoProfile_UHDTV is the UHDTV profile
	IMF_VideoProfile_UHDTV = "UHDTV"
)

// Supported Video Resolutions for IMF
const (
	// IMF_Resolution_2K_Flat is 2K flat (1998x1080)
	IMF_Resolution_2K_Flat = "1998x1080"
	
	// IMF_Resolution_2K_Scope is 2K scope (2048x858)  
	IMF_Resolution_2K_Scope = "2048x858"
	
	// IMF_Resolution_2K_Full is 2K full container (2048x1080)
	IMF_Resolution_2K_Full = "2048x1080"
	
	// IMF_Resolution_4K_Flat is 4K flat (3996x2160)
	IMF_Resolution_4K_Flat = "3996x2160"
	
	// IMF_Resolution_4K_Scope is 4K scope (4096x1716)
	IMF_Resolution_4K_Scope = "4096x1716"
	
	// IMF_Resolution_4K_Full is 4K full container (4096x2160)
	IMF_Resolution_4K_Full = "4096x2160"
	
	// IMF_Resolution_HD is HD (1920x1080)
	IMF_Resolution_HD = "1920x1080"
	
	// IMF_Resolution_UHD is UHD (3840x2160)
	IMF_Resolution_UHD = "3840x2160"
)

// Supported Frame Rates for IMF
const (
	// IMF_FrameRate_24 is 24 fps
	IMF_FrameRate_24 = "24/1"
	
	// IMF_FrameRate_25 is 25 fps
	IMF_FrameRate_25 = "25/1"
	
	// IMF_FrameRate_30 is 30 fps
	IMF_FrameRate_30 = "30/1"
	
	// IMF_FrameRate_48 is 48 fps
	IMF_FrameRate_48 = "48/1"
	
	// IMF_FrameRate_50 is 50 fps
	IMF_FrameRate_50 = "50/1"
	
	// IMF_FrameRate_60 is 60 fps
	IMF_FrameRate_60 = "60/1"
	
	// IMF_FrameRate_23_976 is 23.976 fps
	IMF_FrameRate_23_976 = "24000/1001"
	
	// IMF_FrameRate_29_97 is 29.97 fps
	IMF_FrameRate_29_97 = "30000/1001"
	
	// IMF_FrameRate_59_94 is 59.94 fps
	IMF_FrameRate_59_94 = "60000/1001"
)

// Audio Constraints (SMPTE 2067-5)
const (
	// IMF_AudioCodec_PCM is the required audio codec for IMF
	IMF_AudioCodec_PCM = "pcm_s24le"
	
	// IMF_AudioSampleRate_48kHz is the standard sample rate
	IMF_AudioSampleRate_48kHz = 48000
	
	// IMF_AudioSampleRate_96kHz is the high quality sample rate
	IMF_AudioSampleRate_96kHz = 96000
	
	// IMF_AudioBitDepth_16 is 16-bit audio
	IMF_AudioBitDepth_16 = 16
	
	// IMF_AudioBitDepth_24 is 24-bit audio (recommended)
	IMF_AudioBitDepth_24 = 24
	
	// IMF_AudioBitDepth_32 is 32-bit audio
	IMF_AudioBitDepth_32 = 32
)

// Supported Audio Channel Configurations
const (
	// IMF_Audio_Mono is mono audio (1.0)
	IMF_Audio_Mono = "1.0"
	
	// IMF_Audio_Stereo is stereo audio (2.0)
	IMF_Audio_Stereo = "2.0"
	
	// IMF_Audio_5_1 is 5.1 surround audio
	IMF_Audio_5_1 = "5.1"
	
	// IMF_Audio_7_1 is 7.1 surround audio
	IMF_Audio_7_1 = "7.1"
	
	// IMF_Audio_Atmos is Dolby Atmos object-based audio
	IMF_Audio_Atmos = "Atmos"
)

// Subtitle Constraints (SMPTE 2067-3)
const (
	// IMF_SubtitleFormat_IMSC1 is IMSC1 subtitle format
	IMF_SubtitleFormat_IMSC1 = "IMSC1"
	
	// IMF_SubtitleFormat_SMPTE_TT is SMPTE Timed Text
	IMF_SubtitleFormat_SMPTE_TT = "SMPTE-TT"
	
	// IMF_SubtitleCodec_XML is XML-based subtitle codec
	IMF_SubtitleCodec_XML = "xml"
)

// Color Space and HDR Constraints
const (
	// IMF_ColorPrimaries_BT709 is Rec. 709 color primaries
	IMF_ColorPrimaries_BT709 = "bt709"
	
	// IMF_ColorPrimaries_BT2020 is Rec. 2020 color primaries
	IMF_ColorPrimaries_BT2020 = "bt2020"
	
	// IMF_ColorPrimaries_P3D65 is P3-D65 color primaries
	IMF_ColorPrimaries_P3D65 = "smpte432"
	
	// IMF_TransferCharacteristic_BT709 is Rec. 709 transfer characteristic
	IMF_TransferCharacteristic_BT709 = "bt709"
	
	// IMF_TransferCharacteristic_SMPTE2084 is SMPTE 2084 (PQ) transfer
	IMF_TransferCharacteristic_SMPTE2084 = "smpte2084"
	
	// IMF_TransferCharacteristic_HLG is HLG transfer characteristic
	IMF_TransferCharacteristic_HLG = "arib-std-b67"
	
	// IMF_MatrixCoefficients_BT709 is Rec. 709 matrix coefficients
	IMF_MatrixCoefficients_BT709 = "bt709"
	
	// IMF_MatrixCoefficients_BT2020 is Rec. 2020 matrix coefficients
	IMF_MatrixCoefficients_BT2020 = "bt2020nc"
)

// HDR Constraints
const (
	// IMF_MaxLuminance_SDR is maximum luminance for SDR content (100 nits)
	IMF_MaxLuminance_SDR = 100.0
	
	// IMF_MaxLuminance_HDR10 is maximum luminance for HDR10 content (10000 nits)
	IMF_MaxLuminance_HDR10 = 10000.0
	
	// IMF_MaxLuminance_HLG is maximum luminance for HLG content (1000 nits)
	IMF_MaxLuminance_HLG = 1000.0
	
	// IMF_MinLuminance_HDR is minimum luminance for HDR content (0.01 nits)
	IMF_MinLuminance_HDR = 0.01
	
	// IMF_MaxContentLightLevel_Default is default max content light level
	IMF_MaxContentLightLevel_Default = 1000
	
	// IMF_MaxFrameAverageLightLevel_Default is default max frame average light level
	IMF_MaxFrameAverageLightLevel_Default = 400
)

// Netflix IMF Profile Constants
const (
	// Netflix_VideoCodec is Netflix required video codec
	Netflix_VideoCodec = "jpeg2000"
	
	// Netflix_AudioCodec is Netflix required audio codec
	Netflix_AudioCodec = "pcm_s24le"
	
	// Netflix_SubtitleFormat is Netflix required subtitle format
	Netflix_SubtitleFormat = "IMSC1"
	
	// Netflix_MinBitrate is Netflix minimum video bitrate (Mbps)
	Netflix_MinBitrate = 25
	
	// Netflix_MaxBitrate is Netflix maximum video bitrate (Mbps)
	Netflix_MaxBitrate = 300
	
	// Netflix_TargetLoudness is Netflix target loudness level (LUFS)
	Netflix_TargetLoudness = -27.0
	
	// Netflix_LoudnessTolerance is Netflix loudness tolerance (+/- LUFS)
	Netflix_LoudnessTolerance = 2.0
	
	// Netflix_TruePeakLimit is Netflix true peak limit (dBTP)
	Netflix_TruePeakLimit = -2.0
)

// Validation Severity Levels
const (
	// Validation_Error indicates a critical compliance failure
	Validation_Error = "error"
	
	// Validation_Warning indicates a non-critical compliance issue
	Validation_Warning = "warning"
	
	// Validation_Info indicates informational validation result
	Validation_Info = "info"
)

// Hash Algorithm Support
const (
	// Hash_SHA1 is SHA-1 hash algorithm (deprecated)
	Hash_SHA1 = "SHA-1"
	
	// Hash_SHA256 is SHA-256 hash algorithm (recommended)
	Hash_SHA256 = "SHA-256"
	
	// Hash_SHA512 is SHA-512 hash algorithm
	Hash_SHA512 = "SHA-512"
	
	// Hash_MD5 is MD5 hash algorithm (not recommended)
	Hash_MD5 = "MD5"
)

// File Extensions
const (
	// CPL_Extension is the file extension for CPL files
	CPL_Extension = ".xml"
	
	// PKL_Extension is the file extension for PKL files
	PKL_Extension = ".xml"
	
	// AssetMap_Extension is the file extension for Asset Map files
	AssetMap_Extension = ".xml"
	
	// MXF_Extension is the file extension for MXF files
	MXF_Extension = ".mxf"
)

// XML Namespace URIs
const (
	// CPL_Namespace_2016 is the CPL namespace for 2016 spec
	CPL_Namespace_2016 = "http://www.smpte-ra.org/schemas/2067-3/2016"
	
	// PKL_Namespace_2016 is the PKL namespace for 2016 spec
	PKL_Namespace_2016 = "http://www.smpte-ra.org/schemas/2067-2/2016"
	
	// AssetMap_Namespace is the Asset Map namespace
	AssetMap_Namespace = "http://www.smpte-ra.org/schemas/433/2008/dcmlTypes/"
)

// Analysis Quality Settings
const (
	// IMF_FastAnalysisTimeout is timeout for fast IMF analysis
	IMF_FastAnalysisTimeout = 5 * time.Minute
	
	// IMF_StandardAnalysisTimeout is timeout for standard IMF analysis
	IMF_StandardAnalysisTimeout = 15 * time.Minute
	
	// IMF_DeepAnalysisTimeout is timeout for comprehensive IMF analysis
	IMF_DeepAnalysisTimeout = 30 * time.Minute
	
	// MaxConcurrentTrackAnalysis is max concurrent track file analyses
	MaxConcurrentTrackAnalysis = 4
	
	// MemoryLimitPerAnalysis is estimated memory limit per IMF analysis (MB)
	MemoryLimitPerAnalysis = 1024
)

// Supported IMF Application Profiles
var (
	// SupportedIMFProfiles lists all supported IMF application profiles
	SupportedIMFProfiles = []string{
		"SMPTE 2067-20:2016",  // Application #2
		"SMPTE 2067-21:2020",  // Application #2E Extended
		"SMPTE 2067-40:2020",  // Application #4 (HDTV)
		"SMPTE 2067-50:2017",  // Application #5 (UHDTV)
		"Netflix IMF Profile", // Netflix-specific profile
		"DPP IMF Profile",     // Digital Production Partnership profile
	}
	
	// RequiredIMFFiles lists the minimum required files in an IMF package
	RequiredIMFFiles = []string{
		"ASSETMAP.xml",
		"PKL*.xml",
		"CPL*.xml",
	}
	
	// IllegalFilenameCharacters lists characters not allowed in IMF filenames
	IllegalFilenameCharacters = []string{
		"<", ">", ":", "\"", "|", "?", "*",
		"/", "\\", // Path separators
	}
)

// Compliance Level Definitions
const (
	// Compliance_Full indicates full compliance with specifications
	Compliance_Full = "full"
	
	// Compliance_Partial indicates partial compliance with minor issues
	Compliance_Partial = "partial"
	
	// Compliance_NonCompliant indicates significant compliance failures
	Compliance_NonCompliant = "non_compliant"
	
	// Compliance_Unknown indicates compliance could not be determined
	Compliance_Unknown = "unknown"
)