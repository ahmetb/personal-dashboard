package google

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics/store"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// measurement is an internal type
type measurement struct {
	Source string    `datastore:"source"`
	Date   time.Time `datastore:"date"`
	Value  float64   `datastore:"value,noindex"`
}

func init() {
	store.Register("googledatastore", newFromParameters)
}

func newFromParameters(params map[string]string) (metrics.Datastore, error) {
	project := params["project"]
	if project == "" {
		return nil, errors.New("google: 'project' parameter not specified")
	}
	kind := params["kind"]
	if kind == "" {
		return nil, errors.New("google: 'kind' parameter not specified")
	}
	return newDatastore(project, kind)
}

func newDatastore(project, kind string) (metrics.Datastore, error) {
	cl, err := datastore.NewClient(context.Background(), project)
	return &googleCloudDatastore{cl, kind}, errors.Wrap(err, "google: cannot initialize cloud datastore client")
}

type googleCloudDatastore struct {
	cl   *datastore.Client
	kind string
}

func (s *googleCloudDatastore) Save(m metrics.Measurement) error {
	// convert to datastore-annotated type
	v := measurement{
		Source: m.Source,
		Date:   m.Date,
		Value:  m.Value,
	}
	_, err := s.cl.Put(context.Background(), datastore.NameKey(s.kind, key(v), nil), &v)
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
func key(d measurement) string {
	return fmt.Sprintf("%s@%s", d.Source, d.Date.UTC().Format(time.RFC3339))
}
