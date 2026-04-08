package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

const (
	defaultImageModel   = "nai-diffusion-4-5-full"
	defaultImageAction  = "generate"
	defaultImageWidth   = 1024
	defaultImageHeight  = 1024
	defaultImageSteps   = 28
	defaultImageScale   = 5
	defaultImageSampler = "k_euler_ancestral"
	defaultImageSamples = 1
)

type ImageService struct {
	Client novelai.ImageClient
}

func (s *ImageService) Generate(ctx context.Context, token string, req model.ImageGenerateRequest) (*model.ImageResponse, error) {
	if req.Prompt == "" && !hasRawInput(req.RawRequest) {
		return nil, fmt.Errorf("prompt is required (or provide raw_request.input)")
	}
	normalized, err := normalizeImageGenerateRequest(req)
	if err != nil {
		return nil, err
	}
	files, err := decodeFiles(normalized.Files)
	if err != nil {
		return nil, err
	}
	if s.Client != nil {
		result, err := s.Client.Generate(ctx, token, novelai.ImageGenerateRequest{
			Prompt:            normalized.Prompt,
			NegativePrompt:    normalized.NegativePrompt,
			Model:             normalized.Model,
			Action:            normalized.Action,
			Width:             normalized.Width,
			Height:            normalized.Height,
			Steps:             normalized.Steps,
			Scale:             normalized.Scale,
			Sampler:           normalized.Sampler,
			Seed:              normalized.Seed,
			NSamples:          normalized.NSamples,
			Parameters:        normalized.Parameters,
			RawRequest:        normalized.RawRequest,
			Files:             files,
			UseNewSharedTrial: normalized.UseNewSharedTrial,
		})
		if err != nil {
			return nil, err
		}
		return toImageResponse(result), nil
	}
	payload := base64.StdEncoding.EncodeToString([]byte("stub-image:" + normalized.Prompt))
	return &model.ImageResponse{
		Images: []model.ImagePayload{{MIMEType: "image/png", Base64: payload}},
		Meta:   map[string]int64{"seed": normalized.Seed},
	}, nil
}

func (s *ImageService) GenerateStream(ctx context.Context, token string, req model.ImageGenerateRequest) ([]novelai.ImageStreamEvent, error) {
	if req.Prompt == "" && !hasRawInput(req.RawRequest) {
		return nil, fmt.Errorf("prompt is required (or provide raw_request.input)")
	}
	normalized, err := normalizeImageGenerateRequest(req)
	if err != nil {
		return nil, err
	}
	files, err := decodeFiles(normalized.Files)
	if err != nil {
		return nil, err
	}
	if s.Client != nil {
		return s.Client.GenerateStream(ctx, token, novelai.ImageGenerateRequest{
			Prompt:            normalized.Prompt,
			NegativePrompt:    normalized.NegativePrompt,
			Model:             normalized.Model,
			Action:            normalized.Action,
			Width:             normalized.Width,
			Height:            normalized.Height,
			Steps:             normalized.Steps,
			Scale:             normalized.Scale,
			Sampler:           normalized.Sampler,
			Seed:              normalized.Seed,
			NSamples:          normalized.NSamples,
			Parameters:        normalized.Parameters,
			RawRequest:        normalized.RawRequest,
			Files:             files,
			UseNewSharedTrial: normalized.UseNewSharedTrial,
		})
	}
	return []novelai.ImageStreamEvent{{EventType: "final", Image: []byte("stub-image"), SampIX: 0, StepIX: normalized.Steps}}, nil
}

func ValidateDirectorTool(tool string) error {
	switch tool {
	case "remove_bg", "line_art", "sketch", "colorize", "emotion", "declutter":
		return nil
	default:
		return fmt.Errorf("unsupported director tool: %s", tool)
	}
}

type EncodeVibeResult struct {
	VibeCode string `json:"vibe_code"`
}

func (s *ImageService) DirectorTool(ctx context.Context, token string, req model.DirectorToolRequest) (*model.ImageResponse, error) {
	if err := ValidateDirectorTool(req.Tool); err != nil {
		return nil, err
	}
	if s.Client != nil {
		imageBytes, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			return nil, err
		}
		result, err := s.Client.DirectorTool(ctx, token, novelai.DirectorToolRequest{Tool: req.Tool, Image: imageBytes})
		if err != nil {
			return nil, err
		}
		return toImageResponse(result), nil
	}
	payload := base64.StdEncoding.EncodeToString([]byte("tool:" + req.Tool))
	return &model.ImageResponse{Images: []model.ImagePayload{{MIMEType: "image/png", Base64: payload}}}, nil
}

