package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/trebuhs/asa-cli/internal/config"
)

const (
	tokenURL    = "https://appleid.apple.com/auth/oauth2/token"
	tokenAud    = "https://appleid.apple.com"
	tokenScope  = "searchadsorg"
	jwtLifetime = 180 * 24 * time.Hour // 180 days max
)

type TokenCache struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TokenProvider struct {
	cfg   *config.Config
	mu    sync.Mutex
	token *TokenCache
}

func NewTokenProvider(cfg *config.Config) *TokenProvider {
	return &TokenProvider{cfg: cfg}
}

func (tp *TokenProvider) GetToken() (string, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Try loading from cache
	if tp.token == nil {
		tp.token = loadCachedToken()
	}

	// Return cached token if still valid (with 5 min buffer)
	if tp.token != nil && time.Now().Add(5*time.Minute).Before(tp.token.ExpiresAt) {
		return tp.token.AccessToken, nil
	}

	// Generate new token
	token, err := tp.exchangeToken()
	if err != nil {
		return "", err
	}

	tp.token = token
	saveCachedToken(token)
	return token.AccessToken, nil
}

func (tp *TokenProvider) exchangeToken() (*TokenCache, error) {
	clientSecret, err := tp.generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("generating client secret: %w", err)
	}

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {tp.cfg.ClientID},
		"client_secret": {clientSecret},
		"scope":         {tokenScope},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error without leaking full response body
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("token exchange failed (HTTP %d)", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return &TokenCache{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

func (tp *TokenProvider) generateClientSecret() (string, error) {
	key, err := loadPrivateKey(tp.cfg.PrivateKeyPath)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    tp.cfg.TeamID,
		Subject:   tp.cfg.ClientID,
		Audience:  jwt.ClaimStrings{tokenAud},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(jwtLifetime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = tp.cfg.KeyID

	return token.SignedString(key)
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in private key file")
	}

	// Try PKCS#8 first
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if ecKey, ok := key.(*ecdsa.PrivateKey); ok {
			return ecKey, nil
		}
		return nil, fmt.Errorf("PKCS#8 key is not ECDSA")
	}

	// Try SEC1/EC format
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("unable to parse private key (tried PKCS#8 and SEC1 formats)")
}

func cachePath() string {
	return filepath.Join(config.ConfigDir(), "token_cache.json")
}

func loadCachedToken() *TokenCache {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil
	}
	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}
	return &cache
}

func saveCachedToken(token *TokenCache) {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(cachePath()), 0700)
	_ = os.WriteFile(cachePath(), data, 0600)
}

func ValidateConfig(cfg *config.Config) error {
	var missing []string
	if cfg.ClientID == "" {
		missing = append(missing, "client_id")
	}
	if cfg.TeamID == "" {
		missing = append(missing, "team_id")
	}
	if cfg.KeyID == "" {
		missing = append(missing, "key_id")
	}
	if cfg.PrivateKeyPath == "" {
		missing = append(missing, "private_key_path")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s\nRun 'asa-cli configure' to set up credentials", strings.Join(missing, ", "))
	}

	// Validate key file exists
	if _, err := os.Stat(cfg.PrivateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("private key file not found: %s", cfg.PrivateKeyPath)
	}

	return nil
}
