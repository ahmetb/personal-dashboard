package metrics_test

import (
	"testing"
	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestDaily(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.Nil(t, err)
	v := time.Date(2016, 12, 17, 20, 34, 58, 1, loc)
	require.EqualValues(t, time.Date(2016, 12, 17, 0, 0, 0, 0, loc), metrics.Daily(v))
}

func TestHourly(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.Nil(t, err)
	v := time.Date(2016, 12, 17, 20, 34, 58, 1, loc)
	require.EqualValues(t, time.Date(2016, 12, 17, 20, 0, 0, 0, loc), metrics.Hourly(v))
}

func TestNewMeasurement(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.Nil(t, err)
	date := time.Date(2016, 12, 17, 12, 34, 58, 1, loc)
	m := metrics.Metric{
		Name: "some.metric",
		Kind: metrics.Hourly}.NewMeasurement(date, 64.0)

	require.EqualValues(t, metrics.Measurement{
		Source: "some.metric",
		Value:  64.0,
		Date:   time.Date(2016, 12, 17, 17, 0, 0, 0, time.UTC)}, m)
}
