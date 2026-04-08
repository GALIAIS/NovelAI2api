package model

type TokenizerEncodeResponse struct {
	TokenIDs []int `json:"token_ids"`
	Count    int   `json:"count"`
}

type TokenizerDecodeResponse struct {
	Text string `json:"text"`
}
