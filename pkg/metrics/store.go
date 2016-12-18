package metrics

import "time"

type Measurement struct {
	Source string    `datastore:"source"`
	Date   time.Time `datastore:"date"`
	Value  float64   `datastore:"value,noindex"`
}

type DataStore interface {
	// Load queries the measurements since the given date and returns the
	// measurements as chronologically ordered.
	Load(source string, since time.Time) ([]Measurement, error)

	// Save persists the given measurement, or updates existing one with a new value.
	Save(m Measurement) error
}
