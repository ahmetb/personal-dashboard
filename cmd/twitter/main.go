package main

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/metrics"
	"github.com/ahmetalpbalkan/personal-dashboard/pkg/task"
	"github.com/jinzhu/now"
)

var GitSummary string // provided by govvv

type config struct {
	Tasks struct {
		Twitter struct {
			ConsumerKey       string `toml:"consumer_key"`
			ConsumerSecret    string `toml:"consumer_secret"`
			AccessToken       string `toml:"access_token"`
			AccessTokenSecret string `toml:"access_token_secret"`
			ExcludeMentions   bool   `toml:"exclude_mentions"`
			ExcludeRetweets   bool   `toml:"exclude_retweets"`
		} `toml:"twitter"`
	} `toml:"tasks"`
}

var (
	log = task.LoggerWithTask("twitter", GitSummary)
)

func main() {
	store, err := task.GetDatastore()
	if err != nil {
		task.LogFatal(log, "error", err)
	}

	if len(os.Args) < 2 {
		task.LogFatal(log, "error", "not enough arguments. specify task")
	} else if len(os.Args) > 2 {
		task.LogFatal(log, "error", "too many arguments")
	}

	var c config
	if err := task.ReadConfig(&c); err != nil {
		task.LogFatal(log, "error", err)
	}

	cmd := os.Args[1]
	log = log.With("cmd", cmd)
	switch cmd {
	case "followers":
		followers(c, store)
	case "tweets":
		tweets(c, store)
	default:
		task.LogFatal(log, "error", "unknown command", "cmd", cmd)
	}
}

func followers(c config, store metrics.Datastore) {
	requireCreds(c)
	u, err := mkAPI(c).GetSelf(nil)
	if err != nil {
		task.LogFatal(log, "msg", "failed to get self", "error", err)
	}

	log.Log("msg", "retrieved user", "screen_name", u.ScreenName, "followers", u.FollowersCount, "following", u.FriendsCount)
	if err := store.Save(metrics.Metric{
		Kind: metrics.Daily,
		Name: "twitter.followers",
	}.NewMeasurement(time.Now(), float64(u.FollowersCount))); err != nil {
		task.LogFatal(log, "error", err)
	}
	log.Log("msg", "saved followers")
}

func tweets(c config, store metrics.Datastore) {
	requireCreds(c)
	today := now.New(time.Now().UTC()).BeginningOfDay()
	timeline, err := mkAPI(c).GetUserTimeline(url.Values{"count": []string{"200"}})
	if err != nil {
		task.LogFatal(log, "msg", "failed to get the user's timeline", "error", err)
	}
	count := 0
	for _, t := range timeline {
		date, err := t.CreatedAtTime()
		if err != nil {
			task.LogFatal(log, "msg", "cannot parsed date", "id", t.Id, "date", t.CreatedAt, "error", err)
		}

		tweetDay := now.New(date.UTC()).BeginningOfDay()
		log.Log("id", t.Id, "tweet_date", tweetDay, "today_date", today)
		if c.Tasks.Twitter.ExcludeMentions && strings.HasPrefix(t.Text, "@") {
			continue
		} else if c.Tasks.Twitter.ExcludeRetweets && t.Retweeted {
			continue
		} else if today.Equal(tweetDay) {
			count++
		} else {
			break
		}
	}
	log.Log("tweets_count", count, "day", today)
	if err := store.Save(metrics.Metric{
		Kind: metrics.Daily,
		Name: "twitter.tweets",
	}.NewMeasurement(time.Now(), float64(count))); err != nil {
		task.LogFatal(log, "error", err)
	}
	log.Log("msg", "saved tweets")
}

func mkAPI(c config) *anaconda.TwitterApi {
	anaconda.SetConsumerKey(c.Tasks.Twitter.ConsumerKey)
	anaconda.SetConsumerSecret(c.Tasks.Twitter.ConsumerSecret)
	return anaconda.NewTwitterApi(c.Tasks.Twitter.AccessToken, c.Tasks.Twitter.AccessTokenSecret)
}

func requireCreds(c config) {
	task.RequireConfig(log, c.Tasks.Twitter.ConsumerKey, "consumer_key")
	task.RequireConfig(log, c.Tasks.Twitter.ConsumerSecret, "consumer_secret")
	task.RequireConfig(log, c.Tasks.Twitter.AccessToken, "access_token")
	task.RequireConfig(log, c.Tasks.Twitter.AccessTokenSecret, "access_token_secret")
}
