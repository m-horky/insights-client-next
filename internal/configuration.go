package internal

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/ini/v2"
)

var configurationInitialized bool = false
var cachedConfiguration Configuration = getDefaultConfiguration()

// ClearConfiguration clears the configuration cached in memory.
func ClearConfiguration() {
	configurationInitialized = false
}

type Configuration struct {
	APIProtocol         string        `config:"api_protocol"`
	APIHost             string        `config:"api_host"`
	APIPort             uint          `config:"api_port"`
	HTTPTimeout         time.Duration `config:"http_timeout"`
	LogLevel            slog.Level    `config:"loglevel"`
	IdentityCertificate string        `config:"identity_certificate"`
	IdentityKey         string        `config:"identity_key"`
	CACertificate       string        `config:"ca_certificate"`
}

// update in-place updates the values of the configuration.
func (c *Configuration) update(data map[string]string) {
	for key, value := range data {
		switch key {
		case "api_protocol":
			c.APIProtocol = value
		case "api_host":
			c.APIHost = value
		case "api_port":
			if number, err := strconv.ParseUint(value, 10, 32); err == nil {
				c.APIPort = uint(number)
			} else {
				slog.Warn("ignoring malformed API port", slog.String("value", value))
			}
		case "loglevel":
			switch strings.ToLower(value) {
			case "debug":
				c.LogLevel = slog.LevelDebug
			case "info":
				c.LogLevel = slog.LevelInfo
			case "warn":
				c.LogLevel = slog.LevelWarn
			default:
				c.LogLevel = slog.LevelError
			}
		case "identity_certificate":
			c.IdentityCertificate = value
		case "identity_key":
			c.IdentityKey = value
		case "ca_certificate":
			c.CACertificate = value
		}
	}
}

// GetConfiguration loads configuration from a filesystem.
//
// It caches its value internally, so it can be called multiple times with no overhead.
// Call ClearConfiguration to force reload.
func GetConfiguration() Configuration {
	if configurationInitialized {
		return cachedConfiguration
	}

	config := getConfigurationFromPath(ConfigPath)
	cachedConfiguration = config
	configurationInitialized = true
	return cachedConfiguration
}

// getDefaultConfiguration loads sane defaults.
func getDefaultConfiguration() Configuration {
	return Configuration{
		APIProtocol:         "https",
		APIHost:             "cert.console.redhat.com",
		APIPort:             443,
		HTTPTimeout:         10 * time.Second,
		LogLevel:            slog.LevelDebug,
		IdentityCertificate: "/etc/pki/consumer/cert.pem",
		IdentityKey:         "/etc/pki/consumer/key.pem",
		CACertificate:       "/etc/rhsm/ca/redhat-ep.pem",
	}
}

// getConfigurationFromPath loads configuration from path and its .d/ subdirectory.
func getConfigurationFromPath(path string) Configuration {
	config := getDefaultConfiguration()
	for _, file := range expandConfigurationPath(path) {
		err := ini.LoadFiles(file)
		if err != nil {
			slog.Warn("ignoring malformed file", slog.String("file", file), slog.String("error", err.Error()))
			continue
		}
		config.update(ini.Data()["insights-client"])
	}
	return config
}

// expandConfigurationPath expands `.../path.conf` to also include all files under `.../path.conf.d/`.
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
		slog.Warn("could not list configuration files", slog.String("error", err.Error()))
		return []string{path}
	}

	var result = make([]string, len(paths)+1)
	result[0] = path
	for i, file := range paths {
		result[i+1] = filepath.Join(dir, file.Name())
	}
	return result
}
