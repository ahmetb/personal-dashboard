# personal-dashboard

This piece of software is collecting data about me all the time and aggregating them.

# Installation

### Requirements

You should get the following packages on your system:

* `python`
* `supervisor`
* `python-pip` (`easy_install pip`)
* `virtualenvwrapper` (to python packages folder tidy)

## Setting up virtual environment

`personal-dashboard` and `simplegauges` are not in PyPi yet. So you'll use GitHub to get the bits.

Go to directory where you want to install and:

    mkvirtualenv pd
    workon pd
    git clone https://github.com/ahmetalpbalkan/personal-dashboard.git
    cd personal-dashboard
    git clone https://github.com/ahmetalpbalkan/simplegauges.git
    pip install pytz python-dateutil APScheduler
    
## Setting up supervisor

To keep collecting data all the time we need to configure `supervisor`.

Find out your `python` binary in `~/.virtualenvs/pd/bin/python` (e.g. `/home/pi/.virtualenvs/pd/bin/python`)

Create `/etc/supervisor/conf.d/pd.conf` by adding:

```
[program:pd]
command=/home/pi/.virtualenvs/pd/bin/python -u /home/pi/personal-dashboard/taskhost.py
directory=/home/pi/personal-dashboard
autostart=unexpected
redirect_stderr=true
stdout_logfile=/var/log/pd.log
```

Restart the supervisor:

    service supervisor stop
    service supervisor start

Supervisor logs will show up in `/var/log/supervisor/supervisord.log` and `personal-dasboard` logs will start to show up in `/var/log/pd.log` in this case.

If you are seeing this in `supervisord.log` you are good:

    2013-07-19 00:33:30,292 INFO success: pd entered RUNNING state, process has stayed up for > than 1 seconds (startsecs)

## Python package dependencies for tasks

* `fb.py`: `facebook-sdk`
* `twitter.py`: `tweepy`
* `runkeeper.py` : `healthgraph-api`
* `kloutcom.py` : `klout`
* `reporting.py` : `azure`
* `foursq.py` : `foursquare`
* `atelog` : `feedparser`
* `tmp102` : `i2c-tools` & `python-smbus` (Ubuntu/Debian packages, use `apt-get`)
