package config

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// INI represents the INI parser helper for sabnzbd.ini files.
type INI struct{}

// INIParser returns a new INI parser instance.
func INIParser() *INI {
	return &INI{}
}

// Unmarshal parses INI file content and returns a map of configuration values.
// It handles sabnzbd.ini format with sections like [misc] and key-value pairs.
func (p *INI) Unmarshal(b []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	currentSection := "misc" // Default section for sabnzbd.ini
	
	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		
		// Skip version and encoding headers
		if strings.Contains(line, "sabnzbd.ini_version__") || strings.Contains(line, "__encoding__") {
			continue
		}
		
		// Check for section header [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}
		
		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		value = strings.Trim(value, `"`)
		
		// Store as section.key
		fullKey := currentSection + "." + key
		result[fullKey] = value
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning INI file: %w", err)
	}
	
	return result, nil
}

// Marshal is not implemented for INI parser (read-only).
func (p *INI) Marshal(o map[string]interface{}) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Merge returns a merge function that constructs the URL from host/port fields
// and extracts the API key from the INI configuration.
// The INI parser will create nested keys like "misc.api_key", "misc.host", "misc.port".
func (p *INI) Merge(baseURL string) func(src, dest map[string]interface{}) error {
	return func(src, dest map[string]interface{}) error {
		// Extract API key from misc.api_key
		if apiKey, ok := getStringFromMap(src, "misc.api_key"); ok && apiKey != "" {
			dest["api-key"] = apiKey
		}

		// Extract host and port from misc section
		host, hostOk := getStringFromMap(src, "misc.host")
		port, portOk := getStringFromMap(src, "misc.port")
		enableHTTPS, httpsOk := getStringFromMap(src, "misc.enable_https")

		// If we have host and port, construct/modify URL
		if hostOk && portOk && host != "" && port != "" {
			// Determine protocol based on enable_https
			protocol := "http"
			if httpsOk && enableHTTPS == "1" {
				protocol = "https"
				// Check if https_port is specified
				if httpsPort, ok := getStringFromMap(src, "misc.https_port"); ok && httpsPort != "" {
					port = httpsPort
				}
			}

			// Handle IPv6 addresses (host might be "::" or "[::]")
			host = strings.TrimSpace(host)
			// Convert "::" or empty to localhost first, before IPv6 handling
			if host == "::" || host == "" {
				host = "localhost"
			} else if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
				// Wrap IPv6 addresses in brackets
				host = "[" + host + "]"
			}

			// Start with baseURL if provided, otherwise construct from scratch
			var u *url.URL
			var err error
			if baseURL != "" {
				u, err = url.Parse(baseURL)
				if err != nil {
					return fmt.Errorf("failed to parse base URL: %w", err)
				}
				// Override host and port from INI
				u.Host = fmt.Sprintf("%s:%s", host, port)
				u.Scheme = protocol
			} else {
				// Construct URL from scratch
				constructedURL := fmt.Sprintf("%s://%s:%s", protocol, host, port)
				u, err = url.Parse(constructedURL)
				if err != nil {
					return fmt.Errorf("failed to parse constructed URL: %w", err)
				}
			}

			dest["url"] = u.String()
		}

		return nil
	}
}

// getStringFromMap retrieves a string value from a map, supporting both
// flat dot-notation keys (e.g., "misc.api_key") and nested maps.
func getStringFromMap(m map[string]interface{}, key string) (string, bool) {
	// First try direct access (koanf may flatten keys with dot notation)
	if val, exists := m[key]; exists {
		return convertToString(val), true
	}

	// Fall back to nested map traversal
	parts := strings.Split(key, ".")
	if len(parts) == 0 {
		return "", false
	}

	var current interface{} = m
	for i, part := range parts {
		mm, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}

		val, exists := mm[part]
		if !exists {
			return "", false
		}

		// If this is the last part, return the string value
		if i == len(parts)-1 {
			return convertToString(val), true
		}

		current = val
	}

	return "", false
}

// convertToString converts various types to string.
func convertToString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

