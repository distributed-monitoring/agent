#!/bin/bash

CONFFILE=$1

curl -v 'http://192.0.2.15:12345/collectd/conf' -F "file=@${CONFFILE}"

