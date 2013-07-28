# coding: utf-8

from . import requires, today_utc, epoch_for_day
import requests


@requires('lastfm.api_key', 'lastfm.user')
def tracks_listened(gauge_factory, config, logger):
    """Number tracks listened today
    """

    gauge = gauge_factory('lastfm.listened')
    epoch_today = epoch_for_day(today_utc())

    params = {
        'method': 'user.getrecenttracks',
        'user': config['lastfm.user'],
        'api_key': config['lastfm.api_key'],
        'from': epoch_today,
        'format': 'json'
    }

    r = requests.get('http://ws.audioscrobbler.com/2.0', params=params)
    resp = r.json()['recenttracks']

    listened = int(resp['@attr']['total']) if '@attr' in resp else 0

    gauge.save(today_utc(), listened)
    logger.info('Saved {0} last.fm tracks for {1}'.format(listened, today_utc()))
