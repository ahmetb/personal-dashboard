package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

type config struct {
	Tasks struct {
		Reporting struct {
			Metrics  []metric `toml:"metric"`
			Uploader map[string]map[string]string
		} `toml:"reporting"`
	} `toml:"tasks"`
}

type metric struct {
	Source      string   `toml:"source"`
	Output      string   `toml:"output"`
	DaysBack    int      `toml:"days_back"`
	Aggregators []string `toml:"aggregators"`

	aggregatorFuncs []aggregateFunc
}

func main() {
	log := task.LoggerWithTask("reporting")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	metrics, uploader, err := parseConfig()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	out := make(map[string][]record)
	log.Log("msg", "loaded configuration", "metrics", len(metrics))
	for _, m := range metrics {
		log.Log("msg", "collecting metric", "source", m.Source, "output", m.Output, "days_back", m.DaysBack, "aggregators", len(m.aggregatorFuncs))
		rs, err := readData(store, m.Source, m.DaysBack)
		if err != nil {
			task.LogFatal(log, "error", err)
		}
		rs = process(rs, m.aggregatorFuncs)
		out[m.Output] = rs
	}

	file, err := toPJSON(out)
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	if err := uploader.save(file); err != nil {
		task.LogFatal(log, "msg", "failed to upload report", "error", err)
	}
	log.Log("msg", "uploaded report")
}

// toPJSON encapsulates data in a javascript function.
func toPJSON(data map[string][]record) ([]byte, error) {
	type doc struct {
		Generated time.Time           `json:"generated"`
		Data      map[string][]record `json:"data"`
	}
	v := doc{Generated: time.Now().UTC(),
		Data: data}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert to json")
	}
	return []byte(fmt.Sprintf(`renderData(%s)`, string(b))), nil
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

func parseConfig() ([]metric, uploader, error) {
	var c config
	if err := task.ReadConfig(&c); err != nil {
		return nil, nil, err
	}

	// parse metrics
	if len(c.Tasks.Reporting.Metrics) == 0 {
		return nil, nil, errors.New("no metrics specified for reporting")
	}
	for i, m := range c.Tasks.Reporting.Metrics {
		if m.Source == "" {
			return nil, nil, errors.Errorf("source not specified for metric #%d", i)
		}
		if m.Output == "" {
			return nil, nil, errors.Errorf("output not specified for metric %s", m.Source)
		}
		if m.DaysBack <= 0 {
			return nil, nil, errors.Errorf("days_back not specified or invalid for metric '%s'", m.Source)
		}

		for _, v := range m.Aggregators {
			f, ok := aggregators[v]
			if !ok {
				return nil, nil, errors.Errorf("unknown aggregator func '%s' for metric '%s'", v, m.Source)
			}
			c.Tasks.Reporting.Metrics[i].aggregatorFuncs = append(c.Tasks.Reporting.Metrics[i].aggregatorFuncs, f)
		}
	}

	// parse uploader
	uploaders := c.Tasks.Reporting.Uploader
	if len(uploaders) != 1 {
		return nil, nil, errors.Errorf("expected only 1 uploader; got %d", len(uploaders))
	}
	var uploaderName string
	for k := range uploaders {
		uploaderName = k
	}
	uf, ok := uploadDrivers[uploaderName]
	if !ok {
		return nil, nil, errors.Errorf("unknown upload driver '%s'", uploaderName)
	}
	u, err := uf(uploaders[uploaderName])
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize uploader")
	}

	return c.Tasks.Reporting.Metrics, u, nil
}
