package service

import (
	"testing"

	"novelai/internal/model"
	"novelai/internal/novelai"
)

func TestNormalizeTokenizerName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "gpt2", want: "GPT2"},
		{input: "pile", want: "PILE"},
		{input: "pile-nai", want: "PILE_NAI"},
		{input: "genji", want: "GENJI"},
		{input: "clip", want: "CLIP"},
		{input: "nerdstash", want: "NERDSTASH"},
		{input: "nerdstash-v2", want: "NERDSTASH_V2"},
		{input: "llama3", want: "LLAMA3"},
		{input: "glm", want: "GLM"},
		{input: "t5", want: "T5"},
	}

	for _, tc := range tests {
		got, err := NormalizeTokenizerName(tc.input)
		if err != nil {
			t.Fatalf("%s: %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.input, got, tc.want)
		}
	}
}

func TestTokenizerServiceEncodeDecode(t *testing.T) {
	svc := &TokenizerService{Tokenizer: novelai.NewLocalTokenizer()}
	encoded, err := svc.Encode(model.TokenizerEncodeRequest{Text: "hi", Tokenizer: "gpt2"})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := svc.Decode(model.TokenizerDecodeRequest{TokenIDs: encoded.TokenIDs, Tokenizer: "gpt2"})
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Text != "hi" {
		t.Fatalf("decoded text = %q", decoded.Text)
	}
}

func TestNormalizeTokenizerNameRejectsUnknown(t *testing.T) {
	if _, err := NormalizeTokenizerName("unknown"); err == nil {
		t.Fatal("expected unsupported tokenizer error")
	}
}
