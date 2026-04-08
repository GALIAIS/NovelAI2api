package service

import (
	"fmt"
	"strings"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

type TokenizerService struct {
	Tokenizer novelai.Tokenizer
}

func NormalizeTokenizerName(name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if mapped, ok := novelai.SupportedTokenizers[normalized]; ok {
		return mapped, nil
	}
	return "", fmt.Errorf("unsupported tokenizer: %s", name)
}

func (s *TokenizerService) Encode(req model.TokenizerEncodeRequest) (*model.TokenizerEncodeResponse, error) {
	name, err := NormalizeTokenizerName(req.Tokenizer)
	if err != nil {
		return nil, err
	}
	tokenIDs, err := s.Tokenizer.Encode(req.Text, name)
	if err != nil {
		return nil, err
	}
	return &model.TokenizerEncodeResponse{TokenIDs: tokenIDs, Count: len(tokenIDs)}, nil
}

func (s *TokenizerService) Decode(req model.TokenizerDecodeRequest) (*model.TokenizerDecodeResponse, error) {
	name, err := NormalizeTokenizerName(req.Tokenizer)
	if err != nil {
		return nil, err
	}
	text, err := s.Tokenizer.Decode(req.TokenIDs, name)
	if err != nil {
		return nil, err
	}
	return &model.TokenizerDecodeResponse{Text: text}, nil
}
