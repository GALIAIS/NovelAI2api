package novelai

type LoginResult struct {
	AccessToken string `json:"accessToken"`
}

type UserDataResult struct {
	Subscription map[string]any `json:"subscription"`
	Priority     map[string]any `json:"priority"`
	Information  map[string]any `json:"information"`
}
