# coding: utf-8
from . import requires, today_utc
import requests
from datetime import timedelta


DAYS_BACK = 4

@requires('jawboneup.access_token')
def sleeps(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.sleeps')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/users/@me/sleeps',
                     headers=headers)

    for i in range(DAYS_BACK):
        day = today_utc() - timedelta(days=i)
        resp = r.json()
        sleeps = resp['data']['items']
        day_date = day.date()
        today_fmt = int(day_date.strftime('%Y%m%d'))
        today_sleeps = filter(lambda s: s['date'] == today_fmt, sleeps)
        
        duration = sum([s['details']['duration'] - s['details']['awake']
                       for s in today_sleeps]) / 60.0 # in minutes
        if duration == 0:
            logger.info('Sleeps not found on {0}, not saving.'.format(day_date))
        else:
            gauge.save(day, duration)
            logger.info('Saved {0}min sleep on {1}'.format(duration, day_date))


@requires('jawboneup.access_token')
def steps(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.steps')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/users/@me/moves',
                     headers=headers)

    for i in range(DAYS_BACK):
        day = today_utc() - timedelta(days=i)
        resp = r.json()
        moves = resp['data']['items']
        day_date = day.date()
        today_fmt = int(day_date.strftime('%Y%m%d'))
        today_moves = filter(lambda m: m['date'] == today_fmt, moves)
        steps = sum([m['details']['steps'] for m in today_moves])

        if steps == 0:
            logger.info('Steps not found on {0}, not saving.'.format(day_date))
        else:
            gauge.save(day, steps)
            logger.info('Saved {0} steps on {1}'.format(steps, day_date))

@requires('jawboneup.access_token')
def caffeine(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.caffeine')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/users/@me/meals',
                     headers=headers)

    for i in range(DAYS_BACK):
        day = today_utc() - timedelta(days=i)
        resp = r.json()
        meals = resp['data']['items']
        day_date = day.date()
        today_fmt = int(day_date.strftime('%Y%m%d'))
        today_meals = filter(lambda m: m['date'] == today_fmt, meals)
        caffeine = sum([m['details']['caffeine'] for m in today_meals])

        if caffeine == 0:
            logger.info('Caffeine not found on {0}, not saving.'.format(day_date))
        else:
            gauge.save(day, caffeine)
            logger.info('Saved {0}mg. caffeine  on {1}'.format(caffeine, day_date))

@requires('jawboneup.access_token')
def heart_rate(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.resting_heartrate')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/v.1.1/users/@me/heartrates',
                     headers=headers)
    for i in range(DAYS_BACK):
        day = today_utc() - timedelta(days=i)
        resp = r.json()
        items = resp['data']['items']
        day_date = day.date()
        today_fmt = int(day_date.strftime('%Y%m%d'))
        today = filter(lambda hr: hr['date'] == today_fmt, items)

        if not today or len(today) == 0 or not today[0]['resting_heartrate']:
             logger.info('Heart rate not found on {0}, not saving.'.format(day_date))
        else:
            hr = today[0]['resting_heartrate']
            gauge.save(day, hr)
            logger.info('Saved heart rate {0} on {1}'.format(hr, day_date))
