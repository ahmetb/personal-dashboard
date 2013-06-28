# coding: utf-8

from . import requires
import datetime
from dateutil.tz import tzutc
import healthgraph


@requires('runkeeper.access_token')
def activities_and_calories(gauge_factory, config, logger):
    activity_gauge = gauge_factory('runkeeper.activities')
    calorie_gauge = gauge_factory('runkeeper.calories_burned')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))

    activities = list(user.get_fitness_activity_iter())
    #TODO code above loads all fitness activities, inefficient

    #TODO runkeeper returns start_time in local time, convert to UTC
    today_utc = datetime.datetime.now(tzutc()).date()
    today_data = filter(lambda a: a['start_time'].date() == today_utc,
                        activities)
    total_activities = len(today_data)
    total_calories = int(sum([a['total_calories'] for a in today_data]))

    activity_gauge.save(today_utc, total_activities)
    calorie_gauge.save(today_utc, total_calories)
    logger.info('Saved {0} activities ({1} cal) for {2}'
                .format(total_activities, total_calories, today_utc))


@requires('runkeeper.access_token')
def sleeps(gauge_factory, config, logger):
    gauge = gauge_factory('runkeeper.sleeps')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))

    sleeps = list(user.get_sleep_measurement_iter())
    #TODO code above loads all sleep measurements, inefficient

    today_utc = datetime.datetime.now(tzutc()).date()
    today_sleeps = filter(lambda s: s['timestamp'].date() == today_utc, sleeps)
    total_sleep_mins = sum([a['total_sleep'] for a in today_sleeps])

    gauge.save(today_utc, total_sleep_mins)
    logger.info('Saved {0} min. sleep for {1}'.format(total_sleep_mins,
                                                      today_utc))


@requires('runkeeper.access_token')
def weight(gauge_factory, config, logger):
    """Saves last known weight (if any) for today
    """
    
    gauge = gauge_factory('runkeeper.weight')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))

    weights = list(user.get_weight_measurement_iter())
    #TODO code above loads all weight measurements, inefficient

    # since items are loaded in descending order, first result is latest weight
    if weights:
        today_utc = datetime.datetime.now(tzutc()).date()
        last_known_weight = weights[0]['weight']
        gauge.save(today_utc, last_known_weight)
        logger.info('Saved {0} kg weight for {1}'.format(last_known_weight,
                                                         today_utc))
    else:
        logger.warning('Runkeeper account has no weight measurement data.')
