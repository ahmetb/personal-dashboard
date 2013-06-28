#!/usr/bin/env python
# coding: utf-8

import sys
import time
import json
import logging
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

    gauges_ds = simplegauges.datastores.azuretable.AzureGaugeDatastore(
        config['azure.account'], config['azure.key'], config['azure.table'])
    gauge_factory = simplegauges.daily_gauge_factory(gauges_ds)
    tasks.set_simplegauges_factory(gauge_factory)
    tasks.set_config(config)

    import fixture  # should be imported after setting configs for decorators

    if not fixture.tasks:
        logger.error('No tasks found in the fixture.py')
        sys.exit(1)

    for task in fixture.tasks:
        method = task[0]
        name = '{0}.{1}'.format(method.__module__, method.__name__)

        # if not task[1]:
        #     logger.critical('Task {0} scheduling interval is invalid'
        #                     .format(name))
        try:
            task[0]()
            logger.info('Successfully bootstrapped: {0}'.format(name))
        except Exception as e:
            logger.error('Error while bootstrapping: {0}'.format(name))
            logger.error(e)
            raise e

    # at this point all tasks ran once successfully
    sched = Scheduler()

    # schedule tasks
    for task in fixture.tasks:
        cron_kwargs = parse_cron_tuple(task[1])
        sched.add_cron_job(task[0], **cron_kwargs)

    sched.start()

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
