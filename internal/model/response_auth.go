package model

type LoginResponse struct {
	SessionID    string         `json:"session_id"`
	User         UserSummary    `json:"user"`
	Subscription map[string]any `json:"subscription,omitempty"`
	Priority     map[string]any `json:"priority,omitempty"`
	Information  map[string]any `json:"information,omitempty"`
}

type UserSummary struct {
	Email string `json:"email"`
}

type MeResponse struct {
	Email        string         `json:"email"`
	Subscription map[string]any `json:"subscription,omitempty"`
	Priority     map[string]any `json:"priority,omitempty"`
	Information  map[string]any `json:"information,omitempty"`
}
