package metrics_test

import (
	"os"
	"testing"

	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func Test_googleCloudDatastore(t *testing.T) {
	project := os.Getenv("DATASTORE_PROJECT_ID")
	if project == "" {
		t.Skip("DATASTORE_PROJECT_ID not specified for Google Cloud Datastore tests.")
	}
	ds, err := metrics.NewGoogleCloudDatastore(project, "Metrics")
	require.Nil(t, err)

	var (
		m1 = metrics.Measurement{
			Source: "test",
			Date:   time.Date(2016, 12, 17, 21, 30, 0, 0, time.UTC), // 21:30
			Value:  42.1,
		}
		m2 = metrics.Measurement{
			Source: "test",
			Date:   time.Date(2016, 12, 17, 23, 30, 0, 0, time.UTC), // 23:30
			Value:  -42.1,
		}
		m3 = metrics.Measurement{
			Source: "test",
			Date:   time.Date(2016, 12, 17, 22, 30, 0, 0, time.UTC), // 22:30
			Value:  123.4,
		}
	)

	require.Nil(t, ds.Save(m1))
	require.Nil(t, ds.Save(m2))
	require.Nil(t, ds.Save(m3))

	res1, err := ds.Load("test", time.Date(2016, 12, 17, 22, 30, 0, 0, time.UTC))
	require.Nil(t, err)
	require.EqualValues(t, []metrics.Measurement{m3, m2}, res1)

	res2, err := ds.Load("test", time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC))
	require.Nil(t, err)
	require.EqualValues(t, []metrics.Measurement{m1, m3, m2}, res2)
}
