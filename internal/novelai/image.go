package novelai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type ImageClient interface {
	Generate(ctx context.Context, token string, req ImageGenerateRequest) (*ImageGenerateResult, error)
	GenerateStream(ctx context.Context, token string, req ImageGenerateRequest) ([]ImageStreamEvent, error)
	DirectorTool(ctx context.Context, token string, req DirectorToolRequest) (*ImageGenerateResult, error)
	EncodeVibe(ctx context.Context, token string, req EncodeVibeRequest) (*EncodeVibeResult, error)
}

type ImageGenerateRequest struct {
	Prompt            string
	NegativePrompt    string
	Model             string
	Action            string
	Width             int
	Height            int
	Steps             int
	Scale             int
	Sampler           string
	Seed              int64
	NSamples          int
	Parameters        map[string]any
	RawRequest        map[string]any
	Files             map[string][]byte
	UseNewSharedTrial *bool
}

type DirectorToolRequest struct {
	Tool  string
	Image []byte
}

type EncodeVibeRequest struct {
	Image                []byte
	Model                string
	InformationExtracted int
	Mask                 []byte
}

type ImageGenerateResult struct {
	Images []ImageData
	Seed   int64
}

type EncodeVibeResult struct {
	VibeCode string
}

func (c *Client) Generate(ctx context.Context, token string, req ImageGenerateRequest) (*ImageGenerateResult, error) {
	events, err := c.generateStreamMultipart(ctx, token, req)
	if err != nil {
		return nil, err
	}
	images := collectImagesFromStream(events)
	return &ImageGenerateResult{Images: images, Seed: req.Seed}, nil
}

