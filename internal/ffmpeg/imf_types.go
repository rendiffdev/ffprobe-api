package ffmpeg

// IMF (Interoperable Master Format) analysis types and structures.
// This file contains all data structures used for IMF compliance validation
// in accordance with SMPTE 2067 standards and Netflix delivery specifications.

// IMFAnalysis contains comprehensive IMF compliance analysis results.
// This is the main container for all IMF-related analysis data, providing detailed
// assessment of Interoperable Master Format packages for broadcast delivery.
//
// The analysis follows industry standards:
//   - SMPTE 2067: IMF core constraints and application profiles
//   - Netflix IMF Profile: Netflix-specific delivery requirements
//   - DPP IMF Profile: Digital Production Partnership specifications
//   - EIDR Registry: Entertainment ID Registry compliance
type IMFAnalysis struct {
	IsIMFCompliant        bool                   `json:"is_imf_compliant"`
	IMFProfile            string                 `json:"imf_profile,omitempty"`
	CPLAnalysis           *CPLAnalysis           `json:"cpl_analysis,omitempty"`
	PKLAnalysis           *PKLAnalysis           `json:"pkl_analysis,omitempty"`
	AssetMapAnalysis      *AssetMapAnalysis      `json:"asset_map_analysis,omitempty"`
	TrackFileAnalysis     []TrackFileAnalysis    `json:"track_file_analysis,omitempty"`
	NetflixCompliance     *NetflixIMFCompliance  `json:"netflix_compliance,omitempty"`
	SMPTE2067Compliance   *SMPTE2067Compliance   `json:"smpte_2067_compliance,omitempty"`
	ValidationResults     *IMFValidationResults  `json:"validation_results,omitempty"`
	RecommendedActions    []string               `json:"recommended_actions,omitempty"`
}

// CPLAnalysis contains Composition Playlist analysis for IMF packages.
// The CPL is the core document that defines the timeline and structure of IMF content.
type CPLAnalysis struct {
	CPLExists             bool                 `json:"cpl_exists"`
	CPLID                 string               `json:"cpl_id,omitempty"`
	CPLTitle              string               `json:"cpl_title,omitempty"`
	EditRate              string               `json:"edit_rate,omitempty"`
	Duration              string               `json:"duration,omitempty"`
	SegmentCount          int                  `json:"segment_count"`
	VirtualTrackCount     int                  `json:"virtual_track_count"`
	AudioTrackCount       int                  `json:"audio_track_count"`
	VideoTrackCount       int                  `json:"video_track_count"`
	SubtitleTrackCount    int                  `json:"subtitle_track_count"`
	IMFSegments           []IMFSegment         `json:"imf_segments,omitempty"`
}

// PKLAnalysis contains Packing List analysis for IMF packages.
// The PKL provides cryptographic hashes for asset integrity verification.
type PKLAnalysis struct {
	PKLExists             bool                 `json:"pkl_exists"`
	PKLID                 string               `json:"pkl_id,omitempty"`
	AssetCount            int                  `json:"asset_count"`
	HashAlgorithm         string               `json:"hash_algorithm,omitempty"`
	AllHashesValid        bool                 `json:"all_hashes_valid"`
	InvalidHashes         []string             `json:"invalid_hashes,omitempty"`
	MissingAssets         []string             `json:"missing_assets,omitempty"`
}

// AssetMapAnalysis contains Asset Map analysis for IMF packages.
// The Asset Map provides the file system structure and asset locations.
type AssetMapAnalysis struct {
	AssetMapExists        bool                 `json:"asset_map_exists"`
	AssetMapID            string               `json:"asset_map_id,omitempty"`
	AssetCount            int                  `json:"asset_count"`
	VolumeCount           int                  `json:"volume_count"`
	PathCompliance        *PathValidation      `json:"path_compliance,omitempty"`
	AssetReferences       []string             `json:"asset_references,omitempty"`
	OrphanedAssets        []string             `json:"orphaned_assets,omitempty"`
}

// TrackFileAnalysis contains analysis of individual track files in IMF packages.
type TrackFileAnalysis struct {
	TrackID               string               `json:"track_id"`
	TrackType             string               `json:"track_type"` // video, audio, subtitle
	FilePath              string               `json:"file_path"`
	Duration              string               `json:"duration"`
	EditRate              string               `json:"edit_rate"`
	EssenceDescriptor     *EssenceDescriptor   `json:"essence_descriptor,omitempty"`
	MXFCompliance         *MXFCompliance       `json:"mxf_compliance,omitempty"`
	ColorCompliance       *ColorCompliance     `json:"color_compliance,omitempty"`
	AudioCompliance       *AudioCompliance     `json:"audio_compliance,omitempty"`
	SubtitleCompliance    *SubtitleCompliance  `json:"subtitle_compliance,omitempty"`
}

