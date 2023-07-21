#!/bin/bash -ue

TV_ADDRESS="0.0.0.0"

# switch tv on
echo "sending on command to tv"
echo "on $TV_ADDRESS" | cec-client -s -d 1 > /dev/null

# now poll tv for on state
echo "waiting for tv to turn on"
while true; do
        # get tv status
        STATUS=$(echo "pow $TV_ADDRESS" | cec-client -s -d 1 | grep "power status:" | awk '{print $3}')

        # if tv is on, break
        if [ "$STATUS" = "on" ]; then
                echo "TV is on"
                break
        fi

        # wait 1 second
        sleep 1
done

# switch tv to correct input
echo "switching tv to correct input"
echo "as" | cec-client -s -d 1 > /dev/null
