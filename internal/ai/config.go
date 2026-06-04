package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds API keys for all supported providers.
type Config struct {
	OpenAIKey    string `json:"openai_api_key"`
	AnthropicKey string `json:"anthropic_api_key"`
	GeminiKey    string `json:"gemini_api_key"`
	AWSRegion    string `json:"aws_region"` // for Kiro/Bedrock
}

// LoadConfig loads API keys from environment variables, falling back to
// ~/.config/iqtoolkit-analyzer/config.json if env vars are not set.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		OpenAIKey:    os.Getenv("OPENAI_API_KEY"),
		AnthropicKey: os.Getenv("ANTHROPIC_API_KEY"),
		GeminiKey:    os.Getenv("GEMINI_API_KEY"),
		AWSRegion:    os.Getenv("AWS_REGION"),
	}

	// If any key is already set via env, return early.
	if cfg.OpenAIKey != "" || cfg.AnthropicKey != "" || cfg.GeminiKey != "" {
		return cfg, nil
	}

	// Try config file.
	path := configPath()
	f, err := os.Open(path)
	if err != nil {
		return cfg, nil // file missing is not an error
	}
	defer f.Close()

	var fileCfg Config
	if err := json.NewDecoder(f).Decode(&fileCfg); err != nil {
		return nil, err
	}

	// File values fill in blanks only.
	if cfg.OpenAIKey == "" {
		cfg.OpenAIKey = fileCfg.OpenAIKey
	}
	if cfg.AnthropicKey == "" {
		cfg.AnthropicKey = fileCfg.AnthropicKey
	}
	if cfg.GeminiKey == "" {
		cfg.GeminiKey = fileCfg.GeminiKey
	}
	if cfg.AWSRegion == "" {
		cfg.AWSRegion = fileCfg.AWSRegion
	}
	return cfg, nil
}

// ClientFromConfig creates a Client for the given provider using loaded config.
func ClientFromConfig(provider Provider) (*Client, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	var key string
	switch provider {
	case OpenAI:
		key = cfg.OpenAIKey
	case Anthropic:
		key = cfg.AnthropicKey
	case Gemini:
		key = cfg.GeminiKey
	case Kiro:
		// Kiro uses AWS credentials, not an API key
	}
	c := NewClient(provider, key)
	if cfg.AWSRegion != "" {
		c.Region = cfg.AWSRegion
	}
	return c, nil
}

func configPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "iqtoolkit-analyzer", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "iqtoolkit-analyzer", "config.json")
}