func (s *ImageService) EncodeVibe(ctx context.Context, token string, req model.EncodeVibeRequest) (*EncodeVibeResult, error) {
	if req.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	if s.Client != nil {
		imageBytes, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			return nil, err
		}
		maskBytes := []byte(nil)
		if req.Mask != "" {
			maskBytes, err = base64.StdEncoding.DecodeString(req.Mask)
			if err != nil {
				return nil, err
			}
		}
		result, err := s.Client.EncodeVibe(ctx, token, novelai.EncodeVibeRequest{
			Image:                imageBytes,
			Model:                req.Model,
			InformationExtracted: req.InformationExtracted,
			Mask:                 maskBytes,
		})
		if err != nil {
			return nil, err
		}
		return &EncodeVibeResult{VibeCode: result.VibeCode}, nil
	}
	return &EncodeVibeResult{VibeCode: base64.StdEncoding.EncodeToString([]byte(req.Image))}, nil
}

func toImageResponse(result *novelai.ImageGenerateResult) *model.ImageResponse {
	resp := &model.ImageResponse{
		Images: make([]model.ImagePayload, 0, len(result.Images)),
		Meta:   map[string]int64{"seed": result.Seed},
	}
	for _, image := range result.Images {
		resp.Images = append(resp.Images, model.ImagePayload{
			MIMEType: image.MIMEType,
			Base64:   base64.StdEncoding.EncodeToString(image.Bytes),
		})
	}
	return resp
}

func normalizeImageGenerateRequest(req model.ImageGenerateRequest) (model.ImageGenerateRequest, error) {
	if req.Action == "" {
		req.Action = defaultImageAction
	}
	if req.Width == 0 {
		req.Width = defaultImageWidth
	}
	if req.Height == 0 {
		req.Height = defaultImageHeight
	}
	if req.Steps == 0 {
		req.Steps = defaultImageSteps
	}
	if req.Scale == 0 {
		req.Scale = defaultImageScale
	}
	if req.Sampler == "" {
		req.Sampler = defaultImageSampler
	}
	if req.NSamples == 0 {
		req.NSamples = defaultImageSamples
	}
	if req.Model == "" {
		req.Model = defaultImageModel
	}
	if req.Parameters == nil {
		req.Parameters = map[string]any{}
	}
	if req.Files == nil {
		req.Files = map[string]string{}
	}
	if req.RawRequest == nil {
		req.RawRequest = map[string]any{}
	}
	if req.RecaptchaToken != "" {
		setDefaultTopLevel(req.RawRequest, "recaptcha_token", req.RecaptchaToken)
	}

	// Merge scalar fields into parameters only when caller did not already pass them.
	setDefaultParameter(req.Parameters, "width", req.Width)
	setDefaultParameter(req.Parameters, "height", req.Height)
	setDefaultParameter(req.Parameters, "steps", req.Steps)
	setDefaultParameter(req.Parameters, "scale", req.Scale)
	setDefaultParameter(req.Parameters, "sampler", req.Sampler)
	setDefaultParameter(req.Parameters, "seed", req.Seed)
	setDefaultParameter(req.Parameters, "n_samples", req.NSamples)
	setDefaultParameter(req.Parameters, "uc", req.NegativePrompt)
	applyOfficialImageParameterDefaults(req.Parameters, req)

	// Keep minimal sanity checks only; no paid-range safety guardrails anymore.
	if req.Width < 1 || req.Height < 1 {
		return req, fmt.Errorf("width and height must be positive")
	}
	if req.Steps < 1 {
		return req, fmt.Errorf("steps must be positive")
	}
	if req.NSamples < 1 {
		return req, fmt.Errorf("n_samples must be positive")
	}
	return req, nil
}

func hasRawInput(raw map[string]any) bool {
	if len(raw) == 0 {
		return false
	}
	value, ok := raw["input"]
	if !ok {
		return false
	}
	input, ok := value.(string)
	return ok && input != ""
}

func setDefaultParameter(parameters map[string]any, key string, value any) {
	if _, exists := parameters[key]; exists {
		return
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			return
		}
	case int:
		if v == 0 {
			return
		}
	case int64:
		if v == 0 {
			return
		}
	}
	parameters[key] = value
}

func setDefaultTopLevel(payload map[string]any, key string, value any) {
	if payload == nil {
		return
	}
	if _, exists := payload[key]; exists {
		return
	}
	payload[key] = value
}

func setDefaultParameterBool(parameters map[string]any, key string, value *bool) {
	if value == nil {
		return
	}
	if _, exists := parameters[key]; exists {
		return
	}
	parameters[key] = *value
}

func setDefaultParameterFloat(parameters map[string]any, key string, value *float64) {
	if value == nil {
		return
	}
	if _, exists := parameters[key]; exists {
		return
	}
	parameters[key] = *value
}

