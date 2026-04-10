package auth

import (
	"fmt"
	"strings"
)

func ExtractBearerToken(header string) (string, error) {
	if strings.TrimSpace(header) == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization scheme")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("empty bearer token")
	}

	return token, nil
}