# coding: utf-8

from . import requires
from pytz import timezone
from datetime import datetime
import smbus


"""Reads the instant temperature from TMP102 temperature sensor.
Requires packages:

    apt-get install i2c-tools
    apt-get install python-smbus

Configuration parameters:

    tmp102.tz: IANA timezone name to save the time of record. Use one of the
        strings from TZ column here for your local time:
        https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

    tmp102.bus: I2C bus number. if command "i2cdetect -y 1" shows anything
        use 1, or try 2 and so on
"""


@requires('tmp102.tz', 'tmp102.bus')
def temperature(gauge_factory, config, logger):
    """Saves current temperature in local time to hourly gauge.
    """

    gauge = gauge_factory('tmp102.temperature', gauge_type='hourly')
    tz = timezone(config['tmp102.tz'])

    # Code snippet taken from
    # http://bradsmc.blogspot.com/2013/04/reading-temperature-from-tmp02.html
    bus = smbus.SMBus(config['tmp102.bus'])
    data = bus.read_i2c_block_data(0x48, 0)
    msb = data[0]
    lsb = data[1]
    temp = (((msb << 8) | lsb) >> 4) * 0.0625
    now_local = datetime.now(tz)

    gauge.save(now_local, temp)
    logger.info('Saved temperature {0} C'.format(temp))
