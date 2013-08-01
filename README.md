# personal-dashboard

This piece of software is collecting data about me all the time and aggregating them. [**Read about this project more &rarr;**](http://ahmetalpbalkan.com/blog/personal-dashboard/)

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
    

## Setting up tasks

You can find each task under `tasks/` directory. If you open up e.g. `twitter.py` you will see the folllowing:

    @requires('twitter.consumer_key', 'twitter.consumer_secret',
              'twitter.access_token', 'twitter.access_secret',
              'twitter.exclude_mentions')
              
This means you need to add these keys to `tasks.config` file. In the `personal-dashboard` directory, you will find [`tasks.config.sample`](tasks.config.sample) file, which is a JSON file. You can rename this file to `tasks.config` and add new keys as needed.

After that you have to fetch the dependencies for `twitter` task (see the section below) and set the fixture on this task.

If you go to [`fixture.py`](fixture.py), a sample configuration provided along with examples and how to set up the scheduling for the tasks you need.

## Python package dependencies for tasks

* `fb.py`: `facebook-sdk`
* `twitter.py`: `tweepy`
* `runkeeper.py` : `healthgraph-api`
* `kloutcom.py` : `klout`
* `reporting.py` : `azure`
* `foursq.py` : `foursquare`
* `lastfm` : `requests`
* `tmp102` : `i2c-tools` & `python-smbus` (Ubuntu/Debian packages, use `apt-get`)
 
## Starting data collector manually

Data collector is called `taskhost.py`. After running `workon pd`, you can run `python taskhost.py` and it will:

1. read the configuration from `tasks.config` file
2. read the task schedules from `fixture.py`
3. run each task successfully once
4. schedule tasks at the specified periods in the fixture
5. keep running

This process has to run all the time to collect data continuously and it may crash sporadically. This is why we need `supervisor`, a process monitoring system that could restart process when it unexpectedly quits or machine reboots.

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


## Reporting

I have a task called `reporting.py`. This system is not quite flexible and I designed it only for myself. It aggregates data from various gauges and uploads to Azure Blob Storage as a JSON file with a JSONP callback.

You may take this as a reference implementation and write your own.