package novelai

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type CompletionChunk struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func (c CompletionChunk) Text() string {
	var out strings.Builder
	for _, choice := range c.Choices {
		out.WriteString(choice.Text)
	}
	return out.String()
}

func ParseCompletionStream(r io.Reader) ([]CompletionChunk, error) {
	scanner := bufio.NewScanner(r)
	var out []CompletionChunk
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}
		var chunk CompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return nil, err
		}
		out = append(out, chunk)
	}
	return out, scanner.Err()
}