// NetflixIMFCompliance contains Netflix-specific IMF delivery requirements.
type NetflixIMFCompliance struct {
	NetflixProfileCompliant bool                     `json:"netflix_profile_compliant"`
	VideoRequirements       *NetflixVideoReqs        `json:"video_requirements,omitempty"`
	AudioRequirements       *NetflixAudioReqs        `json:"audio_requirements,omitempty"`
	SubtitleRequirements    *NetflixSubtitleReqs     `json:"subtitle_requirements,omitempty"`
	DeliveryRequirements    *NetflixDeliveryReqs     `json:"delivery_requirements,omitempty"`
	ComplianceIssues        []string                 `json:"compliance_issues,omitempty"`
}

// SMPTE2067Compliance contains SMPTE 2067 standard compliance analysis.
type SMPTE2067Compliance struct {
	CoreConstraints       *CoreConstraints         `json:"core_constraints,omitempty"`
	EssenceConstraints    *EssenceConstraints      `json:"essence_constraints,omitempty"`
	PackagingConstraints  *PackagingConstraints    `json:"packaging_constraints,omitempty"`
	ApplicationProfile    string                   `json:"application_profile,omitempty"`
	ComplianceLevel       string                   `json:"compliance_level"` // full, partial, non_compliant
}

// IMFSegment represents a segment within an IMF composition.
type IMFSegment struct {
	SegmentID             string               `json:"segment_id"`
	Duration              string               `json:"duration"`
	TrackCount            int                  `json:"track_count"`
	Tracks                []IMFTrack           `json:"tracks,omitempty"`
}

// IMFTrack represents a virtual track within an IMF segment.
type IMFTrack struct {
	TrackID               string               `json:"track_id"`
	TrackType             string               `json:"track_type"`
	EditRate              string               `json:"edit_rate"`
	Duration              string               `json:"duration"`
	ResourceCount         int                  `json:"resource_count"`
	Resources             []IMFAsset           `json:"resources,omitempty"`
}

// IMFAsset represents an asset reference within an IMF track.
type IMFAsset struct {
	AssetID               string               `json:"asset_id"`
	FilePath              string               `json:"file_path"`
	EditRate              string               `json:"edit_rate"`
	Duration              string               `json:"duration"`
	EntryPoint            string               `json:"entry_point,omitempty"`
	ChunkMapping          *ChunkMapping        `json:"chunk_mapping,omitempty"`
}

// ChunkMapping represents chunk-to-frame mapping for IMF assets.
type ChunkMapping struct {
	ChunkCount            int                  `json:"chunk_count"`
	FramesPerChunk        int                  `json:"frames_per_chunk"`
	TotalFrames           int                  `json:"total_frames"`
	ChunkBoundaries       []int                `json:"chunk_boundaries,omitempty"`
}

// EssenceDescriptor contains essence-level metadata for IMF track files.
type EssenceDescriptor struct {
	EssenceType           string               `json:"essence_type"`
	Codec                 string               `json:"codec"`
	BitRate               int64                `json:"bit_rate,omitempty"`
	SampleRate            int                  `json:"sample_rate,omitempty"`
	ChannelCount          int                  `json:"channel_count,omitempty"`
	Resolution            string               `json:"resolution,omitempty"`
	FrameRate             string               `json:"frame_rate,omitempty"`
	ColorSpace            string               `json:"color_space,omitempty"`
}

// MXFCompliance contains MXF-specific compliance information.
type MXFCompliance struct {
	MXFCompliant          bool                 `json:"mxf_compliant"`
	MXFVersion            string               `json:"mxf_version,omitempty"`
	OperationalPattern    string               `json:"operational_pattern,omitempty"`
	EssenceContainer      string               `json:"essence_container,omitempty"`
	IndexTableCompliant   bool                 `json:"index_table_compliant"`
	HeaderMetadataValid   bool                 `json:"header_metadata_valid"`
	FooterExists          bool                 `json:"footer_exists"`
	RandomIndexPackExists bool                 `json:"random_index_pack_exists"`
}

