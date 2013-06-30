# coding: utf-8

import logging
import datetime
from dateutil.tz import tzutc


_simplegauges_factory = None
_tasks_config = {}


def set_simplegauges_factory(gauge_factory):
    global _simplegauges_factory
    _simplegauges_factory = gauge_factory


def set_config(config):
    global _tasks_config
    _tasks_config = config


def extract_keys(dictionary, keys):
    d = {}
    for k in keys:
        if k not in dictionary:
            raise Exception('task configuration "{0}" does not exist.'
                            .format(k))
        d[k] = dictionary[k]
    return d


class requires(object):
    def __init__(self, *config_key_argz):
        if config_key_argz:
            self.config_keys = list(config_key_argz)

    def __call__(self, func):
        def decorator():
            config = {}
            if self.config_keys:
                config = extract_keys(_tasks_config, self.config_keys)
            logger = logging.getLogger('{0}.{1}'.format(func.__module__,
                                                        func.__name__))
            return func(_simplegauges_factory, config, logger)
        decorator.__name__ = func.__name__
        return decorator


def today_utc():
    return datetime.datetime.now(tzutc()).date()


def epoch_for_day(day):
    day_time = datetime.datetime.combine(day, datetime.time(tzinfo=tzutc()))
    epoch = datetime.datetime(1970, 1, 1, tzinfo=tzutc())
    return int((day_time - epoch).total_seconds())
