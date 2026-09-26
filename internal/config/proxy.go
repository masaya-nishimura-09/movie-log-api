package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func TrustedProxies() ([]string, error) {
	raw := os.Getenv("TRUSTED_PROXIES")
	if raw == "" {
		return nil, nil
	}

	proxies := strings.Split(raw, ",")
	for _, p := range proxies {
		if net.ParseIP(p) == nil {
			if _, _, err := net.ParseCIDR(p); err != nil {
				return nil, fmt.Errorf("invalid TRUSTED_PROXIES: %w", err)
			}
		}
	}
	return proxies, nil
}
