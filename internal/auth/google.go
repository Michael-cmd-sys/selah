package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GoogleTokenInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Aud     string `json:"aud"`
}

func VerifyGoogleIDToken(idToken string) (*GoogleTokenInfo, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google tokeninfo returned status %d", resp.StatusCode)
	}

	var info GoogleTokenInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse tokeninfo: %w", err)
	}
	if info.Sub == "" {
		return nil, fmt.Errorf("missing sub in google token")
	}
	return &info, nil
}
