package model

type TokenizerEncodeRequest struct {
	Text      string `json:"text" binding:"required"`
	Tokenizer string `json:"tokenizer" binding:"required"`
}

type TokenizerDecodeRequest struct {
	TokenIDs  []int  `json:"token_ids" binding:"required"`
	Tokenizer string `json:"tokenizer" binding:"required"`
}
