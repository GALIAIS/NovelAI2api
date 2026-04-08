package model

type ImageGenerateRequest struct {
	Prompt                                   string            `json:"prompt,omitempty"`
	NegativePrompt                           string            `json:"negative_prompt,omitempty"`
	Model                                    string            `json:"model,omitempty"`
	Action                                   string            `json:"action,omitempty"`
	RecaptchaToken                           string            `json:"recaptcha_token,omitempty"`
	Width                                    int               `json:"width,omitempty"`
	Height                                   int               `json:"height,omitempty"`
	Steps                                    int               `json:"steps,omitempty"`
	Scale                                    int               `json:"scale,omitempty"`
	Sampler                                  string            `json:"sampler,omitempty"`
	Seed                                     int64             `json:"seed,omitempty"`
	NSamples                                 int               `json:"n_samples,omitempty"`
	Stream                                   bool              `json:"stream,omitempty"`
	Strength                                 *float64          `json:"strength,omitempty"`
	Noise                                    *float64          `json:"noise,omitempty"`
	ExtraNoiseSeed                           *int64            `json:"extra_noise_seed,omitempty"`
	ColorCorrect                             *bool             `json:"color_correct,omitempty"`
	AddOriginalImage                         *bool             `json:"add_original_image,omitempty"`
	ImageCacheKey                            string            `json:"image_cache_secret_key,omitempty"`
	InpaintStrength                          *float64          `json:"inpaintImg2ImgStrength,omitempty"`
	NoiseSchedule                            string            `json:"noise_schedule,omitempty"`
	CFGRescale                               *float64          `json:"cfg_rescale,omitempty"`
	QualityToggle                            *bool             `json:"qualityToggle,omitempty"`
	AutoSMEA                                 *bool             `json:"autoSmea,omitempty"`
	SM                                       *bool             `json:"sm,omitempty"`
	SMDyn                                    *bool             `json:"sm_dyn,omitempty"`
	DynamicThreshold                         *bool             `json:"dynamic_thresholding,omitempty"`
	ControlnetStrength                       *float64          `json:"controlnet_strength,omitempty"`
	Legacy                                   *bool             `json:"legacy,omitempty"`
	LegacyV3Extend                           *bool             `json:"legacy_v3_extend,omitempty"`
	SkipCFGAboveSigma                        *float64          `json:"skip_cfg_above_sigma,omitempty"`
	UseCoords                                *bool             `json:"use_coords,omitempty"`
	LegacyUC                                 *bool             `json:"legacy_uc,omitempty"`
	UCPreset                                 *int              `json:"ucPreset,omitempty"`
	NormalizeReference                       *bool             `json:"normalize_reference_strength_multiple,omitempty"`
	CharacterPrompts                         []any             `json:"characterPrompts,omitempty"`
	V4Prompt                                 map[string]any    `json:"v4_prompt,omitempty"`
	V4NegativePrompt                         map[string]any    `json:"v4_negative_prompt,omitempty"`
	DeliberateEulerBug                       *bool             `json:"deliberate_euler_ancestral_bug,omitempty"`
	PreferBrownian                           *bool             `json:"prefer_brownian,omitempty"`
	ImageFormat                              string            `json:"image_format,omitempty"`
	ParamsVersion                            *int              `json:"params_version,omitempty"`
	ImageField                               string            `json:"image,omitempty"`
	MaskField                                string            `json:"mask,omitempty"`
	ReferenceStrengthMultiple                []float64         `json:"reference_strength_multiple,omitempty"`
	ReferenceImageMultipleCached             []map[string]any  `json:"reference_image_multiple_cached,omitempty"`
	DirectorReferenceDescriptions            []map[string]any  `json:"director_reference_descriptions,omitempty"`
	DirectorReferenceInformationExtracted    []int             `json:"director_reference_information_extracted,omitempty"`
	DirectorReferenceStrengthValues          []float64         `json:"director_reference_strength_values,omitempty"`
	DirectorReferenceSecondaryStrengthValues []float64         `json:"director_reference_secondary_strength_values,omitempty"`
	DirectorReferenceImagesCached            []map[string]any  `json:"director_reference_images_cached,omitempty"`
	Parameters                               map[string]any    `json:"parameters,omitempty"`
	RawRequest                               map[string]any    `json:"raw_request,omitempty"`
	Files                                    map[string]string `json:"files,omitempty"` // base64 payloads keyed by multipart field name
	UseNewSharedTrial                        *bool             `json:"use_new_shared_trial,omitempty"`
}

type DirectorToolRequest struct {
	Tool  string `json:"tool" binding:"required"`
	Image string `json:"image" binding:"required"`
}

type EncodeVibeRequest struct {
	Image                string `json:"image" binding:"required"`
	Model                string `json:"model,omitempty"`
	InformationExtracted int    `json:"information_extracted,omitempty"`
	Mask                 string `json:"mask,omitempty"`
}
