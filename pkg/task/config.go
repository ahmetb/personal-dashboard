package task

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

const (
	defaultConfigFile = "/etc/personal-dashboard/config.toml"
	configPathEnv     = "PD_CONFIG_PATH"
)

// parseTOML parses TOML-formatted config data into v.
func parseTOML(data []byte, v interface{}) error {
	_, err := toml.Decode(string(data), v)
	return errors.Wrap(err, "failed to parse config")
}

// ReadConfig reads the default config file from path (or the environment)
// variable that overrides it and parses it into v
func ReadConfig(v interface{}) error {
	b, err := ioutil.ReadFile(configPath())
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}
	return ParseTOML(b, v)
}

// configPath returns the path for the expected config file, allows it to
// be overriden with an environment variable.
func configPath() string {
	env := os.Getenv(configPathEnv)
	if env == "" {
		return defaultConfigFile
	}
	return env
}

// Requireconfig checks for the given val, exits with error if empty.
func RequireConfig(logger *log.Context, val, name string) {
	if val == "" {
		LogFatal(logger, "error", fmt.Sprintf("config key '%s' not configured", name))
	}
}
