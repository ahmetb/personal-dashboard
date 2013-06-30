# coding: utf-8

from . import requires, today_utc
import klout


@requires('klout.api_key', 'klout.screen_name')
def score(gauge_factory, config, logger):
    gauge = gauge_factory('klout.score')
    k = klout.Klout(config['klout.api_key'])

    user = config['klout.screen_name']
    kloutId = k.identity.klout(screenName=user).get('id')
    if not kloutId:
        raise Exception("Klout id not found for screen name {0}".format(user))
    score = k.user.score(kloutId=kloutId).get('score')
    gauge.save(today_utc(), score)

    logger.info('Saved Klout score: {0}'.format(score))