func (c *Client) GenerateStream(ctx context.Context, token string, req ImageGenerateRequest) ([]ImageStreamEvent, error) {
	if len(req.Files) > 0 {
		return c.generateStreamMultipart(ctx, token, req)
	}
	payload := buildImageGeneratePayload(req)
	resp, err := c.do(ctx, "POST", c.ImageBase+"/ai/generate-image-stream", token, "application/json", bytes.NewReader(mustJSON(payload)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	events, err := ParseImageStream(resp.Body)
	if err != nil {
		return nil, err
	}
	if streamErr := streamErrorFromEvents(events); streamErr != nil {
		return nil, streamErr
	}
	return events, nil
}

func (c *Client) DirectorTool(ctx context.Context, token string, req DirectorToolRequest) (*ImageGenerateResult, error) {
	payload := map[string]any{
		"req_type": req.Tool,
		"image":    req.Image,
	}
	encoded, err := msgpack.Marshal(payload)
	if err != nil {
		return nil, err
	}
	blob, err := c.doBytes(ctx, "POST", c.ImageBase+"/ai/augment-image", token, "application/msgpack", encoded)
	if err != nil {
		return nil, err
	}
	images, err := ExtractImagesFromZip(blob)
	if err != nil {
		return nil, err
	}
	return &ImageGenerateResult{Images: images}, nil
}

func (c *Client) EncodeVibe(ctx context.Context, token string, req EncodeVibeRequest) (*EncodeVibeResult, error) {
	payload := map[string]any{
		"image":                 req.Image,
		"model":                 req.Model,
		"information_extracted": req.InformationExtracted,
	}
	if len(req.Mask) > 0 {
		payload["mask"] = req.Mask
	}
	encoded, err := msgpack.Marshal(payload)
	if err != nil {
		return nil, err
	}
	blob, err := c.doBytes(ctx, "POST", c.ImageBase+"/ai/encode-vibe", token, "application/msgpack", encoded)
	if err != nil {
		return nil, err
	}
	return &EncodeVibeResult{VibeCode: base64.StdEncoding.EncodeToString(blob)}, nil
}

func mustJSON(v any) []byte {
	buf, _ := jsonMarshal(v)
	return buf
}

func buildImageGeneratePayload(req ImageGenerateRequest) map[string]any {
	payload := cloneMap(req.RawRequest)
	if payload == nil {
		payload = map[string]any{}
	}
	parameters := map[string]any{}
	if rawParameters, ok := payload["parameters"].(map[string]any); ok {
		for k, v := range rawParameters {
			parameters[k] = v
		}
	}
	for k, v := range req.Parameters {
		if _, exists := parameters[k]; exists {
			continue
		}
		parameters[k] = v
	}

	if _, ok := payload["parameters"]; !ok {
		payload["parameters"] = parameters
	} else {
		payload["parameters"] = parameters
	}
	applyImageParameterDefaults(parameters, req)
	if _, ok := payload["input"]; !ok && req.Prompt != "" {
		payload["input"] = req.Prompt
	}
	if _, ok := payload["model"]; !ok && req.Model != "" {
		payload["model"] = req.Model
	}
	action := req.Action
	if action == "" {
		action = "generate"
	}
	if _, ok := payload["action"]; !ok {
		payload["action"] = action
	}
	for field := range req.Files {
		if field == "" {
			continue
		}
		if _, exists := parameters[field]; !exists {
			parameters[field] = field
		}
	}
	if req.UseNewSharedTrial == nil {
		if _, exists := payload["use_new_shared_trial"]; !exists {
			payload["use_new_shared_trial"] = true
		}
	} else {
		payload["use_new_shared_trial"] = *req.UseNewSharedTrial
	}
	return payload
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func applyImageParameterDefaults(parameters map[string]any, req ImageGenerateRequest) {
	// Keep stream format aligned with current official image stream protocol.
	if _, exists := parameters["stream"]; !exists {
		parameters["stream"] = "msgpack"
	}
	if _, exists := parameters["params_version"]; !exists {
		parameters["params_version"] = 3
	}
	if _, exists := parameters["image_format"]; !exists {
		parameters["image_format"] = "png"
	}
	if req.NegativePrompt != "" {
		if _, exists := parameters["negative_prompt"]; !exists {
			parameters["negative_prompt"] = req.NegativePrompt
		}
	}
	if _, exists := parameters["v4_prompt"]; !exists && req.Prompt != "" {
		parameters["v4_prompt"] = map[string]any{
			"caption": map[string]any{
				"base_caption":  req.Prompt,
				"char_captions": []any{},
			},
			"use_coords": false,
			"use_order":  true,
		}
	}
	if _, exists := parameters["v4_negative_prompt"]; !exists && req.NegativePrompt != "" {
		parameters["v4_negative_prompt"] = map[string]any{
			"caption": map[string]any{
				"base_caption":  req.NegativePrompt,
				"char_captions": []any{},
			},
			"legacy_uc": false,
		}
	}
}

func (c *Client) generateStreamMultipart(ctx context.Context, token string, req ImageGenerateRequest) ([]ImageStreamEvent, error) {
	payload := buildImageGeneratePayload(req)
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	requestHeader := make(textproto.MIMEHeader)
	requestHeader.Set("Content-Disposition", `form-data; name="request"`)
	requestHeader.Set("Content-Type", "application/json")
	requestField, err := writer.CreatePart(requestHeader)
	if err != nil {
		return nil, err
	}
	if _, err := requestField.Write(requestBody); err != nil {
		return nil, err
	}
	for field, data := range req.Files {
		if field == "" || len(data) == 0 {
			continue
		}
		part, err := writer.CreateFormFile(field, field+".bin")
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, "POST", c.ImageBase+"/ai/generate-image-stream", token, writer.FormDataContentType(), &body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return nil, &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	return ParseImageStream(resp.Body)
}

func collectImagesFromStream(events []ImageStreamEvent) []ImageData {
	var finals []ImageData
	var fallback []ImageData
	for _, event := range events {
		if len(event.Image) == 0 {
			continue
		}
		image := ImageData{
			Filename: "image_stream.png",
			Bytes:    event.Image,
			MIMEType: "image/png",
		}
		fallback = append(fallback, image)
		if event.EventType == "final" {
			finals = append(finals, image)
		}
	}
	if len(finals) > 0 {
		return finals
	}
	return fallback
}

func streamErrorFromEvents(events []ImageStreamEvent) error {
	var messages []string
	for _, event := range events {
		if event.Message != "" && (event.EventType == "error" || event.EventType == "retry") {
			messages = append(messages, event.Message)
		}
		if event.EventType == "error" {
			code := strings.TrimSpace(event.Code)
			if code != "" {
				return fmt.Errorf("image stream error (code=%s): %s", code, event.Message)
			}
			if event.Message != "" {
				return fmt.Errorf("image stream error: %s", event.Message)
			}
			return fmt.Errorf("image stream error")
		}
	}
	if len(events) > 0 {
		hasImage := false
		for _, event := range events {
			if len(event.Image) > 0 {
				hasImage = true
				break
			}
		}
		if !hasImage && len(messages) > 0 {
			return fmt.Errorf("image stream produced no image: %s", messages[len(messages)-1])
		}
	}
	return nil
}
