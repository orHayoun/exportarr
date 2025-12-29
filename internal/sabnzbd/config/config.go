// Package config provides configuration loading for Sabnzbd exporter.
// It supports loading configuration from defaults, environment variables,
// command-line flags, and sabnzbd.ini files.
//
// The configuration is loaded in the following priority order:
// 1. Defaults
// 2. Environment variables
// 3. Command-line flags
// 4. sabnzbd.ini file (if --config flag is provided)
//
// Example usage:
//
//	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
//	config.RegisterSabnzbdFlags(flags)
//	conf := base_config.Config{
//		URL:    "http://localhost:8080",
//		ApiKey: "default-api-key",
//	}
//	cfg, err := config.LoadSabnzbdConfig(conf, flags)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := cfg.Validate(); err != nil {
//		log.Fatal(err)
//	}
package config

import (
	"strings"

	"github.com/gookit/validate"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"

	base_config "github.com/onedr0p/exportarr/internal/config"
)

// RegisterSabnzbdFlags registers command-line flags for Sabnzbd configuration.
func RegisterSabnzbdFlags(flags *flag.FlagSet) {
	flags.StringP("config", "c", "", "sabnzbd.ini config file for parsing authentication information")
}

// SabnzbdConfig holds the configuration for Sabnzbd exporter.
type SabnzbdConfig struct {
	App              string `koanf:"app"`
	INIConfig        string `koanf:"config"`
	URL              string `koanf:"url" validate:"required|url"`
	ApiKey           string `koanf:"api-key" validate:"required|regex:(^[a-zA-Z0-9]{20,32}$)"`
	DisableSSLVerify bool   `koanf:"disable-ssl-verify"`
	k                *koanf.Koanf
}

// LoadSabnzbdConfig loads Sabnzbd configuration from defaults, environment variables,
// command-line flags, and optionally from a sabnzbd.ini file.
//
// The configuration is loaded in the following priority order:
// 1. Defaults
// 2. Environment variables (converted to lowercase with dashes)
// 3. Command-line flags
// 4. sabnzbd.ini file (if --config flag is provided)
//
// Example usage:
//
//	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
//	config.RegisterSabnzbdFlags(flags)
//	conf := base_config.Config{
//		URL:    "http://localhost:8080",
//		ApiKey: "default-api-key",
//	}
//	cfg, err := config.LoadSabnzbdConfig(conf, flags)
func LoadSabnzbdConfig(conf base_config.Config, flags *flag.FlagSet) (*SabnzbdConfig, error) {
	k := koanf.New(".")

	// Defaults
	err := k.Load(confmap.Provider(map[string]interface{}{}, "."), nil)
	if err != nil {
		return nil, err
	}

	// Environment
	err = k.Load(env.Provider("", ".", func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "__", ".")
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}), nil)
	if err != nil {
		return nil, err
	}

	// Flags
	if err := k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		return nil, err
	}

	// INIConfig
	iniConfig := k.String("config")
	if iniConfig != "" {
		err := k.Load(file.Provider(iniConfig), INIParser(), koanf.WithMergeFunc(INIParser().Merge(conf.URL)))
		if err != nil {
			return nil, err
		}
	}

	out := &SabnzbdConfig{
		App:              conf.App,
		URL:              conf.URL,
		ApiKey:           conf.ApiKey,
		DisableSSLVerify: conf.DisableSSLVerify,
		k:                k,
	}
	if err = k.Unmarshal("", out); err != nil {
		return nil, err
	}
	return out, nil
}

// Validate validates the Sabnzbd configuration.
func (c *SabnzbdConfig) Validate() error {
	v := validate.Struct(c)
	if !v.Validate() {
		return v.Errors
	}
	return nil
}

// Messages returns custom validation error messages.
func (c SabnzbdConfig) Messages() map[string]string {
	return validate.MS{
		"ApiKey.regex": "api-key must be a 20-32 character alphanumeric string",
	}
}

// Translates returns field name translations for validation.
func (c SabnzbdConfig) Translates() map[string]string {
	return validate.MS{
		"INIConfig": "config",
		"ApiKey":    "api-key",
	}
}
