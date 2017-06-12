package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"code.cloudfoundry.org/loggregator/metricemitter"

	"google.golang.org/grpc"

	"code.cloudfoundry.org/loggregator/plumbing"
	"code.cloudfoundry.org/loggregator/profiler"

	"code.cloudfoundry.org/loggregator/rlp/app"
)

func main() {
	egressPort := flag.Int("egress-port", 0, "The port of the Egress server")
	ingressAddrsList := flag.String("ingress-addrs", "", "The addresses of Dopplers")
	pprofPort := flag.Int("pprof-port", 6061, "The port of pprof for health checks")
	healthAddr := flag.String("health-addr", "localhost:22222", "The address for the health endpoint")

	caFile := flag.String("ca", "", "The file path for the CA cert")
	certFile := flag.String("cert", "", "The file path for the client cert")
	keyFile := flag.String("key", "", "The file path for the client key")

	metronAddr := flag.String("metron-addr", "localhost:3458", "The GRPC address to inject metrics to")
	metricEmitterInterval := flag.Duration("metric-emitter-interval", time.Minute, "The interval to send batched metrics to metron")

	flag.Parse()

	dopplerCredentials, err := plumbing.NewCredentials(
		*certFile,
		*keyFile,
		*caFile,
		"doppler",
	)
	if err != nil {
		log.Fatalf("Could not use TLS config: %s", err)
	}

	hostPorts := strings.Split(*ingressAddrsList, ",")
	if len(hostPorts) == 0 {
		log.Fatal("no Ingress Addrs were provided")
	}

	metronCredentials, err := plumbing.NewCredentials(
		*certFile,
		*keyFile,
		*caFile,
		"metron",
	)
	if err != nil {
		log.Fatalf("Could not use TLS config: %s", err)
	}

	// metric-documentation-v2: setup function
	metric, err := metricemitter.NewClient(
		*metronAddr,
		metricemitter.WithGRPCDialOptions(grpc.WithTransportCredentials(metronCredentials)),
		metricemitter.WithOrigin("loggregator.rlp"),
		metricemitter.WithPulseInterval(*metricEmitterInterval),
	)
	if err != nil {
		log.Fatalf("Couldn't connect to metric emitter: %s", err)
	}

	rlp := app.NewRLP(
		metric,
		app.WithEgressPort(*egressPort),
		app.WithIngressAddrs(hostPorts),
		app.WithIngressDialOptions(grpc.WithTransportCredentials(dopplerCredentials)),
		app.WithEgressServerOptions(grpc.Creds(dopplerCredentials)),
		app.WithHealthAddr(*healthAddr),
	)
	go rlp.Start()

	profiler.New(uint32(*pprofPort)).Start()
}