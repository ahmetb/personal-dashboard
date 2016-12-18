package google

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

func NewDatastore(project, kind string) (metrics.DataStore, error) {
	cl, err := datastore.NewClient(context.Background(), project)
	return &googleCloudDatastore{cl, kind}, errors.Wrap(err, "cannot initialize cloud datastore client")
}

type googleCloudDatastore struct {
	cl   *datastore.Client
	kind string
}

func (s *googleCloudDatastore) Save(m metrics.Measurement) error {
	_, err := s.cl.Put(context.Background(), datastore.NameKey(s.kind, key(m), nil), &m)
	return errors.Wrap(err, "google: failed to upsert")
}

func (s *googleCloudDatastore) Load(source string, since time.Time) ([]metrics.Measurement, error) {
	var out []metrics.Measurement
	query := datastore.NewQuery(s.kind).
		Filter("source =", source).
		Filter("date >=", since).
		Order("date")
	it := s.cl.Run(context.Background(), query)
	for {
		var m metrics.Measurement
		_, err := it.Next(&m)
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, errors.Wrap(err, "google: failed to query")
		}
		m.Date = m.Date.UTC() // GCD returns times in PST
		out = append(out, m)
	}
	return out, nil
}

// key gives a unique key to the measurement to be used in upserts.
func key(d metrics.Measurement) string {
	return fmt.Sprintf("%s@%s", d.Source, d.Date.UTC().Format(time.RFC3339))
}
