package config_test

import (
	"testing"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/config"
	"github.com/stretchr/testify/require"
)

func Test_parseTOML(t *testing.T) {
	type vv struct {
		Name string `toml:"name"`
		Age  int    `toml:"age"`
	}
	var v vv
	err := config.Parse([]byte(`# toml test document
name="Rob Pike"
age = 42`), &v)
	require.Nil(t, err)
	require.EqualValues(t, vv{"Rob Pike", 42}, v)
}

func Test_parseTOML_error(t *testing.T) {
	type vv struct {
		Name string `toml:"name"`
		Age  int    `toml:"age"`
	}
	var v vv
	err := config.Parse([]byte(`
name=Rob Pike
age = 42`), &v)
	require.NotNil(t, err)
}
