# coding: utf-8

from . import requires, today_utc
import facebook


@requires('facebook.access_token')
def friends_count(gauge_factory, config, logger):
    #TODO does not refresh the existing long-living access token
    #TODO does not use paging (I have <5000 friends)
    gauge = gauge_factory('facebook.friends')

    graph = facebook.GraphAPI(config['facebook.access_token'])
    resp = graph.fql("SELECT friend_count FROM user WHERE uid = me()")
    friends = resp[0]['friend_count']

    gauge.save(today_utc(), friends)
    logger.info('Saved Facebook friend count: {0}'.format(friends))