// ColorCompliance contains color space and HDR compliance information.
type ColorCompliance struct {
	ColorSpaceCompliant   bool                 `json:"color_space_compliant"`
	ColorPrimaries        string               `json:"color_primaries,omitempty"`
	TransferCharacteristic string              `json:"transfer_characteristic,omitempty"`
	MatrixCoefficients    string               `json:"matrix_coefficients,omitempty"`
	HDRCompliant          bool                 `json:"hdr_compliant"`
	MaxLuminance          float64              `json:"max_luminance,omitempty"`
	MinLuminance          float64              `json:"min_luminance,omitempty"`
	MaxContentLightLevel  int                  `json:"max_content_light_level,omitempty"`
	MaxFrameAverageLightLevel int              `json:"max_frame_average_light_level,omitempty"`
}

// AudioCompliance contains audio track compliance information.
type AudioCompliance struct {
	AudioCompliant        bool                 `json:"audio_compliant"`
	ChannelConfiguration  string               `json:"channel_configuration,omitempty"`
	SampleRate            int                  `json:"sample_rate"`
	BitDepth              int                  `json:"bit_depth"`
	AudioCodec            string               `json:"audio_codec,omitempty"`
	LoudnessCompliant     bool                 `json:"loudness_compliant"`
	IntegratedLoudness    float64              `json:"integrated_loudness,omitempty"`
	LoudnessRange         float64              `json:"loudness_range,omitempty"`
	TruePeak              float64              `json:"true_peak,omitempty"`
}

// SubtitleCompliance contains subtitle track compliance information.
type SubtitleCompliance struct {
	SubtitleCompliant     bool                 `json:"subtitle_compliant"`
	SubtitleFormat        string               `json:"subtitle_format,omitempty"`
	LanguageTag           string               `json:"language_tag,omitempty"`
	SubtitleCodec         string               `json:"subtitle_codec,omitempty"`
	TimingAccuracy        bool                 `json:"timing_accuracy"`
	FontCompliance        bool                 `json:"font_compliance"`
	ColorCompliance       bool                 `json:"color_compliance"`
}

// Netflix-specific requirement structures

// NetflixVideoReqs contains Netflix video delivery requirements.
type NetflixVideoReqs struct {
	ResolutionCompliant   bool                 `json:"resolution_compliant"`
	FrameRateCompliant    bool                 `json:"frame_rate_compliant"`
	CodecCompliant        bool                 `json:"codec_compliant"`
	BitrateCompliant      bool                 `json:"bitrate_compliant"`
	ColorSpaceCompliant   bool                 `json:"color_space_compliant"`
	HDRCompliant          bool                 `json:"hdr_compliant"`
	RequiredResolution    string               `json:"required_resolution,omitempty"`
	RequiredFrameRate     string               `json:"required_frame_rate,omitempty"`
	RequiredCodec         string               `json:"required_codec,omitempty"`
}

// NetflixAudioReqs contains Netflix audio delivery requirements.
type NetflixAudioReqs struct {
	ChannelConfigCompliant bool                `json:"channel_config_compliant"`
	SampleRateCompliant   bool                 `json:"sample_rate_compliant"`
	CodecCompliant        bool                 `json:"codec_compliant"`
	LoudnessCompliant     bool                 `json:"loudness_compliant"`
	RequiredChannelConfig string               `json:"required_channel_config,omitempty"`
	RequiredSampleRate    int                  `json:"required_sample_rate,omitempty"`
	RequiredCodec         string               `json:"required_codec,omitempty"`
}

// NetflixSubtitleReqs contains Netflix subtitle delivery requirements.
type NetflixSubtitleReqs struct {
	FormatCompliant       bool                 `json:"format_compliant"`
	LanguageCompliant     bool                 `json:"language_compliant"`
	TimingCompliant       bool                 `json:"timing_compliant"`
	RequiredFormat        string               `json:"required_format,omitempty"`
	RequiredLanguages     []string             `json:"required_languages,omitempty"`
}

// NetflixDeliveryReqs contains Netflix package delivery requirements.
type NetflixDeliveryReqs struct {
	PackageStructureCompliant bool            `json:"package_structure_compliant"`
	NamingConventionCompliant bool            `json:"naming_convention_compliant"`
	MetadataCompliant         bool            `json:"metadata_compliant"`
	SecurityCompliant         bool            `json:"security_compliant"`
	RequiredPackageStructure  string          `json:"required_package_structure,omitempty"`
	RequiredNamingConvention  string          `json:"required_naming_convention,omitempty"`
}

