package metrics

import "time"

type Measurement struct {
	Source string
	Date   time.Time
	Value  float64
}

// Datastore stores measurements.
type Datastore interface {
	// Load queries the measurements since the given date and returns the
	// measurements as chronologically ordered.
	Load(source string, since time.Time) ([]Measurement, error)

	// Save persists the given measurement, or updates existing one with a new value.
	Save(m Measurement) error
}
