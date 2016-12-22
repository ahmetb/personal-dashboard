package task

import (
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics/store"
	"github.com/pkg/errors"

	_ "github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics/store/google" // register
)

// GetDatastore reads the configuration file and initializes a datastore
// with its configuration.
func GetDatastore() (metrics.Datastore, error) {
	var v struct {
		Datastores map[string]map[string]string `toml:"datastore"`
	}
	err := ReadConfig(&v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get datastore config")
	}
	if len(v.Datastores) == 0 {
		return nil, errors.New("no datastore specified in configuration")
	}
	if len(v.Datastores) > 1 {
		return nil, errors.Errorf("multiple (%d) datastores specified", len(v.Datastores))
	}

	// there is only one datastore
	var s string
	for k := range v.Datastores {
		s = k
	}

	return store.Create(s, v.Datastores[s])
}
