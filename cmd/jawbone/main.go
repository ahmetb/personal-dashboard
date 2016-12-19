package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"net/http"

	"io/ioutil"

	"encoding/json"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	logger "github.com/go-kit/kit/log"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

type config struct {
	Tasks struct {
		Jawbone struct {
			AccessToken string `toml:"access_token"`
		} `toml:"jawbone"`
	} `toml:"tasks"`
}

type subcmd func(log *logger.Context, c config, daysBack int, store metrics.Datastore) error

var (
	subcmds = map[string]subcmd{
		"sleeps": sleeps,
	}
)

func main() {
	var cfg config

	log := task.LoggerWithTask("jawbone")
	log.Log("msg", "starting")

	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	if len(os.Args) < 3 {
		task.LogFatal(log, "msg", "insufficient arguments", "usage", fmt.Sprintf("%s <COMMAND> <DAYS_BACK>", os.Args[0]))
	} else if len(os.Args) > 3 {
		task.LogFatal(log, "msg", "too many arguments")
	}
	cmdS := os.Args[1]
	daysBackS := os.Args[2]

	cmd, ok := subcmds[cmdS]
	if !ok {
		task.LogFatal(log, "msg", "unknown subcommand", "cmd", cmdS)
	}

	daysBack, err := strconv.Atoi(daysBackS)
	if err != nil {
		task.LogFatal(log, "msg", "failed to parse integer", "val", daysBackS, "error", err)
	}

	if err := cmd(log.With("cmd", cmdS), cfg, daysBack, store); err != nil {

	}
}

func sleeps(log *logger.Context, c config, daysBack int, store metrics.Datastore) error {
	resp, err := doRequest("https://jawbone.com/nudge/api/users/@me/sleeps", c.Tasks.Jawbone.AccessToken)
	if err != nil {
		return err
	}

	var v struct {
		Data struct {
			Sleeps []struct {
				Date    int `json:"date"`
				Details struct {
					AsleepTime uint64 `json:"asleep_time"`
					AwakeTime  uint64 `json:"awake_time"`
				} `json:"details"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &v); err != nil {
		{
			return errors.Wrap(err, "failed to parse the response")
		}
	}

	metric := metrics.Metric{
		Name: "sleeps",
		Kind: metrics.Daily}

	for i := 0; i < daysBack; i++ {
		day := now.New(time.Now()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(i))
		dayInt, _ := strconv.Atoi(fmt.Sprintf("%d%02d%02d", day.Year(), day.Month(), day.Day()))

		totalMins := 0.0
		for _, sleep := range v.Data.Sleeps { // filter sleeps by today
			if sleep.Date == dayInt {
				totalMins = float64(sleep.Details.AwakeTime-sleep.Details.AsleepTime) / 60
			}
		}

		if totalMins == 0 {
			log.Log("msg", "no sleeps found", "day", dayInt)
		} else {
			if err := store.Save(metric.NewMeasurement(day, totalMins)); err != nil { // report minutes
				return errors.Wrap(err, "failed to save measurement")
			}
			log.Log("msg", "saved sleep measurement", "day", dayInt, "hours", totalMins/60)
		}
	}
	return nil
}

func doRequest(url, accessToken string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("bad status=%q body=%s", resp.Status, string(b))
	}
	return b, nil
}
