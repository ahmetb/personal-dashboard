package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/json"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

func main() {
	var cfg struct {
		Tasks struct {
			Foursquare struct {
				AccessToken string `toml:"access_token"`
			} `toml:"Foursquare"`
		} `toml:"tasks"`
	}

	log := task.LoggerWithTask("foursquare")
	log.Log("msg", "starting")

	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}
	task.RequireConfig(log, cfg.Tasks.Foursquare.AccessToken, "access_token")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	todayCheckins, totalCheckins, err := todayCheckins(cfg.Tasks.Foursquare.AccessToken)
	if err != nil {
		task.LogFatal(log, "error", err)
	}
	log.Log("msg", "parsed response", "checkins_today", todayCheckins, "total", totalCheckins)

	if err := store.Save(metrics.Metric{
		Name: "foursquare.checkins",
		Kind: metrics.Daily}.NewMeasurement(time.Now(), float64(todayCheckins))); err != nil {
		task.LogFatal(log, "error", err)
	}
	log.Log("msg", "saved measurement")
}

func todayCheckins(accessToken string) (today int, total int, err error) {
	epochToday := now.BeginningOfDay().Unix()
	url := fmt.Sprintf("https://api.foursquare.com/v2/users/self/checkins?oauth_token=%s&afterTimestamp=%d&v=20161219",
		accessToken, epochToday)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to query foursquare")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return 0, 0, errors.Errorf("bad status=%q from foursquare body=%s", resp.Status, string(b))
	}

	var v struct {
		Response struct {
			Checkins struct {
				Count int `json:"count"`
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			} `json:"checkins"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return 0, 0, errors.Wrap(err, "failed to decode json response")
	}
	return len(v.Response.Checkins.Items), v.Response.Checkins.Count, nil
}
