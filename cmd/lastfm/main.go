package main

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
)

type config struct {
	Tasks struct {
		LastFM struct {
			APIKey string `toml:"api_key"`
			User   string `toml:"user"`
		} `toml:"lastfm"`
	} `toml:"tasks"`
}

func main() {
	var cfg config
	log := task.LoggerWithTask("lastfm")
	log.Log("msg", "starting")

	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}
	task.RequireConfig(log, cfg.Tasks.LastFM.APIKey, "api_key")
	task.RequireConfig(log, cfg.Tasks.LastFM.User, "user")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	epochToday := now.BeginningOfDay().Unix()
	url := fmt.Sprintf("http://ws.audioscrobbler.com/2.0?format=json&from=%d&method=user.getrecenttracks&user=%s&api_key=%s&limit=200",
		epochToday, cfg.Tasks.LastFM.User, cfg.Tasks.LastFM.APIKey)

	log.Log("msg", "querying last.fm API")
	resp, err := http.Get(url)
	if err != nil {
		task.LogFatal(log, "msg", "error querying last.fm", "error")
	}
	defer resp.Body.Close()

	var v struct {
		RecentTracks struct {
			Track []struct {
				Name string `json:"name"`
			} `json:"track"`
		} `json:"recenttracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		task.LogFatal(log, "msg", "failed to decode json response", "error", err)
	}
	log.Log("msg", "parsed response", "songs", len(v.RecentTracks.Track))

	if err := store.Save(metrics.Metric{
		Name: "tracks_listened",
		Kind: metrics.Daily,
	}.NewMeasurement(time.Now(), float64(len(v.RecentTracks.Track)))); err != nil {
		task.LogFatal(log, "error", err)
	}

	log.Log("msg", "saved measurement")
}
