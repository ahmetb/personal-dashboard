# coding: utf-8

from . import requires, now_utc, today_utc
from simplegauges import interpolators, postprocessors, aggregators
import datetime
from datetime import timedelta
import json
from azure.storage import BlobService


JSONP_CALLBACK_NAME = 'renderData'

zero_fill_daily = lambda data: postprocessors.day_fill(data, 0)
zero_fill_weekly = lambda data: postprocessors.week_fill(data, 0)
monthly_max = lambda data: aggregators.monthly(data, max)
weekly_max = lambda data: aggregators.weekly(data, max)
weekly_min = lambda data: aggregators.weekly(data, min)
weekly_sum = lambda data: aggregators.weekly(data, sum)


@requires('azure.account', 'azure.key', 'azure.blob.container',
          'azure.blob.name')
def generate_and_upload(gauge_factory, config, logger):
    start = datetime.datetime.now()
    twitter_followers = gauge_factory('twitter.followers')
    twitter_tweets = gauge_factory('twitter.tweets')
    fb_friends = gauge_factory('facebook.friends')
    foursq_checkins = gauge_factory('foursquare.checkins')
    klout_score = gauge_factory('klout.score')
    runkeeper_activities = gauge_factory('runkeeper.activities')
    runkeeper_calories = gauge_factory('runkeeper.calories_burned')
    runkeeper_weight = gauge_factory('runkeeper.weight')
    tmp102_celsius = gauge_factory('tmp102.temperature', gauge_type='hourly')
    lastfm_listened = gauge_factory('lastfm.listened')

    data = {}
    data_sources = [
        # (output key, gauge, days back, aggregator, postprocessors)
        ('twitter.followers', twitter_followers, 30, None,
            [zero_fill_daily, interpolators.linear]),
        ('twitter.tweets', twitter_tweets, 20, None, [zero_fill_daily]),
        ('facebook.friends', fb_friends, 180, monthly_max, None),
        ('foursquare.checkins', foursq_checkins, 14, None, [zero_fill_daily]),
        ('lastfm.listened', lastfm_listened, 14, None, [zero_fill_daily]),
        ('klout.score', klout_score, 30, weekly_max, [zero_fill_weekly,
                                                      interpolators.linear]),
        ('runkeeper.calories', runkeeper_calories, 60, weekly_sum,
            [zero_fill_weekly]),
        ('runkeeper.activities', runkeeper_activities, 60, weekly_sum,
            [zero_fill_weekly]),
        ('runkeeper.weight', runkeeper_weight, 180, weekly_min,
            [zero_fill_weekly, interpolators.linear]),
        ('tmp102.temperature', tmp102_celsius, 2.5, None, None)
    ]

    for ds in data_sources:
        data[ds[0]] = ds[1].aggregate(today_utc() - timedelta(days=ds[2]),
                                      aggregator=ds[3],
                                      post_processors=ds[4])

    report = {
        'generated': str(now_utc()),
        'data': data,
        'took': (datetime.datetime.now() - start).seconds
    }
    report_json = json.dumps(report, indent=4, default=json_date_serializer)
    report_content = '{0}({1})'.format(JSONP_CALLBACK_NAME, report_json)
    
    blob_service = BlobService(config['azure.account'], config['azure.key'])
    blob_service.create_container(config['azure.blob.container'])
    blob_service.set_container_acl(config['azure.blob.container'],
                                   x_ms_blob_public_access='container')
    blob_service.put_blob(config['azure.blob.container'],
                          config['azure.blob.name'], report_content, 'BlockBlob')

    took = (datetime.datetime.now() - start).seconds
    logger.info('Report generated and uploaded. Took {0} s.'.format(took))


def json_date_serializer(obj):
    if isinstance(obj, datetime.datetime) or isinstance(obj, datetime.date):
        return str(obj)
    return obj
