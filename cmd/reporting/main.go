package main

import (
	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

type config struct {
	Tasks struct {
		Reporting struct {
			Metrics []struct {
				Source      string   `toml:"source"`
				Output      string   `toml:"output"`
				DaysBack    int      `toml:"days_back"`
				Aggregators []string `toml:"aggregators"`

				aggregatorFuncs []aggregateFunc
			} `toml:"metric"`
		} `toml:"reporting"`
	} `toml:"tasks"`
}

func main() {
	log := task.LoggerWithTask("reporting")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	c, err := parseConfig()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	metrics := c.Tasks.Reporting.Metrics
	log.Log("msg", "loaded configuration", "metrics", len(metrics))
	for _, m := range metrics {
		log.Log("msg", "collecting metric", "source", m.Source, "output", m.Output, "days_back", m.DaysBack, "aggregators", len(m.aggregatorFuncs))
		rs, err := readData(store, m.Source, m.DaysBack)
		if err != nil {
			task.LogFatal(log, "error", err)
		}
		rs = process(rs, m.aggregatorFuncs)

		for _, v := range rs {
			log.Log("key", v.Date, "data", v.Value, "fake", v.Interpolated)
		}
	}
}

func readData(store metrics.Datastore, source string, daysBack int) ([]record, error) {
	since := now.New(time.Now().UTC()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(daysBack))
	m, err := store.Load(source, since)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load the data")
	}
	out := make([]record, len(m))
	for i, v := range m {
		out[i] = record{Date: v.Date, Value: v.Value}
	}
	return out, nil
}

func process(in []record, funcs []aggregateFunc) []record {
	for _, f := range funcs {
		in = f(in)
	}
	return in
}

func parseConfig() (*config, error) {
	var c config
	if err := task.ReadConfig(&c); err != nil {
		return nil, err
	}

	if len(c.Tasks.Reporting.Metrics) == 0 {
		return nil, errors.New("no metrics specified for reporting")
	}
	for i, m := range c.Tasks.Reporting.Metrics {
		if m.Source == "" {
			return nil, errors.Errorf("source not specified for metric #%d", i)
		}
		if m.Output == "" {
			return nil, errors.Errorf("output not specified for metric %s", m.Source)
		}
		if m.DaysBack <= 0 {
			return nil, errors.Errorf("days_back not specified or invalid for metric '%s'", m.Source)
		}

		for _, v := range m.Aggregators {
			f, ok := aggregators[v]
			if !ok {
				return nil, errors.Errorf("unknown aggregator func '%s' for metric '%s'", v, m.Source)
			}
			c.Tasks.Reporting.Metrics[i].aggregatorFuncs = append(c.Tasks.Reporting.Metrics[i].aggregatorFuncs, f)
		}
	}
	return &c, nil
}
