#!/bin/bash



count=$((${1:-20} - 1))
for i in $(seq 0 $count); do
    health_port=$((22000 + $i))
    grpc_port=$((23000 + $i))
    ws_port=$((24000 + $i))
    udp_port=$((25000 + $i))
    pprof_port=$((26000 + $i))

    doppler --config <(cat config/doppler.json |
        jq .HealthAddr=\"localhost:$health_port\" |
        jq .GRPC.Port=$grpc_port |
        jq .OutgoingPort=$ws_port |
        jq .IncomingUDPPort=$udp_port |
        jq .PPROFPort=$pprof_port) &
done

trap "killall doppler" EXIT
wait
