package service

import (
	"encoding/base64"
	"testing"

	"novelai/internal/model"
)

func TestImageServiceGenerate(t *testing.T) {
	svc := &ImageService{}
	resp, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{Prompt: "cat"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Images) != 1 {
		t.Fatalf("len = %d", len(resp.Images))
	}
}

func TestImageServiceGenerateAppliesSafeDefaults(t *testing.T) {
	svc := &ImageService{}
	events, err := svc.GenerateStream(t.Context(), "token", model.ImageGenerateRequest{Prompt: "cat", Stream: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len = %d", len(events))
	}
	if events[0].StepIX != defaultImageSteps {
		t.Fatalf("step_ix = %d", events[0].StepIX)
	}
}

func TestImageServiceGenerateAcceptsRawInputWithoutPrompt(t *testing.T) {
	svc := &ImageService{}
	_, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{
		RawRequest: map[string]any{
			"input": "cat",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImageServiceAcceptsCustomLargeValues(t *testing.T) {
	svc := &ImageService{}
	_, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{
		Prompt:   "cat",
		Width:    2048,
		Height:   2048,
		Steps:    40,
		NSamples: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImageServiceMergesParametersOverride(t *testing.T) {
	svc := &ImageService{}
	_, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{
		Prompt: "cat",
		Width:  1024,
		Parameters: map[string]any{
			"width": 1536,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestImageServiceRejectsInvalidFileBase64(t *testing.T) {
	svc := &ImageService{}
	_, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{
		Prompt: "cat",
		Files: map[string]string{
			"image": "%%%not-base64%%%",
		},
	})
	if err == nil {
		t.Fatal("expected invalid base64 error")
	}
}

func TestImageServiceAcceptsFilePayload(t *testing.T) {
	svc := &ImageService{}
	_, err := svc.Generate(t.Context(), "token", model.ImageGenerateRequest{
		Prompt: "cat",
		Files: map[string]string{
			"image": base64.StdEncoding.EncodeToString([]byte("png")),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDirectorTool(t *testing.T) {
	if err := ValidateDirectorTool("remove_bg"); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeImageGenerateRequestMapsOfficialTopLevelParameters(t *testing.T) {
	strength := 0.7
	noise := 0.0
	ucPreset := 0
	qualityToggle := true
	req, err := normalizeImageGenerateRequest(model.ImageGenerateRequest{
		Prompt:         "cat",
		Strength:       &strength,
		Noise:          &noise,
		UCPreset:       &ucPreset,
		QualityToggle:  &qualityToggle,
		NoiseSchedule:  "karras",
		ImageField:     "image",
		RecaptchaToken: "captcha-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Parameters["strength"] != strength {
		t.Fatalf("strength = %v", req.Parameters["strength"])
	}
	if req.Parameters["noise"] != noise {
		t.Fatalf("noise = %v", req.Parameters["noise"])
	}
	if req.Parameters["ucPreset"] != ucPreset {
		t.Fatalf("ucPreset = %v", req.Parameters["ucPreset"])
	}
	if req.Parameters["qualityToggle"] != qualityToggle {
		t.Fatalf("qualityToggle = %v", req.Parameters["qualityToggle"])
	}
	if req.Parameters["noise_schedule"] != "karras" {
		t.Fatalf("noise_schedule = %v", req.Parameters["noise_schedule"])
	}
	if req.Parameters["image"] != "image" {
		t.Fatalf("image = %v", req.Parameters["image"])
	}
	if req.RawRequest["recaptcha_token"] != "captcha-token" {
		t.Fatalf("recaptcha_token = %v", req.RawRequest["recaptcha_token"])
	}
}

func TestNormalizeImageGenerateRequestKeepsParametersPriorityOverTopLevel(t *testing.T) {
	strength := 0.7
	req, err := normalizeImageGenerateRequest(model.ImageGenerateRequest{
		Prompt:   "cat",
		Strength: &strength,
		Parameters: map[string]any{
			"strength": 0.2,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Parameters["strength"] != 0.2 {
		t.Fatalf("strength = %v", req.Parameters["strength"])
	}
}
