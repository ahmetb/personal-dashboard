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
		"sleeps":    sleeps,
		"steps":     steps,
		"caffeine":  caffeine,
		"heartrate": restingHeartrate}
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

	v, err := parseJawboneResponse(resp)
	if err != nil {
		return err
	}

	metric := metrics.Metric{
		Name: "sleeps",
		Kind: metrics.Daily}

	for i := 0; i < daysBack; i++ {
		day := now.New(time.Now().UTC()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(i))
		dayInt, _ := strconv.Atoi(fmt.Sprintf("%d%02d%02d", day.Year(), day.Month(), day.Day()))

		totalMins := 0.0
		for _, sleep := range v.Data.Items { // filter sleeps by today
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

func steps(log *logger.Context, c config, daysBack int, store metrics.Datastore) error {
	resp, err := doRequest("https://jawbone.com/nudge/api/users/@me/moves", c.Tasks.Jawbone.AccessToken)
	if err != nil {
		return err
	}

	v, err := parseJawboneResponse(resp)
	if err != nil {
		return err
	}

	metric := metrics.Metric{
		Name: "steps",
		Kind: metrics.Daily}

	for i := 0; i < daysBack; i++ {
		day := now.New(time.Now().UTC()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(i))
		dayInt, _ := strconv.Atoi(fmt.Sprintf("%d%02d%02d", day.Year(), day.Month(), day.Day()))

		total := 0
		for _, move := range v.Data.Items {
			if move.Date == dayInt {
				total += move.Details.Steps
			}
		}

		if total == 0 {
			log.Log("msg", "no steps found", "day", dayInt)
		} else {
			if err := store.Save(metric.NewMeasurement(day, float64(total))); err != nil {
				return errors.Wrap(err, "failed to save measurement")
			}
			log.Log("msg", "saved step measurement", "day", dayInt, "steps", total)
		}
	}
	return nil
}

func caffeine(log *logger.Context, c config, daysBack int, store metrics.Datastore) error {
	resp, err := doRequest("https://jawbone.com/nudge/api/users/@me/meals", c.Tasks.Jawbone.AccessToken)
	if err != nil {
		return err
	}

	v, err := parseJawboneResponse(resp)
	if err != nil {
		return err
	}

	metric := metrics.Metric{
		Name: "caffeine_intake",
		Kind: metrics.Daily}

	for i := 0; i < daysBack; i++ {
		day := now.New(time.Now().UTC()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(i))
		dayInt, _ := strconv.Atoi(fmt.Sprintf("%d%02d%02d", day.Year(), day.Month(), day.Day()))

		total := 0
		for _, move := range v.Data.Items {
			if move.Date == dayInt {
				total += move.Details.Caffeine
			}
		}

		if total == 0 {
			log.Log("msg", "no caffeine found", "day", dayInt)
		} else {
			if err := store.Save(metric.NewMeasurement(day, float64(total))); err != nil {
				return errors.Wrap(err, "failed to save measurement")
			}
			log.Log("msg", "saved caffeine measurement", "day", dayInt, "caffeine_mg", total)
		}
	}
	return nil
}

func restingHeartrate(log *logger.Context, c config, daysBack int, store metrics.Datastore) error {
	resp, err := doRequest("https://jawbone.com/nudge/api/users/@me/heartrates", c.Tasks.Jawbone.AccessToken)
	if err != nil {
		return err
	}

	v, err := parseJawboneResponse(resp)
	if err != nil {
		return err
	}

	metric := metrics.Metric{
		Name: "resting_heartrate",
		Kind: metrics.Daily}

	for i := 0; i < daysBack; i++ {
		day := now.New(time.Now().UTC()).BeginningOfDay().Add(-time.Hour * 24 * time.Duration(i))
		dayInt, _ := strconv.Atoi(fmt.Sprintf("%d%02d%02d", day.Year(), day.Month(), day.Day()))

		for _, item := range v.Data.Items {
			if item.Date == dayInt {
				hr := item.RestingHeartrate
				if hr == 0 {
					log.Log("msg", "no heartrate found", "day", dayInt)
				} else {
					if err := store.Save(metric.NewMeasurement(day, float64(hr))); err != nil {
						return errors.Wrap(err, "failed to save measurement")
					}
					log.Log("msg", "saved resting heartrate measurement", "day", dayInt, "bpm", hr)
				}
			}
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

// jawboneResponse is a common type that holds information coming from various
// jawbone endpoints used.
type jawboneResponse struct {
	Data struct {
		Items []struct {
			Date    int `json:"date"`
			Details struct {
				Caffeine   int    `json:"caffeine"`
				Steps      int    `json:"steps"`
				AsleepTime uint64 `json:"asleep_time"`
				AwakeTime  uint64 `json:"awake_time"`
			} `json:"details"`
			RestingHeartrate int `json:"resting_heartrate"`
		} `json:"items"`
	} `json:"data"`
}

func parseJawboneResponse(b []byte) (jawboneResponse, error) {
	var v jawboneResponse
	err := json.Unmarshal(b, &v)
	return v, errors.Wrap(err, "failed to parse jawbone response")
}
