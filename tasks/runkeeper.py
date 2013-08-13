# coding: utf-8

from . import requires, today_utc
import healthgraph


@requires('runkeeper.access_token')
def activities_and_calories(gauge_factory, config, logger):
    #TODO runkeeper returns activity start_time in local time, convert to UTC
    activity_gauge = gauge_factory('runkeeper.activities')
    calorie_gauge = gauge_factory('runkeeper.calories_burned')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))
    activities_iter = user.get_fitness_activity_iter()

    today = today_utc().date()
    today_activities = []
    for a in activities_iter:  # breaking early prevents loading all results
        day = a['start_time'].date()
        if day == today:
            today_activities.append(a)
        elif (today - day).days > 2:
            break

    total_activities = len(today_activities)
    total_calories = int(sum([a['total_calories'] for a in today_activities]))

    activity_gauge.save(today_utc(), total_activities)
    calorie_gauge.save(today_utc(), total_calories)
    logger.info('Saved {0} activities ({1} cal) for {2}'
                .format(total_activities, total_calories, today_utc()))


@requires('runkeeper.access_token')
def sleeps(gauge_factory, config, logger):
    gauge = gauge_factory('runkeeper.sleeps')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))
    sleeps_iter = user.get_sleep_measurement_iter()
    today = today_utc().date()
    today_sleeps = []
    for s in sleeps_iter:  # breaking early prevents loading all results
        day = s['timestamp'].date()
        if day == today:
            today_sleeps.append(s)
        elif (today - day).days > 2:
            break

    total_sleep_mins = sum([s['total_sleep'] for s in today_sleeps])

    gauge.save(today_utc(), total_sleep_mins)
    logger.info('Saved {0} min. sleep for {1}'.format(total_sleep_mins,
                                                      today_utc()))


@requires('runkeeper.access_token')
def weight(gauge_factory, config, logger):
    """Saves last known weight (if any) for today
    """

    gauge = gauge_factory('runkeeper.weight')

    user = healthgraph.User(session=healthgraph.
                            Session(config['runkeeper.access_token']))

    weight = None
    weights_iter = user.get_weight_measurement_iter()

    # since items are loaded in descending order, first result is latest weight
    for w in weights_iter:
        weight = w['weight']
        break  # no need to load results further

    if weight:
        gauge.save(today_utc(), weight)
        logger.info('Saved {0} kg weight for {1}'.format(weight, today_utc()))
    else:
        logger.warning('Runkeeper has no weight measurement data.')