func setDefaultParameterInt64(parameters map[string]any, key string, value *int64) {
	if value == nil {
		return
	}
	if _, exists := parameters[key]; exists {
		return
	}
	parameters[key] = *value
}

func setDefaultParameterInt(parameters map[string]any, key string, value *int) {
	if value == nil {
		return
	}
	if _, exists := parameters[key]; exists {
		return
	}
	parameters[key] = *value
}

func setDefaultParameterSlice[T any](parameters map[string]any, key string, value []T) {
	if len(value) == 0 {
		return
	}
	setDefaultParameter(parameters, key, value)
}

func setDefaultParameterMap(parameters map[string]any, key string, value map[string]any) {
	if len(value) == 0 {
		return
	}
	setDefaultParameter(parameters, key, value)
}

func applyOfficialImageParameterDefaults(parameters map[string]any, req model.ImageGenerateRequest) {
	setDefaultParameterFloat(parameters, "strength", req.Strength)
	setDefaultParameterFloat(parameters, "noise", req.Noise)
	setDefaultParameterInt64(parameters, "extra_noise_seed", req.ExtraNoiseSeed)
	setDefaultParameterBool(parameters, "color_correct", req.ColorCorrect)
	setDefaultParameterBool(parameters, "add_original_image", req.AddOriginalImage)
	setDefaultParameter(parameters, "image_cache_secret_key", req.ImageCacheKey)
	setDefaultParameterFloat(parameters, "inpaintImg2ImgStrength", req.InpaintStrength)
	setDefaultParameter(parameters, "noise_schedule", req.NoiseSchedule)
	setDefaultParameterFloat(parameters, "cfg_rescale", req.CFGRescale)
	setDefaultParameterBool(parameters, "qualityToggle", req.QualityToggle)
	setDefaultParameterBool(parameters, "autoSmea", req.AutoSMEA)
	setDefaultParameterBool(parameters, "sm", req.SM)
	setDefaultParameterBool(parameters, "sm_dyn", req.SMDyn)
	setDefaultParameterBool(parameters, "dynamic_thresholding", req.DynamicThreshold)
	setDefaultParameterFloat(parameters, "controlnet_strength", req.ControlnetStrength)
	setDefaultParameterBool(parameters, "legacy", req.Legacy)
	setDefaultParameterBool(parameters, "legacy_v3_extend", req.LegacyV3Extend)
	setDefaultParameterFloat(parameters, "skip_cfg_above_sigma", req.SkipCFGAboveSigma)
	setDefaultParameterBool(parameters, "use_coords", req.UseCoords)
	setDefaultParameterBool(parameters, "legacy_uc", req.LegacyUC)
	setDefaultParameterInt(parameters, "ucPreset", req.UCPreset)
	setDefaultParameterBool(parameters, "normalize_reference_strength_multiple", req.NormalizeReference)
	setDefaultParameterSlice(parameters, "characterPrompts", req.CharacterPrompts)
	setDefaultParameterMap(parameters, "v4_prompt", req.V4Prompt)
	setDefaultParameterMap(parameters, "v4_negative_prompt", req.V4NegativePrompt)
	setDefaultParameterBool(parameters, "deliberate_euler_ancestral_bug", req.DeliberateEulerBug)
	setDefaultParameterBool(parameters, "prefer_brownian", req.PreferBrownian)
	setDefaultParameter(parameters, "image_format", req.ImageFormat)
	setDefaultParameterInt(parameters, "params_version", req.ParamsVersion)
	setDefaultParameter(parameters, "image", req.ImageField)
	setDefaultParameter(parameters, "mask", req.MaskField)
	setDefaultParameterSlice(parameters, "reference_strength_multiple", req.ReferenceStrengthMultiple)
	setDefaultParameterSlice(parameters, "reference_image_multiple_cached", req.ReferenceImageMultipleCached)
	setDefaultParameterSlice(parameters, "director_reference_descriptions", req.DirectorReferenceDescriptions)
	setDefaultParameterSlice(parameters, "director_reference_information_extracted", req.DirectorReferenceInformationExtracted)
	setDefaultParameterSlice(parameters, "director_reference_strength_values", req.DirectorReferenceStrengthValues)
	setDefaultParameterSlice(parameters, "director_reference_secondary_strength_values", req.DirectorReferenceSecondaryStrengthValues)
	setDefaultParameterSlice(parameters, "director_reference_images_cached", req.DirectorReferenceImagesCached)
}

func decodeFiles(files map[string]string) (map[string][]byte, error) {
	decoded := make(map[string][]byte, len(files))
	for name, value := range files {
		if name == "" || value == "" {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, fmt.Errorf("files[%s] invalid base64: %w", name, err)
		}
		decoded[name] = data
	}
	return decoded, nil
}
