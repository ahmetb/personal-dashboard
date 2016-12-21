package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/pkg/errors"
)

func main() {
	var cfg struct {
		Tasks struct {
			WakaTime struct {
				APIKey   string `toml:"api_key"`
				Timezone string `toml:"timezone"`
			} `toml:"wakatime"`
		} `toml:"tasks"`
	}

	log := task.LoggerWithTask("wakatime")
	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}
	task.RequireConfig(log, cfg.Tasks.WakaTime.APIKey, "api_key")
	task.RequireConfig(log, cfg.Tasks.WakaTime.Timezone, "timezone")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	loc, err := time.LoadLocation(cfg.Tasks.WakaTime.Timezone)
	if err != nil {
		task.LogFatal(log, "msg", "timezone not found", "error", err)
	}

	now := time.Now().In(loc)
	today := fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())

	log.Log("msg", "retrieving activities", "day", today, "tz", cfg.Tasks.WakaTime.Timezone)
	v, err := getActivities(cfg.Tasks.WakaTime.APIKey, today)
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	totalSecs := 0.0
	for _, activity := range v.Data {
		totalSecs += activity.Duration
	}
	totalMins := totalSecs / 60
	log.Log("msg", "retrieved activities", "durations", len(v.Data), "total_mins", totalMins)

	if err := store.Save(metrics.Metric{
		Name: "coding",
		Kind: metrics.Daily,
	}.NewMeasurement(now, totalMins)); err != nil {
		task.LogFatal(log, "msg", "failed to save measurement", "error", err)
	}
	log.Log("msg", "saved measurement")
}

type activities struct {
	Data []struct {
		Duration float64 `json:"duration"` // seconds
	} `json:"data"`
}

func getActivities(apiKey, date string) (*activities, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("https://wakatime.com/api/v1/users/current/durations?date=%s&api_key=%s", date, apiKey))
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status=%q body='%s'", resp.Status, string(b))
	}
	var v activities
	err = json.Unmarshal(b, &v)
	return &v, errors.Wrap(err, "failed to parse response")
}
