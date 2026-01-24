package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type OpenAIClient struct {
	apiKey string
	client *http.Client
}

func NewOpenAIClient() *OpenAIClient {
	fmt.Println("Creating OpenAI client")
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("WARNING: OPENAI_API_KEY is not set")
	} else {
		// Log masked key for debugging
		masked := ""
		if len(key) > 8 {
			masked = key[:4] + "..." + key[len(key)-4:]
		} else {
			masked = "***"
		}
		fmt.Printf("OPENAI_API_KEY is set (%s)\n", masked)
	}
	return &OpenAIClient{
		apiKey: key,
		client: &http.Client{},
	}
}

type ChatRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	ResponseFormat ResponseFormat `json:"response_format"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (g *OpenAIClient) Translate(baseLang, targetLang string, texts map[string]string) (map[string]string, error) {
	if g.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is not set")
	}

	// Only send keys that have values
	toTranslate := make(map[string]string)
	for k, v := range texts {
		if v != "" {
			toTranslate[k] = v
		}
	}

	if len(toTranslate) == 0 {
		return nil, fmt.Errorf("no texts to translate")
	}

	prompt := fmt.Sprintf(`You are a professional translator. Translate the following JSON key-value pairs from %s to %s. 
Return ONLY valid JSON with the same keys and translated values. Do not translate the keys.
Preserve any placeholders like {name}, {count}, etc. as is.`, baseLang, targetLang)

	inputJSON, err := json.Marshal(toTranslate)
	if err != nil {
		return nil, err
	}

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: string(inputJSON)},
		},
		ResponseFormat: ResponseFormat{Type: "json_object"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		fmt.Printf("OpenAI API Error Body: %s\n", errBody.String())
		return nil, fmt.Errorf("openai api error: status %d - %s", resp.StatusCode, errBody.String())
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no translation returned")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(chatResp.Choices[0].Message.Content), &result); err != nil {
		return nil, err
	}

	return result, nil
}
