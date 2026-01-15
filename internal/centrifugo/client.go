package centrifugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Client struct {
	apiURL    string
	apiKey    string
	secretKey string
}

func NewClient(apiURL, apiKey, secretKey string) *Client {
	return &Client{
		apiURL:    apiURL,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

func (c *Client) GenerateConnectionToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24*time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.secretKey))
}

func (c *Client) Publish(channel string, data interface{}) error {
	payload := map[string]interface{}{
		"channel": channel,
		"data":    data,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.apiURL+"/api/publish", bytes.NewBuffer(body))
	req.Header.Set("authorization", "apikey "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[Centrifugo] Publish error: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("[Centrifugo] Publish failed: %d - %s\n", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("Centrifugo error: %d", resp.StatusCode)
	}
	fmt.Printf("[Centrifugo] Published to %s\n", channel)
	return nil
}

type CommentEvent struct {
	Type    string      `json:"type"`
	Comment interface{} `json:"comment,omitempty"`
	ID      string      `json:"id,omitempty"`
}