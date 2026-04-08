package novelai

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type StoriesClient interface {
	GetKeystore(ctx context.Context, token string) (map[string]any, error)
	PutKeystore(ctx context.Context, token string, payload map[string]any) (map[string]any, error)
	GetSubscription(ctx context.Context, token string) (map[string]any, error)
	ListObjects(ctx context.Context, token string, objectType string) (map[string]any, error)
	PutObject(ctx context.Context, token string, objectType string, payload map[string]any) (map[string]any, error)
	PatchObject(ctx context.Context, token string, objectType string, id string, payload map[string]any) (map[string]any, error)
	DeleteObject(ctx context.Context, token string, objectType string, id string) error
}

func (c *Client) GetKeystore(ctx context.Context, token string) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, "GET", c.APIBase+"/user/keystore", token, nil, &out)
	return out, err
}

func (c *Client) PutKeystore(ctx context.Context, token string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, "PUT", c.APIBase+"/user/keystore", token, payload, &out)
	return out, err
}

func (c *Client) GetSubscription(ctx context.Context, token string) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, "GET", c.APIBase+"/user/subscription", token, nil, &out)
	return out, err
}

func (c *Client) ListObjects(ctx context.Context, token string, objectType string) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, "GET", c.ImageBase+"/user/objects/"+escapePathSegment(objectType), token, nil, &out)
	return out, err
}

func (c *Client) PutObject(ctx context.Context, token string, objectType string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.doJSON(ctx, "PUT", c.ImageBase+"/user/objects/"+escapePathSegment(objectType), token, payload, &out)
	return out, err
}

func (c *Client) PatchObject(ctx context.Context, token string, objectType string, id string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	path := c.ImageBase + "/user/objects/" + escapePathSegment(objectType) + "/" + escapePathSegment(id)
	err := c.doJSON(ctx, "PATCH", path, token, payload, &out)
	return out, err
}

func (c *Client) DeleteObject(ctx context.Context, token string, objectType string, id string) error {
	path := c.ImageBase + "/user/objects/" + escapePathSegment(objectType) + "/" + escapePathSegment(id)
	resp, err := c.do(ctx, "DELETE", path, token, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	return nil
}

func escapePathSegment(v string) string {
	return strings.ReplaceAll(url.PathEscape(v), "+", "%20")
}

func ValidateObjectType(objectType string) error {
	if strings.TrimSpace(objectType) == "" {
		return fmt.Errorf("object_type is required")
	}
	if strings.Contains(objectType, "/") || strings.Contains(objectType, `\`) {
		return fmt.Errorf("object_type contains invalid slash")
	}
	return nil
}

