package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Parse parses TOML-formatted config data into v.
func Parse(data []byte, v interface{}) error {
	_, err := toml.Decode(string(data), v)
	return errors.Wrap(err, "failed to parse config")
}
