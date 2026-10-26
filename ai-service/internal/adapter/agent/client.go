package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL, APIKey, Model string
	HTTP                   *http.Client
}

type Error struct {
	Status     int
	RetryAfter string
}

func (e *Error) Error() string { return fmt.Sprintf("provider HTTP %d", e.Status) }

func New(url, key, model string) *Client {
	return &Client{url, key, model, &http.Client{Timeout: 45 * time.Second}}
}

func (c *Client) Complete(ctx context.Context, prompt string, input, output any) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	body, err := json.Marshal(map[string]any{"model": c.Model, "messages": []map[string]string{{"role": "system", "content": prompt}, {"role": "user", "content": string(data)}}, "response_format": map[string]string{"type": "json_object"}, "max_tokens": 2048})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("provider unavailable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return &Error{res.StatusCode, res.Header.Get("Retry-After")}
	}

	var completion struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&completion); err != nil {
		return fmt.Errorf("invalid provider response")
	}

	if len(completion.Choices) == 0 || completion.Choices[0].FinishReason != "stop" {
		return fmt.Errorf("incomplete completion")
	}

	log.Printf("model usage: prompt=%d completion=%d total=%d", completion.Usage.PromptTokens, completion.Usage.CompletionTokens, completion.Usage.TotalTokens)
	d := json.NewDecoder(strings.NewReader(completion.Choices[0].Message.Content))
	d.DisallowUnknownFields()
	if err := d.Decode(output); err != nil {
		return fmt.Errorf("invalid model JSON: %w", err)
	}

	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing model output")
	}

	return nil
}
