package model

type LoginRequest struct {
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	APIToken     string `json:"api_token,omitempty"`
	CaptchaToken string `json:"captcha_token,omitempty"`
}
