# coding: utf-8

import logging
import datetime
import pytz


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
            raise Exception('task configuration key "{0}" does not exist.'
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


def now_utc():
    """returns UTC time with tzinfo since datetime.utcnow does not have
    tzinfo
    """

    return datetime.datetime.now(pytz.utc)


def today_utc():
    """returns datetime.datetime only containing year/month/day of UTC now
    where hh:mm:ss.ms cleared off
    """

    dt = now_utc()
    dt = datetime.datetime.combine(dt.date(), datetime.time(tzinfo=pytz.utc))
    return dt


def epoch_for_datetime(dt):
    """converts a given datetime.datetime to UNIX epoch time in seconds
    """

    epoch = datetime.datetime(1970, 1, 1, tzinfo=pytz.utc)
    return int((dt - epoch).total_seconds())


def epoch_for_day(day):
    """converts a given datetime.date to UNIX epoch time in seconds
    """

    day_dt = datetime.datetime.combine(day, datetime.time(tzinfo=pytz.utc))
    return epoch_for_datetime(day_dt)
