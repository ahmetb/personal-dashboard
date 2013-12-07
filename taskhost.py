#!/usr/bin/env python
# coding: utf-8

import sys
import time
import json
import logging
import datetime
from apscheduler.scheduler import Scheduler
import simplegauges
import tasks


_tasks_config_file = 'tasks.config'


def main():
    configure_logging()
    logger = logging.getLogger('taskhost')
    config = {}

    try:
        with open(_tasks_config_file) as f:
            config = json.loads(f.read())
        logger.debug('Successfully read configuration file.')
    except Exception as e:
        logger.critical('Cannot read configuration file: {0}'
                        .format(_tasks_config_file))
        logger.critical(e)
        sys.exit(1)

    from simplegauges.datastores.azuretable import AzureGaugeDatastore
    gauges_ds = AzureGaugeDatastore(config['azure.account'],
                                    config['azure.key'], config['azure.table'])
    gauge_factory = simplegauges.gauge_factory(gauges_ds)
    tasks.set_simplegauges_factory(gauge_factory)
    tasks.set_config(config)

    import fixture  # should be imported after setting configs for decorators

    if not fixture.tasks:
        logger.error('No tasks found in the fixture.py')
        sys.exit(1)

    errors = False
    for task in fixture.tasks:
        method = task[0]
        name = '{0}.{1}'.format(method.__module__, method.__name__)
        try:
            task[0]()
            logger.info('Successfully bootstrapped: {0}'.format(name))
        except Exception as e:
            errors = True
            logger.error('Error while bootstrapping: {0}'.format(name))
            logger.error(e)

    if errors:
        logger.info('Starting scheduler in 10 seconds...')
        time.sleep(10)
    else:
        logger.info('Starting scheduler...')

    # at this point all tasks ran once successfully
    sched = Scheduler()

    # schedule tasks
    for task in fixture.tasks:
        cron_kwargs = parse_cron_tuple(task[1])
        sched.add_cron_job(task[0], **cron_kwargs)

    sched.start()
    logger.info('Scheduler started with {0} jobs.'
                .format(len(sched.get_jobs())))
    now = datetime.datetime.now()
    for j in sched.get_jobs():
        logger.debug('Scheduled: {0}.{1}, next run:{2}'
                     .format(j.func.__module__, j.func.__name__,
                             j.compute_next_run_time(now)))

    # deamonize the process
    while True:
        time.sleep(10)


def parse_cron_tuple(cron_tuple):
    """Parses (hour,minute,second) or (hour,minute) or (hour) cron
    scheduling defintions into kwargs dictionary
    """
    if type(cron_tuple) is not tuple:
        raise Exception('Given cron format is not tuple: {0}'
                        .format(cron_tuple))
    kwargs = {}
    l = len(cron_tuple)

    if l > 0:
        kwargs['hour'] = cron_tuple[0]
    if l > 1:
        kwargs['minute'] = cron_tuple[1]
    if l > 2:
        kwargs['second'] = cron_tuple[2]
    return kwargs


def configure_logging():
    logfmt = '[%(asctime)s] %(levelname)s [%(name)s] %(message)s'

    # configure to StreamHandler with log format
    logging.basicConfig(level=logging.DEBUG, format=logfmt)

    # reduce noise from 3rd party packages
    logging.getLogger('requests.packages.urllib3.connectionpool')\
        .setLevel(logging.CRITICAL)
    logging.getLogger('apscheduler').setLevel(logging.WARNING)


if __name__ == '__main__':
    main()
