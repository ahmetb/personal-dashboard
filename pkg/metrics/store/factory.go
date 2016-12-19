package store

import (
	"fmt"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/pkg/errors"
)

// DatastoreFactory initializes a new datastore instance with given
// configuration parameters.
type DatastoreFactory func(parameters map[string]string) (metrics.Datastore, error)

var datastoreFactories = make(map[string]DatastoreFactory)

// Register should be called by datastore package’s init function to register
// the datastore by its name. If it is already registered, it panics.
func Register(name string, f DatastoreFactory) {
	if f == nil {
		panic("store/factory: cannot register nil datastore factory")
	}
	if _, ok := datastoreFactories[name]; ok {
		panic(fmt.Sprintf("store/factory: datastore %q is already registered", name))
	}
	datastoreFactories[name] = f
}

// Create initializes a registered datastore type with given parameters.
func Create(name string, parameters map[string]string) (metrics.Datastore, error) {
	f, ok := datastoreFactories[name]
	if !ok {
		return nil, errors.Errorf("unknown datastore: %s", name)
	}
	v, err := f(parameters)
	return v, errors.Wrap(err, "failed to create datastore")
}