// Validation result structures

// IMFValidationResults contains comprehensive validation results.
type IMFValidationResults struct {
	OverallValid          bool                 `json:"overall_valid"`
	MetadataValidation    *MetadataValidation  `json:"metadata_validation,omitempty"`
	StructuralValidation  *StructuralValidation `json:"structural_validation,omitempty"`
	HashValidation        *HashValidation      `json:"hash_validation,omitempty"`
	SignatureValidation   *SignatureValidation `json:"signature_validation,omitempty"`
	ValidationErrors      []string             `json:"validation_errors,omitempty"`
	ValidationWarnings    []string             `json:"validation_warnings,omitempty"`
}

// MetadataValidation contains metadata validation results.
type MetadataValidation struct {
	CPLValid              bool                 `json:"cpl_valid"`
	PKLValid              bool                 `json:"pkl_valid"`
	AssetMapValid         bool                 `json:"asset_map_valid"`
	XMLSchemaValid        bool                 `json:"xml_schema_valid"`
	UUIDsValid            bool                 `json:"uuids_valid"`
	TimestampsValid       bool                 `json:"timestamps_valid"`
}

// StructuralValidation contains package structure validation results.
type StructuralValidation struct {
	DirectoryStructureValid bool              `json:"directory_structure_valid"`
	FileNamingValid         bool              `json:"file_naming_valid"`
	AssetReferencesValid    bool              `json:"asset_references_valid"`
	TrackMappingValid       bool              `json:"track_mapping_valid"`
	SegmentContinuityValid  bool              `json:"segment_continuity_valid"`
}

// HashValidation contains cryptographic hash validation results.
type HashValidation struct {
	AllHashesValid        bool                 `json:"all_hashes_valid"`
	HashAlgorithmSupported bool               `json:"hash_algorithm_supported"`
	InvalidAssets         []string             `json:"invalid_assets,omitempty"`
	MissingAssets         []string             `json:"missing_assets,omitempty"`
}

// SignatureValidation contains digital signature validation results.
type SignatureValidation struct {
	SignaturesValid       bool                 `json:"signatures_valid"`
	CertificateChainValid bool                 `json:"certificate_chain_valid"`
	SigningTimeValid      bool                 `json:"signing_time_valid"`
	SignatureAlgorithm    string               `json:"signature_algorithm,omitempty"`
}

// PathValidation contains file path validation results.
type PathValidation struct {
	PathsCompliant        bool                 `json:"paths_compliant"`
	IllegalCharacters     []string             `json:"illegal_characters,omitempty"`
	PathTooLong           []string             `json:"path_too_long,omitempty"`
	CaseSensitivityIssues []string             `json:"case_sensitivity_issues,omitempty"`
}

// SMPTE 2067 constraint structures

// CoreConstraints contains SMPTE 2067 core constraint validation.
type CoreConstraints struct {
	SchemaVersion         string               `json:"schema_version"`
	NamespaceCompliant    bool                 `json:"namespace_compliant"`
	UUIDFormat            bool                 `json:"uuid_format"`
	EditRateCompliant     bool                 `json:"edit_rate_compliant"`
	DurationFormat        bool                 `json:"duration_format"`
	TimecodeCompliant     bool                 `json:"timecode_compliant"`
}

// EssenceConstraints contains essence-level constraint validation.
type EssenceConstraints struct {
	VideoConstraints      bool                 `json:"video_constraints"`
	AudioConstraints      bool                 `json:"audio_constraints"`
	SubtitleConstraints   bool                 `json:"subtitle_constraints"`
	CodecCompliant        bool                 `json:"codec_compliant"`
	ContainerCompliant    bool                 `json:"container_compliant"`
	ProfileCompliant      bool                 `json:"profile_compliant"`
}

// PackagingConstraints contains packaging constraint validation.
type PackagingConstraints struct {
	AssetMapConstraints   bool                 `json:"asset_map_constraints"`
	PKLConstraints        bool                 `json:"pkl_constraints"`
	CPLConstraints        bool                 `json:"cpl_constraints"`
	DirectoryStructure    bool                 `json:"directory_structure"`
	FileNamingConvention  bool                 `json:"file_naming_convention"`
	AssetOrganization     bool                 `json:"asset_organization"`
}