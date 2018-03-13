#!/bin/bash

while :
do
  sudo ping 192.0.2.1 -i 0.00005 -c 1000 -s 1000 -q
  sleep 21
done

