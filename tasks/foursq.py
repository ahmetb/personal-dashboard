# coding: utf-8

from . import requires
import datetime
from dateutil.tz import tzutc
from foursquare import Foursquare


@requires('foursquare.access_token')
def checkins(gauge_factory, config, logger):
    """Check-ins done since current day (UTC) 00:00:00.
    """
    gauge = gauge_factory('foursquare.checkins')
    client = Foursquare(access_token=config['foursquare.access_token'])

    today_utc = datetime.datetime.now(tzutc()).date()
    epoch = epoch_for_day(today_utc)

    checkins = client.users.checkins(params={'afterTimestamp': epoch})
    checkins = checkins['checkins']['items']

    gauge.save(today_utc, len(checkins))
    logger.info('Saved {0} foursquare checkins'.format(len(checkins)))


def epoch_for_day(day):
    day_time = datetime.datetime.combine(day, datetime.time(tzinfo=tzutc()))
    epoch = datetime.datetime(1970, 1, 1, tzinfo=tzutc())
    return int((day_time - epoch).total_seconds())
