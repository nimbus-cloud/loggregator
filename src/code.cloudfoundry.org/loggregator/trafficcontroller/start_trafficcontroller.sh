#!/bin/bash

count=$((${1:-20} - 1))

dopplers="\"localhost:23000\""

for i in $(seq 1 $count); do
    port=$((23000 + $i))
    dopplers="$dopplers,\"localhost:$port\""
done

trafficcontroller -disableAccessControl \
    -config <(cat config/loggregator_trafficcontroller.json | jq .DopplerAddrs=[$dopplers])
