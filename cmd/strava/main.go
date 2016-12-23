package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"net/http"

	"encoding/json"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

type config struct {
	Tasks struct {
		Strava struct {
			AccessToken string `toml:"access_token"`
			DaysBack    int    `toml:"days_back"`
			Timezone    string `toml:"timezone"`
		} `toml:"strava"`
	} `toml:"tasks"`
}

var (
	GitSummary      string // provided by govvv
	defaultDaysBack = 3
)

func main() {
	log := task.LoggerWithTask("strava", GitSummary)
	log.Log("msg", "starting")

	var cfg config

	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}
	task.RequireConfig(log, cfg.Tasks.Strava.AccessToken, "access_token")

	userLocation := time.UTC
	if cfg.Tasks.Strava.Timezone != "" {
		loc, err := time.LoadLocation(cfg.Tasks.Strava.Timezone)
		if err != nil {
			task.LogFatal(log, "msg", "cannot load timezone", "error", err)
		}
		userLocation = loc
	}

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	// Get user's activities
	daysBack := cfg.Tasks.Strava.DaysBack
	if daysBack == 0 {
		daysBack = defaultDaysBack
	}
	log.Log("msg", "starting", "days_back", daysBack)
	today := now.New(time.Now().UTC()).BeginningOfDay()
	since := today.Add(-time.Hour * 24 * time.Duration(daysBack))

	activities, err := getActivities(cfg.Tasks.Strava.AccessToken, since)
	if err != nil {
		task.LogFatal(log, "error", err)
	}
	log.Log("msg", "loaded activities", "n", len(activities))
	for i := 0; i < daysBack; i++ {
		day := today.Add(-time.Hour * 24 * time.Duration(i))
		dayActivities := filter(activities, day, userLocation)
		elapsed := sumTime(dayActivities)
		log.Log("day", day, "activities", len(dayActivities), "elapsed", elapsed)

		// report total activities
		if err := store.Save(metrics.Metric{
			Name: "strava.activities",
			Kind: metrics.Daily}.NewMeasurement(day, float64(len(dayActivities)))); err != nil {
			task.LogFatal(log, "error", err)
		}

		// report activity duration
		if err := store.Save(metrics.Metric{
			Name: "strava.exercise_time",
			Kind: metrics.Daily}.NewMeasurement(day, elapsed)); err != nil {
			task.LogFatal(log, "error", err)
		}
	}
}

type activity struct {
	ID          int       `json:"int"`
	ElapsedTime int       `json:"elapsed_time"`
	StartDate   time.Time `json:"start_date"`
}

func getActivities(accessToken string, since time.Time) ([]activity, error) {
	url := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?per_page=200&after=%d", since.Unix())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response status=%q body='%s'", resp.Status, string(b))
	}

	var v []activity
	return v, errors.Wrap(json.Unmarshal(b, &v), "failed to decode response")
}

func filter(l []activity, day time.Time, userLocation *time.Location) []activity {
	var out []activity
	for _, v := range l {
		activityDateLocal := v.StartDate.In(userLocation)
		activityDay := activityDateLocal.Truncate(time.Hour * 24)
		if activityDay.Equal(day) {
			out = append(out, v)
		}
	}
	return out
}

func sumTime(l []activity) float64 {
	var s float64
	for _, v := range l {
		s += float64(v.ElapsedTime)
	}
	return s
}
