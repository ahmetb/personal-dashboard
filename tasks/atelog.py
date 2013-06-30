# coding: utf-8

from . import requires, today_utc
import feedparser
from dateutil import parser


"""Retrieves nutrition data from a tumblelog (Tumblr blog) RSS using
tags.
"""

COFFEE_TAG = 'coffee'


@requires('atelog.rss')
def coffees(gauge_factory, config, logger):
    gauge = gauge_factory('atelog.coffees')

    feed = feedparser.parse(config['atelog.rss'])
    entries = feed['entries']
    today = today_utc()

    coffees = len(filter(lambda x: parser.parse(x['published']).date() == today
                  and filter(lambda t: t.term == COFFEE_TAG, x['tags']),
                  entries))

    gauge.save(today, coffees)
    logger.info('Saved {0} coffee records for {1}'.format(coffees, today))
