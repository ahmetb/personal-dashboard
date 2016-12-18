package metrics

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

func NewGoogleCloudDatastore(project, kind string) (DataStore, error) {
	cl, err := datastore.NewClient(context.Background(), project)
	return &googleCloudDatastore{cl, kind}, errors.Wrap(err, "cannot initialize cloud datastore client")
}

type googleCloudDatastore struct {
	cl   *datastore.Client
	kind string
}

func (s *googleCloudDatastore) Save(m Measurement) error {
	_, err := s.cl.Put(context.Background(), datastore.NameKey(s.kind, m.key(), nil), &m)
	return errors.Wrap(err, "google: failed to upsert")
}

func (s *googleCloudDatastore) Load(source string, since time.Time) ([]Measurement, error) {
	var out []Measurement
	query := datastore.NewQuery(s.kind).
		Filter("source =", source).
		Filter("date >=", since).
		Order("date")
	it := s.cl.Run(context.Background(), query)
	for {
		var m Measurement
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
