package metrics

import (
	"time"

	"github.com/jinzhu/now"
)

// Frequency bucketizes given time into another time.
type Frequency func(time.Time) time.Time

var (
	// Daily stores a single data point per day
	Daily = func(t time.Time) time.Time { return now.New(t).BeginningOfDay() }

	// Hourly stores a single data point per hour
	Hourly = func(t time.Time) time.Time { return now.New(t).BeginningOfHour() }
)

type Metric struct {
	Name string
	Kind Frequency
}

// NewMeasurement creates a measurement with given date and reduces it to a
// desired date bucket and converts its timezone to UTC.
func (m Metric) NewMeasurement(date time.Time, value float64) Measurement {
	return Measurement{
		Date:   m.Kind(date.UTC()),
		Value:  value,
		Source: m.Name,
	}
}
