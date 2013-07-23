# coding: utf-8

from . import requires, today_utc, epoch_for_day
from foursquare import Foursquare


@requires('foursquare.access_token')
def checkins(gauge_factory, config, logger):
    """Number of foursquare check-ins done since midnight today in UTC.
    """

    gauge = gauge_factory('foursquare.checkins')
    client = Foursquare(access_token=config['foursquare.access_token'])

    epoch = epoch_for_day(today_utc())
    checkins = client.users.checkins(params={'afterTimestamp': epoch})
    checkins = checkins['checkins']['items']

    gauge.save(today_utc(), len(checkins))
    logger.info('Saved {0} foursquare checkins'.format(len(checkins)))
