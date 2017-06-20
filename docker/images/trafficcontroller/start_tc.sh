#!/bin/bash

count="${DOPPLER_COUNT:-0}"
echo "connecting to $count dopplers"

dopplers="\"${DEPLOYMENT}_doppler_1:5678\""

for i in $(seq 2 $count); do
    dopplers="$dopplers,\"${DEPLOYMENT}_doppler_$i:5678\""
done

/trafficcontroller -disableAccessControl \
    -config <(cat /config.json | jq .DopplerAddrs=[$dopplers])
