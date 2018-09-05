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
.. code:: bash

    # docker run -tid -p 6379:6379 --name barometer-redis redis

    # cd <Directory that Dockerfile is located>
    # docker build -t opnfv/barometer-localagent --build-arg http_proxy=`echo $http_proxy` \
      --build-arg https_proxy=`echo $https_proxy` -f Dockerfile .
    # docker images

    # cd <Directory that examples of config.toml is located>
    # mkdir /etc/barometer-localagent
    # cp examples/config.toml /etc/barometer-localagent/
    # vi /etc/barometer-localagent/config.toml
    (edit amqp_password and os_password:OpenStack admin password)

    (When there is no key for SSH access authentication)
    # ssh-keygen
    (Press Enter until done)

    (Backup if necessary)
    # cp ~/.ssh/authorized_keys ~/.ssh/authorized_keys_org

    # cat ~/.ssh/authorized_keys_org ~/.ssh/id_rsa.pub \
      > ~/.ssh/authorized_keys

    # docker run -tid --net=host --name server \
      -v /etc/barometer-localagent:/etc/barometer-localagent \
      -v /root/.ssh/id_rsa:/root/.ssh/id_rsa \
      -v /etc/collectd/collectd.conf.d:/etc/collectd/collectd.conf.d \
      opnfv/barometer-localagent /server

    # docker run -tid --net=host --name infofetch \
      -v /etc/barometer-localagent:/etc/barometer-localagent \
      -v /var/run/libvirt:/var/run/libvirt \
      opnfv/barometer-localagent /infofetch

    # docker cp infofetch:/threshold ./
    # ln -s ${PWD}/threshold /usr/local/bin/


Features
==========

* Dynamic setting of collectd
* Annotate...




