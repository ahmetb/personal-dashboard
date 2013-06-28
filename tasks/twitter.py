# coding: utf-8

from . import requires
import datetime
from dateutil.tz import tzutc
import tweepy


def twitter_api_handle(config):
    auth = tweepy.OAuthHandler(config['twitter.consumer_key'],
                               config['twitter.consumer_secret'])
    auth.set_access_token(config['twitter.access_token'],
                          config['twitter.access_secret'])
    return tweepy.API(auth)


def today_utc():
    return datetime.datetime.now(tzutc()).date()


@requires('twitter.consumer_key', 'twitter.consumer_secret',
          'twitter.access_token', 'twitter.access_secret')
def followers_count(gauge_factory, config, logger):
    gauge = gauge_factory('twitter.followers')
    api = twitter_api_handle(config)

    count = api.me().followers_count
    gauge.save(today_utc(), count)
    logger.info('Saved followers count: {0}'.format(count))


@requires('twitter.consumer_key', 'twitter.consumer_secret',
          'twitter.access_token', 'twitter.access_secret')
def tweets_count(gauge_factory, config, logger):
    #TODO if you have tweeted 200+ tweets in a day, records as 200
    gauge = gauge_factory('twitter.tweets')
    api = twitter_api_handle(config)

    timeline = api.user_timeline(count=200)
    c = sum(1 for st in timeline if st.created_at.date() == today_utc())
    logger.info('Saved tweets count: {0} for {1}'.format(c, today_utc()))
    gauge.save(today_utc(), c)
