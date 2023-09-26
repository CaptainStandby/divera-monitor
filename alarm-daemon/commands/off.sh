#!/bin/bash -ue

echo "standby 0.0.0.0" | cec-client -s -d 1 > /dev/null
