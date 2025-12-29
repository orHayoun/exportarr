package config

import (
	"testing"

	base_config "github.com/onedr0p/exportarr/internal/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func testFlagSet() *pflag.FlagSet {
	ret := pflag.NewFlagSet("test", pflag.ContinueOnError)
	RegisterSabnzbdFlags(ret)
	return ret
}

func TestLoadConfig_Defaults(t *testing.T) {
	flags := testFlagSet()
	c := base_config.Config{
		URL:              "http://localhost:8080",
		ApiKey:           "abcdef0123456789abcdef0123456789",
		DisableSSLVerify: true,
	}

	require := require.New(t)

	config, err := LoadSabnzbdConfig(c, flags)
	require.NoError(err)

	// base config values are not overwritten
	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
	require.True(config.DisableSSLVerify)
}

func TestLoadConfig_Environment(t *testing.T) {
	flags := testFlagSet()
	c := base_config.Config{
		URL:              "http://localhost:8080",
		ApiKey:           "abcdef0123456789abcdef0123456789",
		DisableSSLVerify: true,
	}

	require := require.New(t)
	t.Setenv("SAB_CONFIG", "test_fixtures/sabnzbd.ini")

	config, err := LoadSabnzbdConfig(c, flags)
	require.NoError(err)

	require.Equal("test_fixtures/sabnzbd.ini", config.INIConfig)

	// base config values are not overwritten
	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
	require.True(config.DisableSSLVerify)
}

func TestLoadConfig_Flags(t *testing.T) {
	flags := testFlagSet()
	flags.Set("config", "test_fixtures/sabnzbd.ini")
	c := base_config.Config{
		URL:    "http://localhost:8080",
		ApiKey: "abcdef0123456789abcdef0123456789",
	}

	// should be overridden by flags
	t.Setenv("SAB_CONFIG", "other.ini")

	require := require.New(t)
	config, err := LoadSabnzbdConfig(c, flags)
	require.NoError(err)
	require.Equal("test_fixtures/sabnzbd.ini", config.INIConfig)

	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
}

func TestLoadConfig_INIConfig(t *testing.T) {
	flags := testFlagSet()
	flags.Set("config", "test_fixtures/sabnzbd.ini")
	c := base_config.Config{
		URL: "http://localhost",
	}

	config, err := LoadSabnzbdConfig(c, flags)

	require := require.New(t)
	require.NoError(err)

	// URL should be constructed from INI host and port
	// host is "::" which should become "localhost", port is 8080
	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
}

func TestLoadConfig_INIConfigEnv(t *testing.T) {
	flags := testFlagSet()
	t.Setenv("SAB_CONFIG", "test_fixtures/sabnzbd.ini")
	c := base_config.Config{
		URL: "http://localhost",
	}

	config, err := LoadSabnzbdConfig(c, flags)

	require := require.New(t)
	require.NoError(err)

	// URL should be constructed from INI host and port
	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
}

func TestLoadConfig_INIConfigWithBaseURL(t *testing.T) {
	flags := testFlagSet()
	flags.Set("config", "test_fixtures/sabnzbd.ini")
	c := base_config.Config{
		URL: "http://sabnzbd.example.com:9090",
	}

	config, err := LoadSabnzbdConfig(c, flags)

	require := require.New(t)
	require.NoError(err)

	// When base URL is provided, INI port should override it
	require.Equal("http://localhost:8080", config.URL)
	require.Equal("abcdef0123456789abcdef0123456789", config.ApiKey)
}

func TestValidate(t *testing.T) {
	params := []struct {
		name   string
		config *SabnzbdConfig
		valid  bool
	}{
		{
			name: "good-config",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdef0123456789abcdef0123456789",
			},
			valid: true,
		},
		{
			name: "good-api-key-32-len",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdefABCDEF0123456789abcdef0123",
			},
			valid: true,
		},
		{
			name: "good-api-key-20-len",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdefABCDEF01234567",
			},
			valid: true,
		},
		{
			name: "bad-api-key-too-short",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdef0123456789abc",
			},
			valid: false,
		},
		{
			name: "bad-api-key-too-long",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdef0123456789abcdef0123456789abcdef",
			},
			valid: false,
		},
		{
			name: "bad-api-key-invalid-chars",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "abcdef0123456789abcdef01234567-",
			},
			valid: false,
		},
		{
			name: "missing-url",
			config: &SabnzbdConfig{
				URL:    "",
				ApiKey: "abcdef0123456789abcdef0123456789",
			},
			valid: false,
		},
		{
			name: "missing-api-key",
			config: &SabnzbdConfig{
				URL:    "http://localhost:8080",
				ApiKey: "",
			},
			valid: false,
		},
	}
	for _, p := range params {
		t.Run(p.name, func(t *testing.T) {
			require := require.New(t)
			err := p.config.Validate()
			if p.valid {
				require.NoError(err)
			} else {
				require.Error(err)
			}
		})
	}
}
