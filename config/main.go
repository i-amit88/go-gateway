package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	UpstreamURL  string
	ProxyTimeout time.Duration
	APIKey       string
}

func LoadConfig() (Config, error) {
	portText := os.Getenv("PORT")
	upstreamURL := os.Getenv("UPSTREAM_URL")
	proxyTimeoutText := os.Getenv("PROXY_TIMEOUT")
	apiKey := os.Getenv("API_KEY")

	if portText == "" {
		portText = "8080"
	}
	if upstreamURL == "" {
		upstreamURL = "http://localhost:9090"
	}
	if proxyTimeoutText == "" {
		proxyTimeoutText = "2s"
	}
	if apiKey == "" {
		return Config{}, fmt.Errorf("API_KEY must not be empty")
	}

	portNumber, err := strconv.Atoi(portText)
	if err != nil {
		return Config{}, fmt.Errorf("invalid PORT %q: %w", portText, err)
	}
	if portNumber < 1 || portNumber > 65535 {
		return Config{}, fmt.Errorf("invalid PORT %d: must be between 1 and 65535", portNumber)
	}

	proxyTimeout, err := time.ParseDuration(proxyTimeoutText)
	if err != nil {
		return Config{}, fmt.Errorf("invalid PROXY_TIMEOUT %q: %w", proxyTimeoutText, err)
	}
	if proxyTimeout <= 0 {
		return Config{}, fmt.Errorf("invalid PROXY_TIMEOUT %q: must be greater than zero", proxyTimeoutText)
	}

	return Config{
		Port:         strconv.Itoa(portNumber),
		UpstreamURL:  upstreamURL,
		ProxyTimeout: proxyTimeout,
		APIKey:       apiKey,
	}, nil
}
