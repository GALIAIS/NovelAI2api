package novelai

import (
	"context"
	"net/http"
)

func (c *Client) Login(accessKey, captchaToken string) (*LoginResult, error) {
	payload := map[string]string{
		"key": accessKey,
	}
	if captchaToken != "" {
		payload["recaptcha"] = captchaToken
	}
	var out LoginResult
	if err := c.doJSON(context.Background(), http.MethodPost, c.ImageBase+"/user/login", "", payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UserData(token string) (*UserDataResult, error) {
	var out UserDataResult
	if err := c.doJSON(context.Background(), http.MethodGet, c.APIBase+"/user/data", token, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
