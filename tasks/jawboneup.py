# coding: utf-8
from . import requires, today_utc
import requests

@requires('jawboneup.access_token')
def sleeps(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.sleeps')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/users/@me/sleeps',
                     headers=headers)

    day = today_utc()
    resp = r.json()
    sleeps = resp['data']['items']
    today = day.date()
    today_fmt = int(today.strftime('%Y%m%d'))
    today_sleeps = filter(lambda s: s['date'] == today_fmt, sleeps)
    
    duration = sum([s['details']['duration'] - s['details']['awake']
                   for s in today_sleeps]) / 60.0 # in minutes
    if duration == 0:
        logger.info('Sleeps not found today, not saving.')
    else:
        gauge.save(day, duration)
        logger.info('Saved {0} min sleep for {1}'.format(duration, day))


@requires('jawboneup.access_token')
def steps(gauge_factory, config, logger):
    gauge = gauge_factory('jawbone.steps')
    access_token = config['jawboneup.access_token']

    headers = {'Authorization' : 'Bearer {0}'.format(access_token)}
    r = requests.get('https://jawbone.com/nudge/api/users/@me/moves',
                     headers=headers)
    day = today_utc()
    resp = r.json()
    moves = resp['data']['items']
    today = day.date()
    today_fmt = int(today.strftime('%Y%m%d'))
    today_moves = filter(lambda m: m['date'] == today_fmt, moves)
    steps = sum([m['details']['steps'] for m in today_moves])

    if steps == 0:
        logger.info('Steps not found for today yet, not saving.')
    else:
        gauge.save(day, steps)
        logger.info('Saved {0} steps for {1}'.format(steps, day))
