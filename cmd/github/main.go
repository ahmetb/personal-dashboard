package main

import (
	"time"

	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/google/go-github/github"
	"github.com/jinzhu/now"
	"golang.org/x/oauth2"
)

var (
	// events that constitute as contributon, full list: https://developer.github.com/v3/activity/events/types/
	contributionEvents = []string{
		"CommitCommentEvent",
		"CreateEvent",
		"RepositoryEvent",
		"IssuesEvent",
		"IssueCommentEvent",
		"PullRequestEvent",
		"PullRequestReviewEvent",
		"PullRequestReviewCommentEvent",
		"PushEvent"}
)

func main() {
	var cfg struct {
		Tasks struct {
			GitHub struct {
				Username    string `toml:"username"`
				AccessToken string `toml:"access_token"` // generate via  https://github.com/settings/tokens
				PublicOnly  bool   `toml:"public_only"`
			} `toml:"github"`
		} `toml:"tasks"`
	}

	log := task.LoggerWithTask("github")
	if err := task.ReadConfig(&cfg); err != nil {
		task.LogFatal(log, "error", err)
	}
	task.RequireConfig(log, cfg.Tasks.GitHub.Username, "username")
	task.RequireConfig(log, cfg.Tasks.GitHub.AccessToken, "access_token")

	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	log.Log("msg", "starting", "username", cfg.Tasks.GitHub.Username)
	gh := github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Tasks.GitHub.AccessToken})))

	page := 0
	activities := 0

	today := now.New(time.Now().UTC()).BeginningOfDay()
	for {
		var done bool
		events, _, err := gh.Activity.ListEventsPerformedByUser(cfg.Tasks.GitHub.Username, cfg.Tasks.GitHub.PublicOnly, &github.ListOptions{Page: page})
		if err != nil {
			task.LogFatal(log, "msg", "failed to get user events", "error", err)
		}
		log.Log("msg", "retrieved events", "page", page, "count", len(events))

		for _, event := range events {
			if now.New(event.CreatedAt.UTC()).BeginningOfDay().Equal(today) {
				log.Log("event", event.ID, "date", event.CreatedAt, "type", event.Type)
				activities++
			} else {
				done = true
				break
			}
		}
		if done {
			break
		}
		page++
	}
	if err := store.Save(metrics.Metric{Name: "github.activities",
		Kind: metrics.Daily}.NewMeasurement(time.Now(), float64(activities))); err != nil {
		task.LogFatal(log, "msg", "failed to save measurement", "error", err)
	}
	log.Log("msg", "saved measurement", "activities", activities)
}

func isContribEvent(eventType string) bool {
	for _, v := range contributionEvents {
		if v == eventType {
			return true
		}
	}
	return false
}
