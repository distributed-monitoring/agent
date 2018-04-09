.. image:: https://travis-ci.org/distributed-monitoring/agent.svg?branch=master
    :target: https://travis-ci.org/distributed-monitoring/agent
.. image:: https://goreportcard.com/badge/github.com/distributed-monitoring/agent
    :target: https://goreportcard.com/report/github.com/distributed-monitoring/agent

==========
LocalAgent
==========

About
=======

Agent proto type for DMA.

This repositiory contains a local agent
to dynamically change collectd's work
such as function enable setting, collection interval,
and notification policies.

Getting Started
=================

Installation
--------------
Operate by the user who starts collectd. (``root`` for CentOS ) ::

    # go get github.com/distributed-monitoring/agent/cmd/server

AMQP setting is hard-coded with apex's openstack RabbitMQ setting.
(This is WIP. We will make it configurable later.)

* username: guest
* password: (given by environment variable ``AMQP_PASSWORD``)
* IP: 192.0.2.11
* port: 5672

If you want to change these, set the URI of amqp.::

    # cd $GOPATH/src/github.com/distributed-monitoring/agent/cmd/server
    # vi amqp.go
    (Edit variables 'amqpPass' and 'amqpURL')
    # go install
    
Execute server.::

    # server -type <Server Type>

Configure
-----------

(T.B.D.)

Verification
--------------

API server (``server -type api``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
::

    # echo "## test config ##" > /tmp/mytest.conf
    # curl -v http://localhost:12345/collectd/conf -F "file=@/tmp/mytest.conf"
    # systemctl status collectd
    (check the status of collectd is active)
    # cat /etc/collectd/collectd.conf.d/mytest.conf
    ## test config ##

PusSub Subscriber server (``server -type pubsub``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
::

    # echo "## test config ##" > /tmp/mytest.conf
    # cd $GOPATH/src/github.com/distributed-monitoring/agent/test/publish
    # vi emit_conf.py
    (Edit variables 'credentials' and 'connection')
    # ./emit_conf.py /tmp/mytest.conf
    # systemctl status collectd
    (check the status of collectd is active)
    # cat /etc/collectd/collectd.conf.d/mytest.conf
    ## test config ##



Features
==========

* Dynamic setting of collectd
* ...




