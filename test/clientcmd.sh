#!/bin/bash

CONFFILE=$1

curl -v 'http://overcloud-novacompute-0.internalapi:12345/collectd/conf' -F "file=@${CONFFILE}"

