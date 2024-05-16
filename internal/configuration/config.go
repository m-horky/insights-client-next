package configuration

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/ini/v2"
	"github.com/m-horky/insights-client-next/internal/constants"
)

var configurations = make(map[string]Configuration)

// ClearCache deletes cached configurations.
func ClearCache() {
	configurations = make(map[string]Configuration)
}

type Configuration struct {
	APIProtocol         string        `config:"api_protocol"`
	APIHost             string        `config:"api_host"`
	APIPort             uint64        `config:"api_port"`
	HTTPTimeout         time.Duration `config:"http_timeout"`
	LogLevel            slog.Level    `config:"loglevel"`
	IdentityCertificate string        `config:"identity_certificate"`
	IdentityKey         string        `config:"identity_key"`
	CACertificate       string        `config:"ca_certificate"`
}

// update overwrites existing values with new ones.
func (c *Configuration) update(data map[string]string) {
	if value, ok := data["api_protocol"]; ok {
		c.APIProtocol = value
	}
	if value, ok := data["api_host"]; ok {
		c.APIHost = value
	}
	if value, ok := data["api_port"]; ok {
		if number, err := strconv.ParseUint(value, 10, 64); err == nil {
			c.APIPort = number
		}
	}
	if value, ok := data["http_timeout"]; ok {
		if number, err := strconv.ParseUint(value, 10, 64); err == nil {
			c.HTTPTimeout = time.Duration(number) * time.Second
		}
	}
	if value, ok := data["loglevel"]; ok {
		switch strings.ToLower(value) {
		case "debug":
			c.LogLevel = slog.LevelDebug
		case "info":
			c.LogLevel = slog.LevelInfo
		case "warning":
			c.LogLevel = slog.LevelWarn
		default:
			c.LogLevel = slog.LevelError
		}
	}
	if value, ok := data["identity_certificate"]; ok {
		c.IdentityCertificate = value
	}
	if value, ok := data["identity_key"]; ok {
		c.IdentityKey = value
	}
	if value, ok := data["ca_certificate"]; ok {
		c.CACertificate = value
	}
}

func GetDefaultConfiguration() Configuration {
	return Configuration{
		APIProtocol:         "https",
		APIHost:             "cert.cloud.redhat.com",
		APIPort:             443,
		HTTPTimeout:         120 * time.Second,
		LogLevel:            slog.LevelDebug,
		IdentityCertificate: "/etc/pki/consumer/cert.pem",
		IdentityKey:         "/etc/pki/consumer/key.pem",
		CACertificate:       "/etc/rhsm/ca/redhat-ep.pem",
	}
}

// GetConfiguration loads configuration from default path.
func GetConfiguration() Configuration {
	path := constants.ConfigPath
	return GetConfigurationFromPath(path)
}

// GetConfigurationFromPath loads configuration from custom path.
//
// It caches its value internally, so it can be called multiple times
// with no overhead. Call ClearCache before if you don't want to do that.
func GetConfigurationFromPath(path string) Configuration {
	if cached, ok := configurations[path]; ok {
		return cached
	}

	paths := expandConfigurationPath(path)

	config := GetDefaultConfiguration()
	for _, file := range paths {
		err := ini.LoadFiles(file)
		if err != nil {
			slog.Warn(
				"ignoring malformed file",
				slog.String("file", file),
				slog.Any("error", err),
			)
			continue
		}

		config.update(ini.Data()["insights-client"])
	}

	configurations[path] = config
	return config
}

// expandConfigurationPath attempts to expand `path` into all files under `path.d/`.
func expandConfigurationPath(path string) []string {
	dir := fmt.Sprintf("%s.d", path)
	stat, err := os.Stat(dir)
	if err != nil {
		return []string{path}
	}
	if !stat.IsDir() {
		return []string{path}
	}

	paths, err := os.ReadDir(dir)
	if err != nil {
		slog.Warn(
			"could not list configuration files",
			slog.String("path", dir),
			slog.Any("error", err),
		)
	}
	var result = make([]string, len(paths)+1)
	result[0] = path
	for i, path := range paths {
		result[i+1] = filepath.Join(dir, path.Name())
	}
	return result
}
