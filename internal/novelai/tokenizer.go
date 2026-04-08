package novelai

import (
	"fmt"
	"strings"
)

type Tokenizer interface {
	Encode(text string, tokenizer string) ([]int, error)
	Decode(tokenIDs []int, tokenizer string) (string, error)
}

type LocalTokenizer struct{}

func NewLocalTokenizer() *LocalTokenizer {
	return &LocalTokenizer{}
}

func (t *LocalTokenizer) Encode(text string, tokenizer string) ([]int, error) {
	if !isSupportedTokenizer(tokenizer) {
		return nil, fmt.Errorf("unsupported tokenizer: %s", tokenizer)
	}
	out := make([]int, 0, len(text))
	for _, r := range text {
		out = append(out, int(r))
	}
	return out, nil
}

func (t *LocalTokenizer) Decode(tokenIDs []int, tokenizer string) (string, error) {
	if !isSupportedTokenizer(tokenizer) {
		return "", fmt.Errorf("unsupported tokenizer: %s", tokenizer)
	}
	runes := make([]rune, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		runes = append(runes, rune(tokenID))
	}
	return string(runes), nil
}

func isSupportedTokenizer(name string) bool {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if _, ok := SupportedTokenizers[normalized]; ok {
		return true
	}
	for _, canonical := range SupportedTokenizers {
		if canonical == name {
			return true
		}
	}
	return false
}
